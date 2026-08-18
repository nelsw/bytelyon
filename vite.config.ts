import inertia from '@inertiajs/vite';
import { wayfinder } from '@laravel/vite-plugin-wayfinder';
import tailwindcss from '@tailwindcss/vite';
import vue from '@vitejs/plugin-vue';
import laravel from 'laravel-vite-plugin';
import { bunny } from 'laravel-vite-plugin/fonts';
import { defineConfig } from 'vite';

export default defineConfig({
    plugins: [
        laravel({
            input: ['resources/css/app.css', 'resources/js/app.ts'],
            refresh: true,
            fonts: [
                // Display face — headings only.
                bunny('Space Grotesk', {
                    weights: [500, 600, 700],
                }),
                // UI and body copy.
                bunny('Manrope', {
                    weights: [400, 500, 600, 700, 800],
                }),
                // Metrics, IDs, and anything numeric or technical.
                bunny('JetBrains Mono', {
                    weights: [400, 500, 600],
                }),
            ],
        }),
        inertia(),
        tailwindcss(),
        vue({
            template: {
                transformAssetUrls: {
                    base: null,
                    includeAbsolute: false,
                },
            },
        }),
        wayfinder({
            formVariants: true,
        }),
    ],
    build: {
        rollupOptions: {
            onwarn(warning, warn) {
                if (warning.code === 'INVALID_ANNOTATION') {
                    // Suppress erroneous PURE comment warning
                    return;
                }
            },
        },
        manifest: 'manifest.json',
    },
});
