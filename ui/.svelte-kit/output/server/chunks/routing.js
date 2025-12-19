function affects_path(segment) {
  return segment !== "" && !/^\([^)]+\)$/.test(segment);
}
function get_route_segments(route) {
  return route.slice(1).split("/").filter(affects_path);
}
function exec(match, params, matchers) {
  const result = {};
  const values = match.slice(1);
  const values_needing_match = values.filter((value) => value !== void 0);
  let buffered = 0;
  for (let i = 0; i < params.length; i += 1) {
    const param = params[i];
    let value = values[i - buffered];
    if (param.chained && param.rest && buffered) {
      value = values.slice(i - buffered, i + 1).filter((s) => s).join("/");
      buffered = 0;
    }
    if (value === void 0) {
      if (param.rest) result[param.name] = "";
      continue;
    }
    if (!param.matcher || matchers[param.matcher](value)) {
      result[param.name] = value;
      const next_param = params[i + 1];
      const next_value = values[i + 1];
      if (next_param && !next_param.rest && next_param.optional && next_value && param.chained) {
        buffered = 0;
      }
      if (!next_param && !next_value && Object.keys(result).length === values_needing_match.length) {
        buffered = 0;
      }
      continue;
    }
    if (param.optional && param.chained) {
      buffered++;
      continue;
    }
    return;
  }
  if (buffered) return;
  return result;
}
const basic_param_pattern = /\[(\[)?(\.\.\.)?(\w+?)(?:=(\w+))?\]\]?/g;
function resolve_route(id, params) {
  const segments = get_route_segments(id);
  return "/" + segments.map(
    (segment) => segment.replace(basic_param_pattern, (_, optional, rest, name) => {
      const param_value = params[name];
      if (!param_value) {
        if (optional) return "";
        if (rest && param_value !== void 0) return "";
        throw new Error(`Missing parameter '${name}' in route ${id}`);
      }
      if (param_value.startsWith("/") || param_value.endsWith("/"))
        throw new Error(
          `Parameter '${name}' in route ${id} cannot start or end with a slash -- this would cause an invalid route like foo//bar`
        );
      return param_value;
    })
  ).filter(Boolean).join("/");
}
export {
  exec as e,
  resolve_route as r
};
