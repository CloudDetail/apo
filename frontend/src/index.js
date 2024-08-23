import React from 'react'
import { createRoot } from 'react-dom/client'
import { Provider } from 'react-redux'
import 'core-js'

import App from './App'
import { store } from './store/store'
import { ToastProvider } from './components/Toast/ToastContext'
import { ConfigProvider, theme } from 'antd'

createRoot(document.getElementById('root')).render(
  <Provider store={store}>
    <ToastProvider>
      <ConfigProvider theme={{ algorithm: theme.darkAlgorithm }}>
        <App />
      </ConfigProvider>
    </ToastProvider>
  </Provider>,
)
