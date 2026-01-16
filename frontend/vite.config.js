import { sveltekit } from '@sveltejs/kit/vite';
import { defineConfig } from 'vite';

export default defineConfig({
	plugins: [sveltekit()],
	server: {
		port: 5173,
		host: true
	},
	define: {
		__APP_VERSION__: JSON.stringify(process.env.APP_VERSION || 'v1.0.0-dev')
	}
});
