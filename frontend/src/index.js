/**
 * Copyright 2024 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import React, { useState, useEffect } from 'react'
import { createRoot } from 'react-dom/client'
import { Provider } from 'react-redux'
import 'core-js'

import App from './App'
import { store } from 'src/core/store/store'
import { ToastProvider } from 'src/core/components/Toast/ToastContext'
import { ConfigProvider, theme } from 'antd'

import posthog from 'posthog-js'
import { PostHogProvider } from 'posthog-js/react'
import { MessageProvider } from 'src/core/contexts/MessageContext'
import ErrorBoundary from 'src/core/components/ErrorBoundary'
import zhCN from 'antd/es/locale/zh_CN'
import enUS from 'antd/es/locale/en_US'
import { UserProvider } from './core/contexts/UserContext'
import './i18n'
import { useTranslation } from 'react-i18next'

const apiHost = import.meta.env.VITE_PUBLIC_POSTHOG_HOST
const apiKey = import.meta.env.VITE_PUBLIC_POSTHOG_KEY

posthog.init(apiKey, {
  api_host: apiHost,
  person_profiles: 'identified_only',
})

const AppWrapper = () => {
  const { i18n } = useTranslation()
  const [locale, setLocale] = useState(zhCN)

  useEffect(() => {
    setLocale(i18n.language === 'en' ? enUS : zhCN)
  }, [i18n.language])

  return (
    <ErrorBoundary>
      <Provider store={store}>
        <ToastProvider>
          <ConfigProvider
            locale={locale}
            theme={{
              algorithm: theme.darkAlgorithm,
              components: {
                // Segmented: {
                //   itemSelectedBg: '#4096ff',
                // },
                Segmented: {
                  itemActiveBg: '#1c2b4a',
                  itemSelectedBg: '#1c2b4a',
                  trackBg: '#1e2635',
                  itemSelectedColor: '#4d82ff',
                  itemColor: 'rgba(255,255,255, 0.4)',
                },
                Layout: {
                  bodyBg: '#1d222b',
                  siderBg: '#1d222b',
                },
                Tree: {
                  nodeSelectedBg: '#33415580',
                },
              },
            }}
          >
            <MessageProvider>
              <UserProvider>
                <App />
              </UserProvider>
            </MessageProvider>
          </ConfigProvider>
        </ToastProvider>
      </Provider>
    </ErrorBoundary>
  )
}

createRoot(document.getElementById('root')).render(
  apiKey && apiHost ? (
    <PostHogProvider client={posthog}>
      <AppWrapper />
    </PostHogProvider>
  ) : (
    <AppWrapper />
  ),
)
