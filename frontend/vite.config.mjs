/**
 * Copyright 2024 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */
import { defineConfig, loadEnv } from 'vite'
import react from '@vitejs/plugin-react'
import path from 'node:path'
import autoprefixer from 'autoprefixer'
import tailwindcss from 'tailwindcss'
import strip from '@rollup/plugin-strip'
export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, process.cwd(), '')
  const isPro = process.env.VITE_APP_VERSION === 'pro'
  const devServerPort = Number(process.env.VITE_DEV_SERVER_PORT || process.env.PORT || 3000)
  const devProxyTarget = process.env.VITE_DEV_PROXY_TARGET || 'http://localhost:18080'
  const apoBackendTarget =
    env.VITE_APO_DEV_BACKEND_TARGET || devProxyTarget
  const hermesBackendTarget =
    env.VITE_HERMES_DEV_BACKEND_TARGET || 'http://127.0.0.1:8100'
  return {
    base: './',
    build: {
      outDir: 'build',
      // sourcemap: true, // 启用 source map
      // rollupOptons: {
      //   // 设置默认入口文件，即使在 pro 环境中不包含 src/oss 文件夹
      //   input: {
      //     main: path.resolve(__dirname, 'index.html'),
      //     ...(isPro ? { oss: path.resolve(__dirname, 'src/oss') } : {}),
      //   },
      //   // 在 pro 环境下启用 treeshake 以排除未使用的代码
      //   treeshake: isPro,
      // },
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
      port: devServerPort,
      proxy: {
        '/api': {
          target: apoBackendTarget,
          changeOrigin: true,
          ws: true,
        },
        '/hermes-api': {
          target: hermesBackendTarget,
          changeOrigin: true,
          rewrite: (pathValue) => pathValue.replace(/^\/hermes-api/, '/api'),
        },
        '/hermes-ws': {
          target: hermesBackendTarget,
          changeOrigin: true,
          ws: true,
          rewrite: (pathValue) => pathValue.replace(/^\/hermes-ws/, '/ws/hermes'),
        },
      },
    },
  }
})
