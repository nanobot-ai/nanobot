export const manifest = (() => {
function __memo(fn) {
	let value;
	return () => value ??= (value = fn());
}

return {
	appDir: "_app",
	appPath: "_app",
	assets: new Set(["robots.txt"]),
	mimeTypes: {".txt":"text/plain"},
	_: {
		client: {start:"_app/immutable/entry/start.BkfFSqSa.js",app:"_app/immutable/entry/app.DCX1v7xh.js",imports:["_app/immutable/entry/start.BkfFSqSa.js","_app/immutable/chunks/D7RuEkZb.js","_app/immutable/chunks/DO59wc33.js","_app/immutable/entry/app.DCX1v7xh.js","_app/immutable/chunks/DO59wc33.js","_app/immutable/chunks/DsnmJJEf.js","_app/immutable/chunks/ik-EZS2y.js","_app/immutable/chunks/PXDOYhLw.js"],stylesheets:[],fonts:[],uses_env_dynamic_public:false},
		nodes: [
			__memo(() => import('./nodes/0.js')),
			__memo(() => import('./nodes/1.js')),
			__memo(() => import('./nodes/2.js')),
			__memo(() => import('./nodes/3.js')),
			__memo(() => import('./nodes/4.js'))
		],
		remotes: {
			
		},
		routes: [
			{
				id: "/",
				pattern: /^\/$/,
				params: [],
				page: { layouts: [0,], errors: [1,], leaf: 2 },
				endpoint: null
			},
			{
				id: "/c/[id]",
				pattern: /^\/c\/([^/]+?)\/?$/,
				params: [{"name":"id","optional":false,"rest":false,"chained":false}],
				page: { layouts: [0,], errors: [1,], leaf: 3 },
				endpoint: null
			},
			{
				id: "/w/[workspaceId]/t/[taskId]",
				pattern: /^\/w\/([^/]+?)\/t\/([^/]+?)\/?$/,
				params: [{"name":"workspaceId","optional":false,"rest":false,"chained":false},{"name":"taskId","optional":false,"rest":false,"chained":false}],
				page: { layouts: [0,], errors: [1,], leaf: 4 },
				endpoint: null
			}
		],
		prerendered_routes: new Set([]),
		matchers: async () => {
			
			return {  };
		},
		server_assets: {}
	}
}
})();
