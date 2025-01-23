/**
 * Copyright 2024 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import path from 'node:path'
import autoprefixer from 'autoprefixer'
import tailwindcss from 'tailwindcss'
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
      loader: 'tsx', // 主要是支持 TSX/JSX
      include: /src\/.*\.[tj]sx?$/, // 包含 .js, .jsx, .ts, .tsx 文件
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
    plugins: [react()],
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
      proxy: {},
    },
  }
})
