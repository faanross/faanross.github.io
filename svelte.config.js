import adapter from '@sveltejs/adapter-static';
import { vitePreprocess } from '@sveltejs/vite-plugin-svelte';
import { mdsvex } from 'mdsvex';
import { fileURLToPath } from 'url';
import { dirname, join } from 'path';

const __filename = fileURLToPath(import.meta.url);
const __dirname = dirname(__filename);

/** @type {import('@sveltejs/kit').Config} */
const config = {
	extensions: ['.svelte', '.md'],

	preprocess: [
		vitePreprocess(),
		mdsvex({
			extensions: ['.md'],
			layout: {
				workshop: join(__dirname, 'src/lib/layouts/workshop.svelte'),
				workshop02: join(__dirname, 'src/lib/layouts/workshop02.svelte'),
				workshop03: join(__dirname, 'src/lib/layouts/workshop03.svelte'),
				course01: join(__dirname, 'src/lib/layouts/course01.svelte'),
				reflective: join(__dirname, 'src/lib/layouts/reflective.svelte')
			}
		})
	],

	kit: {
		adapter: adapter({
			pages: 'build',
			assets: 'build',
			fallback: '404.html',
			precompress: false,
			strict: true
		}),
		paths: {
			base: ''
		}
	}
};

export default config;
