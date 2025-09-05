/**
 * Copyright 2024 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */
import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import path from 'node:path'
import autoprefixer from 'autoprefixer'
import tailwindcss from 'tailwindcss'
import strip from '@rollup/plugin-strip'
export default defineConfig(() => {
  return {
    base: './',
    build: {
      outDir: 'build',
    },
    css: {
      postcss: {
        plugins: [
          autoprefixer({}), // add options if needed
          tailwindcss(),
        ],
      },
    },
    esbuild: {
      loader: 'tsx',
      include: /src\/.*\.[tj]sx?$/,
      exclude: [],
    },
    optimizeDeps: {
      force: true,
      esbuildOptions: {
        loader: {
          '.js': 'jsx',
        },
      },
    },
    plugins: [
      react(),
      strip({
        include: ['**/*.ts', '**/*.tsx', '**/*.js', '**/*.jsx'],
        functions: ['console.log', 'console.debug'],
        // sourcemap: false,
      }),
    ],
    resolve: {
      alias: [
        {
          find: 'src/',
          replacement: `${path.resolve(__dirname, 'src')}/`,
        },
        {
          find: 'components/',
          replacement: `${path.resolve(__dirname, 'src/core/components')}/`,
        },
        {
          find: 'core/',
          replacement: `${path.resolve(__dirname, 'src/core')}/`,
        },
        {
          find: 'pro/',
          replacement: `${path.resolve(__dirname, 'src/pro')}/`,
        },
      ],
      extensions: ['.mjs', '.js', '.ts', '.jsx', '.tsx', '.json', '.scss'],
    },
    server: {
      port: 3000,
      proxy: {
        '/api': {
          target: 'http://192.168.1.6:31364',  // 后端服务地址
          changeOrigin: true,
        },
      },
    },
  }
})




