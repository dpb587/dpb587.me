import { defineConfig } from 'vite'
import { svelte } from '@sveltejs/vite-plugin-svelte'
import tailwindcss from '@tailwindcss/vite';
import { readlinkSync } from 'fs';
import { resolve } from 'path';

const staticAssetsDir = resolve(import.meta.dirname, '../../static/assets');

function resolveAssetPath(pkg, symlink) {
	const target = readlinkSync(resolve(staticAssetsDir, pkg, symlink));
	return JSON.stringify(`/assets/${pkg}/${target}`);
}

export default defineConfig({
	define: {
		__OPENSEADRAGON_PATH__: resolveAssetPath('openseadragon', 'openseadragon-current'),
		__PANNELLUM_PATH__: resolveAssetPath('pannellum', 'pannellum-current'),
		__PLYR_PATH__: resolveAssetPath('plyr', 'plyr-current'),
		__HLSJS_PATH__: resolveAssetPath('hlsjs', 'hlsjs-current'),
	},
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
