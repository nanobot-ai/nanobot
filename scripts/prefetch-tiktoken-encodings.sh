#!/bin/sh
set -eu

cache_dir=${TIKTOKEN_CACHE_DIR:-}
if [ -z "${cache_dir}" ]; then
  echo "TIKTOKEN_CACHE_DIR must be set" >&2
  exit 1
fi

mkdir -p "${cache_dir}"
chmod 755 "${cache_dir}"

tmp_dir=$(mktemp -d)
cleanup() {
  rm -rf "${tmp_dir}"
}
trap cleanup EXIT INT TERM

download() {
  url=$1
  dest=$2

  if command -v curl >/dev/null 2>&1; then
    curl -fsSL -o "${dest}" "${url}"
    return
  fi

  if command -v wget >/dev/null 2>&1; then
    wget -qO "${dest}" "${url}"
    return
  fi

  echo "curl or wget is required" >&2
  exit 1
}

for url in \
  https://openaipublic.blob.core.windows.net/encodings/o200k_base.tiktoken \
  https://openaipublic.blob.core.windows.net/encodings/cl100k_base.tiktoken \
  https://openaipublic.blob.core.windows.net/encodings/p50k_base.tiktoken \
  https://openaipublic.blob.core.windows.net/encodings/r50k_base.tiktoken
 do
  tmp_file="${tmp_dir}/$(basename "${url}")"
  download "${url}" "${tmp_file}"

  cache_key=$(printf '%s' "${url}" | sha1sum | awk '{print $1}')
  mv "${tmp_file}" "${cache_dir}/${cache_key}"
done

chmod 644 "${cache_dir}"/*
