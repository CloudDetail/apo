import React from 'react'
import { createRoot } from 'react-dom/client'
import { Provider } from 'react-redux'
import 'core-js'

import App from './App'
import { store } from './store/store'
import { ToastProvider } from './components/Toast/ToastContext'
import { UrlParamsProvider } from './contexts/UrlParamsContext'

createRoot(document.getElementById('root')).render(
  <Provider store={store}>
    <ToastProvider>
      <UrlParamsProvider>
        <App />
      </UrlParamsProvider>
    </ToastProvider>
  </Provider>,
)
