export const manifest = (() => {
function __memo(fn) {
	let value;
	return () => value ??= (value = fn());
}

return {
	appDir: "_app",
	appPath: "_app",
	assets: new Set([]),
	mimeTypes: {},
	_: {
		client: {start:"_app/immutable/entry/start.Cik6yrZp.js",app:"_app/immutable/entry/app.BB-OU12h.js",imports:["_app/immutable/entry/start.Cik6yrZp.js","_app/immutable/chunks/B2UfJDbO.js","_app/immutable/chunks/Jfasuq_0.js","_app/immutable/chunks/JCL9T6X1.js","_app/immutable/entry/app.BB-OU12h.js","_app/immutable/chunks/Jfasuq_0.js","_app/immutable/chunks/DPlzlvFY.js","_app/immutable/chunks/DXHSWRci.js","_app/immutable/chunks/BrsUcSPT.js","_app/immutable/chunks/JCL9T6X1.js"],stylesheets:[],fonts:[],uses_env_dynamic_public:false},
		nodes: [
			__memo(() => import('./nodes/0.js')),
			__memo(() => import('./nodes/1.js')),
			__memo(() => import('./nodes/2.js')),
			__memo(() => import('./nodes/3.js')),
			__memo(() => import('./nodes/4.js')),
			__memo(() => import('./nodes/5.js')),
			__memo(() => import('./nodes/6.js')),
			__memo(() => import('./nodes/7.js'))
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
				id: "/nfs",
				pattern: /^\/nfs\/?$/,
				params: [],
				page: { layouts: [0,], errors: [1,], leaf: 3 },
				endpoint: null
			},
			{
				id: "/samba",
				pattern: /^\/samba\/?$/,
				params: [],
				page: { layouts: [0,], errors: [1,], leaf: 4 },
				endpoint: null
			},
			{
				id: "/scsi",
				pattern: /^\/scsi\/?$/,
				params: [],
				page: { layouts: [0,], errors: [1,], leaf: 5 },
				endpoint: null
			},
			{
				id: "/storage",
				pattern: /^\/storage\/?$/,
				params: [],
				page: { layouts: [0,], errors: [1,], leaf: 6 },
				endpoint: null
			},
			{
				id: "/system",
				pattern: /^\/system\/?$/,
				params: [],
				page: { layouts: [0,], errors: [1,], leaf: 7 },
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
