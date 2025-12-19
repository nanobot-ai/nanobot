
// this file is generated — do not edit it


/// <reference types="@sveltejs/kit" />

/**
 * Environment variables [loaded by Vite](https://vitejs.dev/guide/env-and-mode.html#env-files) from `.env` files and `process.env`. Like [`$env/dynamic/private`](https://svelte.dev/docs/kit/$env-dynamic-private), this module cannot be imported into client-side code. This module only includes variables that _do not_ begin with [`config.kit.env.publicPrefix`](https://svelte.dev/docs/kit/configuration#env) _and do_ start with [`config.kit.env.privatePrefix`](https://svelte.dev/docs/kit/configuration#env) (if configured).
 * 
 * _Unlike_ [`$env/dynamic/private`](https://svelte.dev/docs/kit/$env-dynamic-private), the values exported from this module are statically injected into your bundle at build time, enabling optimisations like dead code elimination.
 * 
 * ```ts
 * import { API_KEY } from '$env/static/private';
 * ```
 * 
 * Note that all environment variables referenced in your code should be declared (for example in an `.env` file), even if they don't have a value until the app is deployed:
 * 
 * ```
 * MY_FEATURE_FLAG=""
 * ```
 * 
 * You can override `.env` values from the command line like so:
 * 
 * ```sh
 * MY_FEATURE_FLAG="enabled" npm run dev
 * ```
 */
declare module '$env/static/private' {
	export const MANPATH: string;
	export const STARSHIP_SHELL: string;
	export const NIX_PROFILES: string;
	export const TERM_PROGRAM: string;
	export const NODE: string;
	export const PYENV_ROOT: string;
	export const TRANSIENT_PROMPT_MULTILINE_INDICATOR: string;
	export const INIT_CWD: string;
	export const SHELL: string;
	export const TERM: string;
	export const __ETC_PROFILE_NIX_SOURCED: string;
	export const HOMEBREW_REPOSITORY: string;
	export const TMPDIR: string;
	export const PERL5LIB: string;
	export const TERM_PROGRAM_VERSION: string;
	export const npm_config_npm_globalconfig: string;
	export const WINDOWID: string;
	export const PERL_MB_OPT: string;
	export const npm_config_registry: string;
	export const ZSH: string;
	export const USER: string;
	export const LS_COLORS: string;
	export const TRANSIENT_PROMPT_COMMAND_RIGHT: string;
	export const COMMAND_MODE: string;
	export const npm_config_globalconfig: string;
	export const PNPM_SCRIPT_SRC_DIR: string;
	export const PROMPT_INDICATOR_VI_NORMAL: string;
	export const SSH_AUTH_SOCK: string;
	export const __CF_USER_TEXT_ENCODING: string;
	export const PYENV_VIRTUALENV_INIT: string;
	export const npm_execpath: string;
	export const PAGER: string;
	export const DOLLAR: string;
	export const LSCOLORS: string;
	export const ZED_ENVIRONMENT: string;
	export const npm_config_frozen_lockfile: string;
	export const npm_config_verify_deps_before_run: string;
	export const GOLINE: string;
	export const PATH: string;
	export const NU_VERSION: string;
	export const STARSHIP_CONFIG: string;
	export const npm_config_engine_strict: string;
	export const npm_package_json: string;
	export const __CFBundleIdentifier: string;
	export const PWD: string;
	export const npm_command: string;
	export const CMD_DURATION_MS: string;
	export const GOARCH: string;
	export const EDITOR: string;
	export const npm_config__jsr_registry: string;
	export const npm_lifecycle_event: string;
	export const LANG: string;
	export const npm_package_name: string;
	export const NODE_PATH: string;
	export const XPC_FLAGS: string;
	export const NIX_SSL_CERT_FILE: string;
	export const PROMPT_INDICATOR_VI_INSERT: string;
	export const PROMPT_MULTILINE_INDICATOR: string;
	export const GOFILE: string;
	export const GOPACKAGE: string;
	export const npm_config_node_gyp: string;
	export const XPC_SERVICE_NAME: string;
	export const npm_package_version: string;
	export const pnpm_config_verify_deps_before_run: string;
	export const HOME: string;
	export const PYENV_SHELL: string;
	export const SHLVL: string;
	export const GOROOT: string;
	export const HOMEBREW_PREFIX: string;
	export const PROMPT_INDICATOR: string;
	export const GOOS: string;
	export const PERL_LOCAL_LIB_ROOT: string;
	export const LESS: string;
	export const LOGNAME: string;
	export const STARSHIP_SESSION_KEY: string;
	export const ALACRITTY_WINDOW_ID: string;
	export const npm_lifecycle_script: string;
	export const XDG_DATA_DIRS: string;
	export const ZED_TERM: string;
	export const GOPATH: string;
	export const npm_config_user_agent: string;
	export const HOMEBREW_CELLAR: string;
	export const INFOPATH: string;
	export const LAST_EXIT_CODE: string;
	export const OSLogRateLimit: string;
	export const PERL_MM_OPT: string;
	export const COLORTERM: string;
	export const npm_node_execpath: string;
	export const NODE_ENV: string;
}

/**
 * Similar to [`$env/static/private`](https://svelte.dev/docs/kit/$env-static-private), except that it only includes environment variables that begin with [`config.kit.env.publicPrefix`](https://svelte.dev/docs/kit/configuration#env) (which defaults to `PUBLIC_`), and can therefore safely be exposed to client-side code.
 * 
 * Values are replaced statically at build time.
 * 
 * ```ts
 * import { PUBLIC_BASE_URL } from '$env/static/public';
 * ```
 */
