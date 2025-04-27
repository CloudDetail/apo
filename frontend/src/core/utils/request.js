/**
 * Copyright 2024 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

// src/api/request.js
import axios from 'axios'
import { showToast } from './toast'
import qs from 'qs'
import TranslationCom from 'src/oss/components/TranslationCom'
import i18next from 'i18next'

const namespace = 'core/login'
const MAX_RETRY_ATTEMPTS = 3
const RETRY_DELAY_BASE = 300 // 300ms
let isTokenRefreshing = false
let pendingRequests = []
let lastTokenRefreshTime = 0

// 创建axios实例
const instance = axios.create({
  baseURL: '', // 替换为你的API基础URL
  timeout: 120000,
  headers: { 'Content-Type': 'application/json' },
  paramsSerializer: (params) => qs.stringify(params, { arrayFormat: 'repeat' }),
})

// 请求拦截器
instance.interceptors.request.use(
  (config) => {
    // 在发送请求之前做些什么，比如添加token
    const token = localStorage.getItem('token')
    if (config.url === '/dify/console/api/logout') {
      const difyToken = localStorage.getItem('difyToken')
      config.headers.Authorization = `Bearer ${difyToken}`
    } else if (token && config.url != '/api/user/refresh') {
      config.headers.Authorization = `Bearer ${token}`
    }
    config.headers['APO-Language'] = i18next.language
    config.metadata = {
      ...config.metadata,
      requestTimestamp: Date.now(),
      retryCount: config.metadata?.retryCount || 0
    }
    return config
  },
  (error) => {
    // 对请求错误做些什么
    return Promise.reject(error)
  },
)

// 响应拦截器
instance.interceptors.response.use(
  (response) => {
    const { headers, data, config } = response
    // 检查预期类型
    const expectedType = config.headers.Accept || 'application/json' // 默认期望 JSON
    const responseType = headers['content-type'] || ''

    // 如果期望是 JSON，但返回的却是 HTML，则认为异常
    if (!expectedType.includes('text/html') && responseType.includes('text/html')) {
      return Promise.reject(new Error('Unexpected HTML response for a JSON request'))
    }
    if (config.responseType === 'blob') {
      return response
    }
    return data
  },
  async (error) => {
    if (error.response) {
      const { status, data } = error.response
      const originalRequest = error.config

      if (status === 400 && data.code === 'A0005') {
        const isStaleRequest = originalRequest.metadata?.requestTimestamp < lastTokenRefreshTime
        if (isStaleRequest) {
          if (originalRequest.metadata.retryCount >= MAX_RETRY_ATTEMPTS) {
            return Promise.reject(new Error('Max retry attempts reached'))
          }
          const delay = RETRY_DELAY_BASE * Math.pow(2, error.config.metadata.retryCount)
          error.config.metadata.retryCount += 1

          return new Promise((resolve) => {
            setTimeout(() => {
              console.log(`Retry attempt ${error.config.metadata.retryCount} for ${error.status} ${error.config.url}`)
              resolve(instance(error.config))
            }, delay)
          })
        }

        if (!isTokenRefreshing) {
          isTokenRefreshing = true
          try {
            const newToken = await refreshAccessToken()
            lastTokenRefreshTime = Date.now()
            instance.defaults.headers.common.Authorization = `Bearer ${newToken}`
            pendingRequests.forEach(callback => callback(newToken))
            return instance(originalRequest)
          } catch (refreshError) {
            localStorage.removeItem('token')
            localStorage.removeItem('refresh_token')
            delete instance.defaults.headers.common.Authorization
            window.location.href = '/#/login'
            return Promise.reject(refreshError)
          } finally {
            isTokenRefreshing = false
            pendingRequests = []
          }
        }

        return new Promise((resolve) => {
          pendingRequests.push((newToken) => {
            originalRequest.headers.Authorization = `Bearer ${newToken}`
            resolve(instance(originalRequest))
          })
        })
      }

      switch (status) {
        case 400:
          if (data.code === 'A0004') {
            window.location.href = '/#/login'
            showToast({
              title: <TranslationCom text="request.notLoggedIn" space={namespace} />,
              color: 'danger',
            })
          } else {
            showToast({
              title: data.message,
              color: 'danger',
            })
          }
          break

        case 401:
          break

        case 403:
          showToast({
            title: <TranslationCom text="request.accessDenied" space={namespace} />,
            color: 'danger',
          })
          break

        default:
          showToast({
            title: <TranslationCom text="request.requestFailed" space={namespace} />,
            message: data.message,
            color: 'danger',
          })
      }
    } else {
      showToast({
        title: error.message,
        color: 'danger',
      })
    }
    return Promise.reject(error)
  },
)

// 刷新 accessToken
export const refreshAccessToken = async () => {
  const refreshToken = localStorage.getItem('refreshToken')
  if (!refreshToken) return null

  try {
    // 使用 instance 实例发送请求并排除 Authorization 头部
    const response = await instance.get(`/api/user/refresh`, {
      headers: {
        Authorization: `Bearer ${refreshToken}`,
      },
    })
    const { accessToken } = response
    localStorage.setItem('token', accessToken)
    refreshDifyAccessToken().catch((error) => {
      console.error(' Dify Token Error:', error)
    })
    return accessToken
  } catch (error) {
    return null
  }
}
export const refreshDifyAccessToken = async () => {
  try {
    const refresh_token = globalThis.localStorage.getItem('difyRefreshToken')

    const res = await fetch(`/dify/console/api/refresh-token`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json;utf-8',
      },
      body: JSON.stringify({ refresh_token }),
    })
    if (!res.ok) {
      if (res.status === 401) {
        return Promise.reject('Unauthorized: Dify refresh token invalid')
      }
      return Promise.reject(`Request failed with status: ${res.status}`)
    }

    const { data } = await res.json()
    localStorage.setItem('difyToken', data.access_token)
    localStorage.setItem('difyRefreshToken', data.refresh_token)
  } catch (error) {
    console.error(error)
    return Promise.reject(error)
  } finally {
  }
}
// 封装GET请求
const get = (url, params = {}, config = {}) => {
  return instance
    .get(url, { params, ...config })
    .then((response) => {
      console.log(response)
      // if (config?.responseType === 'blob') return response
      return response
    })
    .catch((error) => {
      throw error
    })
}

// 封装POST请求
const post = (url, data = {}, config = {}) => {
  return instance.post(url, data, config).catch((error) => {
    throw error
  })
}

// 封装DELETE请求
const del = (url, data = {}, config = {}) => {
  return instance.delete(url, { data, ...config }).catch((error) => {
    // 在此处可以捕获到错误信息
    throw error
  })
}

// 封装常用的请求头配置
const headers = {
  formUrlencoded: {
    headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
  },
}

// 导出axios实例和封装的请求方法
export { instance as axiosInstance, get, post, del, headers }
