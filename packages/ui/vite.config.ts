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
				const m = await server.ssrLoadModule('@nanobot-ai/services');
				const services = m.default.requestListener();
				server.middlewares.use((req, res, next) => {
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