declare module '$env/static/public' {
	
}

/**
 * This module provides access to runtime environment variables, as defined by the platform you're running on. For example if you're using [`adapter-node`](https://github.com/sveltejs/kit/tree/main/packages/adapter-node) (or running [`vite preview`](https://svelte.dev/docs/kit/cli)), this is equivalent to `process.env`. This module only includes variables that _do not_ begin with [`config.kit.env.publicPrefix`](https://svelte.dev/docs/kit/configuration#env) _and do_ start with [`config.kit.env.privatePrefix`](https://svelte.dev/docs/kit/configuration#env) (if configured).
 * 
 * This module cannot be imported into client-side code.
 * 
 * ```ts
 * import { env } from '$env/dynamic/private';
 * console.log(env.DEPLOYMENT_SPECIFIC_VARIABLE);
 * ```
 * 
 * > [!NOTE] In `dev`, `$env/dynamic` always includes environment variables from `.env`. In `prod`, this behavior will depend on your adapter.
 */
declare module '$env/dynamic/private' {
	export const env: {
		MANPATH: string;
		STARSHIP_SHELL: string;
		NIX_PROFILES: string;
		TERM_PROGRAM: string;
		NODE: string;
		PYENV_ROOT: string;
		TRANSIENT_PROMPT_MULTILINE_INDICATOR: string;
		INIT_CWD: string;
		SHELL: string;
		TERM: string;
		__ETC_PROFILE_NIX_SOURCED: string;
		HOMEBREW_REPOSITORY: string;
		TMPDIR: string;
		PERL5LIB: string;
		TERM_PROGRAM_VERSION: string;
		npm_config_npm_globalconfig: string;
		WINDOWID: string;
		PERL_MB_OPT: string;
		npm_config_registry: string;
		ZSH: string;
		USER: string;
		LS_COLORS: string;
		TRANSIENT_PROMPT_COMMAND_RIGHT: string;
		COMMAND_MODE: string;
		npm_config_globalconfig: string;
		PNPM_SCRIPT_SRC_DIR: string;
		PROMPT_INDICATOR_VI_NORMAL: string;
		SSH_AUTH_SOCK: string;
		__CF_USER_TEXT_ENCODING: string;
		PYENV_VIRTUALENV_INIT: string;
		npm_execpath: string;
		PAGER: string;
		DOLLAR: string;
		LSCOLORS: string;
		ZED_ENVIRONMENT: string;
		npm_config_frozen_lockfile: string;
		npm_config_verify_deps_before_run: string;
		GOLINE: string;
		PATH: string;
		NU_VERSION: string;
		STARSHIP_CONFIG: string;
		npm_config_engine_strict: string;
		npm_package_json: string;
		__CFBundleIdentifier: string;
		PWD: string;
		npm_command: string;
		CMD_DURATION_MS: string;
		GOARCH: string;
		EDITOR: string;
		npm_config__jsr_registry: string;
		npm_lifecycle_event: string;
		LANG: string;
		npm_package_name: string;
		NODE_PATH: string;
		XPC_FLAGS: string;
		NIX_SSL_CERT_FILE: string;
		PROMPT_INDICATOR_VI_INSERT: string;
		PROMPT_MULTILINE_INDICATOR: string;
		GOFILE: string;
		GOPACKAGE: string;
		npm_config_node_gyp: string;
		XPC_SERVICE_NAME: string;
		npm_package_version: string;
		pnpm_config_verify_deps_before_run: string;
		HOME: string;
		PYENV_SHELL: string;
		SHLVL: string;
		GOROOT: string;
		HOMEBREW_PREFIX: string;
		PROMPT_INDICATOR: string;
		GOOS: string;
		PERL_LOCAL_LIB_ROOT: string;
		LESS: string;
		LOGNAME: string;
		STARSHIP_SESSION_KEY: string;
		ALACRITTY_WINDOW_ID: string;
		npm_lifecycle_script: string;
		XDG_DATA_DIRS: string;
		ZED_TERM: string;
		GOPATH: string;
		npm_config_user_agent: string;
		HOMEBREW_CELLAR: string;
		INFOPATH: string;
		LAST_EXIT_CODE: string;
		OSLogRateLimit: string;
		PERL_MM_OPT: string;
		COLORTERM: string;
		npm_node_execpath: string;
		NODE_ENV: string;
		[key: `PUBLIC_${string}`]: undefined;
		[key: `${string}`]: string | undefined;
	}
}

/**
 * Similar to [`$env/dynamic/private`](https://svelte.dev/docs/kit/$env-dynamic-private), but only includes variables that begin with [`config.kit.env.publicPrefix`](https://svelte.dev/docs/kit/configuration#env) (which defaults to `PUBLIC_`), and can therefore safely be exposed to client-side code.
 * 
 * Note that public dynamic environment variables must all be sent from the server to the client, causing larger network requests — when possible, use `$env/static/public` instead.
 * 
 * ```ts
 * import { env } from '$env/dynamic/public';
 * console.log(env.PUBLIC_DEPLOYMENT_SPECIFIC_VARIABLE);
 * ```
 */
declare module '$env/dynamic/public' {
	export const env: {
		[key: `PUBLIC_${string}`]: string | undefined;
	}
}
