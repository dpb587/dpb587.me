import { defineConfig } from 'vite'
import { svelte } from '@sveltejs/vite-plugin-svelte'
import tailwindcss from '@tailwindcss/vite';

export default defineConfig({
	build: {
		outDir: '../../static/assets/scripts/',
		emptyOutDir: true,
		lib: {
			entry: 'src/elements',
			name: 'Element',
			fileName: 'elements',
			formats: ['iife'],
		},
	},
	plugins: [
		tailwindcss(),
		svelte(),
	],
})
