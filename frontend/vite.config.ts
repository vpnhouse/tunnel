import { defineConfig, loadEnv } from 'vite';
import react from '@vitejs/plugin-react';
import compression from 'vite-plugin-compression';
import { visualizer } from 'rollup-plugin-visualizer';
import checker from 'vite-plugin-checker';
import path from 'path';

export default defineConfig(({ mode }) => {
    const env = loadEnv(mode, process.cwd(), '');

    return {
        plugins: [
            react({
                // Use the new JSX transform
                jsxRuntime: 'automatic'
            }),
            // TypeScript checker - Shows errors in browser overlay (dev only)
            mode === 'development' && checker({
                typescript: true,
                overlay: {
                    initialIsOpen: false,
                    position: 'br' // bottom-right
                }
            }),
            // Gzip compression
            compression({
                algorithm: 'gzip',
                ext: '.gz'
            }),
            // Brotli compression (better than gzip)
            compression({
                algorithm: 'brotliCompress',
                ext: '.br'
            }),
            // Bundle analyzer (only in build)
            mode === 'production' && visualizer({
                filename: 'dist/stats.html',
                open: false,
                gzipSize: true,
                brotliSize: true
            })
        ].filter(Boolean),

        resolve: {
            alias: {
                '@root': path.resolve(__dirname, './src'),
                '@constants': path.resolve(__dirname, './src/constants'),
                '@common': path.resolve(__dirname, './src/common'),
                '@context': path.resolve(__dirname, './src/context'),
                '@apiHooks': path.resolve(__dirname, './src/apiHooks'),
                '@config': path.resolve(__dirname, './src/config.json'),
                '@schema': path.resolve(__dirname, './schema.ts')
            }
        },

        server: {
            port: 8000,
            open: true,
            proxy: {
                '/api': {
                    target: env.API_URL || 'https://vu-de-2.vpnhouse.net',
                    changeOrigin: true
                }
            }
        },

        build: {
            outDir: 'dist',
            sourcemap: false,
            // Optimize chunk size
            rollupOptions: {
                output: {
                    manualChunks: {
                        // Vendor chunks
                        'vendor-react': ['react', 'react-dom', 'react-router-dom'],
                        'vendor-mui': ['@mui/material', '@mui/icons-material', '@mui/lab'],
                        'vendor-mui-pickers': ['@mui/x-date-pickers', 'date-fns'],
                        'vendor-effector': ['effector', 'effector-react'],
                        'vendor-utils': ['clsx', 'uuid', 'qrcode.react', 'js-base64']
                    },
                    // Smaller chunk names
                    chunkFileNames: 'assets/[name]-[hash].js',
                    entryFileNames: 'assets/[name]-[hash].js',
                    assetFileNames: 'assets/[name]-[hash].[ext]'
                }
            },
            // Minification
            minify: 'esbuild',
            // Target modern browsers
            target: 'esnext',
            // CSS code splitting
            cssCodeSplit: true,
            // Report compressed size
            reportCompressedSize: true
        },

        // Optimize dependencies
        optimizeDeps: {
            include: [
                'react',
                'react-dom',
                'react-router-dom',
                '@mui/material',
                '@mui/icons-material',
                'effector',
                'effector-react'
            ]
        },

        // Environment variables
        define: {
            'process.env': JSON.stringify({})
        }
    };
});
