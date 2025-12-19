import tailwindcss from '@tailwindcss/vite';
import { sveltekit } from '@sveltejs/kit/vite';
import { defineConfig } from 'vite';

export default defineConfig({
	plugins: [
		tailwindcss(),
		sveltekit(),
		{
			name: 'services',
			async configureServer(server) {
				server.middlewares.use(async (req, res, next) => {
					const m = await server.ssrLoadModule('@nanobot-ai/services');
					const services = m.default.requestListener();
					if (req.url?.startsWith('/mcp')) {
						services(req, res);
					} else {
						next();
					}
				});
			}
		}
	]
});
