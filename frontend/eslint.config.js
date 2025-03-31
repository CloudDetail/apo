/**
 * Copyright 2025 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

module.exports = [
  {
    files: ['**/*.{js,jsx,ts,tsx}'], // 检查的文件类型
    languageOptions: {
      parser: require('@typescript-eslint/parser'),
      parserOptions: {
        requireConfigFile: false,
        ecmaVersion: 'latest',
        sourceType: 'module',
        ecmaFeatures: {
          jsx: true,
        },
      },
    },

    plugins: {
      react: require('eslint-plugin-react'),
      '@typescript-eslint': require('@typescript-eslint/eslint-plugin'),
      'unused-imports': require('eslint-plugin-unused-imports'),
    },
    rules: {
      'react/react-in-jsx-scope': 'off',
      // 禁止未使用的变量
      '@typescript-eslint/no-unused-vars': [
        'error',
        {
          vars: 'all',
          args: 'after-used',
          ignoreRestSiblings: true,
        },
      ],
      // 自动检测未使用的导入并移除
      'unused-imports/no-unused-imports': 'error',
      // 针对未使用的变量：支持使用 `_` 开头忽略
      'unused-imports/no-unused-vars': [
        'error',
        { vars: 'all', varsIgnorePattern: '^_', args: 'after-used', argsIgnorePattern: '^_' },
      ],
      'no-unused-vars': ['error', { vars: 'all', args: 'after-used', ignoreRestSiblings: true }], //禁止声明未使用的变量
      'no-undef': 'error', //禁止使用未声明的变量
      'react/prop-types': 'warn', // 要求组件使用 PropTypes 进行类型校验。
      '@typescript-eslint/no-unused-vars': ['error'], //禁止未使用的变量。
    },
  },
]
