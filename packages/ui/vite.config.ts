import tailwindcss from '@tailwindcss/vite';
import { sveltekit } from '@sveltejs/kit/vite';
import { defineConfig } from 'vite';
import { createProxyMiddleware } from 'http-proxy-middleware';

export default defineConfig({
	plugins: [
		tailwindcss(),
		sveltekit(),
		{
			name: 'services',
			async configureServer(server) {
				// Proxy /mcp?ui and /mcp/ui requests to nanobot server
				const nanobotProxy = createProxyMiddleware({
					target: 'http://localhost:8080',
					changeOrigin: true,
				});

				server.middlewares.use(async (req, res, next) => {
					// Check if this is a UI MCP request that should go to nanobot
					const url = req.url || '';
					if (url.includes('/mcp') && (url.includes('ui') || url.includes('?ui'))) {
						return nanobotProxy(req, res, next);
					}

					// Proxy /api/* requests to nanobot server
					if (url.startsWith('/api/')) {
						return nanobotProxy(req, res, next);
					}

					// Handle other /mcp/* paths with local services
					if (url.startsWith('/mcp')) {
						const m = await server.ssrLoadModule('@nanobot-ai/services');
						const services = m.default.requestListener();
						services(req, res);
					} else {
						next();
					}
				});
			}
		}
	]
});
