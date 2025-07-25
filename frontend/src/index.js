/**
 * Copyright 2024 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import React, { useState, useEffect, memo } from 'react'
import { createRoot } from 'react-dom/client'
import { Provider, useSelector } from 'react-redux'
import 'core-js'

import App from './App'
import { store } from 'src/core/store/store'
import { ToastProvider } from 'src/core/components/Toast/ToastContext'
import { ConfigProvider, notification, theme } from 'antd'

import posthog from 'posthog-js'
import { PostHogProvider } from 'posthog-js/react'
import { MessageProvider } from 'src/core/contexts/MessageContext'
import ErrorBoundary from 'src/core/components/ErrorBoundary'
import zhCN from 'antd/es/locale/zh_CN'
import enUS from 'antd/es/locale/en_US'
import { UserProvider } from './core/contexts/UserContext'
import './i18n'
import { useTranslation } from 'react-i18next'
import { loader } from '@monaco-editor/react'
import * as monaco from 'monaco-editor'
import { setNotifyApi } from './core/utils/notify'
import { DataGroupProvider } from './core/contexts/DataGroupContext'

loader.config({ monaco })

const apiHost = import.meta.env.VITE_PUBLIC_POSTHOG_HOST
const apiKey = import.meta.env.VITE_PUBLIC_POSTHOG_KEY

posthog.init(apiKey, {
  api_host: apiHost,
  person_profiles: 'identified_only',
})
function getThemeColor(mode) {
  try {
    const config = window.__APP_CONFIG__ || {}
    return (
      (config.themes && config.themes[mode] && config.themes[mode].colorPrimary) || '#1677ff' // 默认颜色
    )
  } catch (e) {
    return '#1677ff'
  }
}

const AntdWrapper = memo(() => {
  const { i18n } = useTranslation()
  const [locale, setLocale] = useState(zhCN)
  const [colorBgBase, setColorBgBase] = useState()
  useEffect(() => {
    setLocale(i18n.language === 'en' ? enUS : zhCN)
  }, [i18n.language])
  const state = useSelector((state) => state.settingReducer)
  const { theme: storeTheme } = state
  const lightColor = getThemeColor(storeTheme)

  useEffect(() => {
    if (storeTheme === 'light') {
      document.documentElement.style.setProperty('--active-collapse-bg', lightColor)
    } else {
      document.documentElement.style.setProperty('--active-collapse-bg', '#285587')
    }
    setColorBgBase(getComputedStyle(document.documentElement).getPropertyValue('--body-bg').trim())
  }, [storeTheme])
  const [api, contextHolder] = notification.useNotification()

  useEffect(() => {
    setNotifyApi(api)
  }, [api])

  return (
    <ConfigProvider
      locale={locale}
      theme={{
        algorithm: storeTheme === 'light' ? theme.defaultAlgorithm : theme.darkAlgorithm,
        token: {
          colorPrimary: getThemeColor(storeTheme),
          colorInfo: getThemeColor(storeTheme),
          colorLink: getThemeColor(storeTheme),
          colorBgLayout: colorBgBase,
        },
        cssVar: true,
        components: {
          // Segmented: {
          //   itemSelectedBg: '#4096ff',
          // },
          Segmented: {
            // itemActiveBg: 'var(--ant-color-bg-layout)',
            // itemSelectedBg: 'var(--ant-color-bg-layout)',
            trackBg: 'var(--body-bg)',
            itemSelectedColor: 'var(--ant-color-primary-text)',
            // itemColor: 'rgba(255,255,255, 0.4)',
          },
          Layout: {
            bodyBg: 'var(--body-bg)',
            siderBg: 'var(--color-sider)',
          },
          Tree: {
            nodeSelectedBg: '#33415580',
          },
          Table: {
            headerBg: 'var(--color-table-bg)',
            cellFontSizeSM: 12,
          },
          Breadcrumb: {
            itemColor: 'var(--color-text)',
            linkColor: 'var(--color-text)',
          },
          Menu: {
            itemBg: 'var(--color-sider)',
            darkItemBg: 'var(--color-sider)',
            itemSelectedBg: getThemeColor(storeTheme),
            itemSelectedColor: 'var(--menu-selected-text-color)',
          },
          Spin: {
            dotSizeLG: 48,
          },
          Splitter: {
            splitTriggerSize: 12,
          },
        },
      }}
    >
      <MessageProvider>
        <UserProvider>
          <DataGroupProvider>
            {contextHolder}
            <App />
          </DataGroupProvider>
        </UserProvider>
      </MessageProvider>
    </ConfigProvider>
  )
})
const AppWrapper = memo(() => {
  return (
    <ErrorBoundary>
      <Provider store={store}>
        <ToastProvider>
          <AntdWrapper />
        </ToastProvider>
      </Provider>
    </ErrorBoundary>
  )
})

createRoot(document.getElementById('root')).render(
  apiKey && apiHost ? (
    <PostHogProvider client={posthog}>
      <AppWrapper />
    </PostHogProvider>
  ) : (
    <AppWrapper />
  ),
)
