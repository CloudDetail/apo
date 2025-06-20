/**
 * Copyright 2024 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import i18n from 'i18next'
import { initReactI18next } from 'react-i18next'
import Backend from 'i18next-http-backend'
import LanguageDetector from 'i18next-browser-languagedetector'

const coreNsList = [
  'mask',
  'alertsIntegration',
  'dataGroup',
  'team',
  'permission',
  'dataIntegration',
  'login',
]
const getNsList = () => {
  return coreNsList.map((ns) => `core/${ns}`)
}
i18n
  .use(Backend) // 支持后端动态加载
  .use(LanguageDetector) // 自动检测用户语言
  .use(initReactI18next) // 绑定 React
  .init({
    fallbackLng: 'en',
    supportedLngs: ['en', 'zh'],
    nonExplicitSupportedLngs: true,
    backend: {
      loadPath: '/locales/{{lng}}/{{ns}}.json', // 翻译文件路径
    },
    ns: ['common', 'oss', 'core', 'pro', ...getNsList()], // 命名空间
    defaultNS: 'common', // 默认命名空间
    interpolation: {
      escapeValue: false, // 防止 XSS，React 已经默认转义
    },
    detection: {
      order: ['querystring', 'cookie', 'localStorage', 'navigator'], // 检测语言顺序
      caches: ['localStorage', 'cookie'], // 缓存语言到 localStorage 和 cookie
    },
  })

export default i18n
