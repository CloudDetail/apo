// src/api/request.js
import axios from 'axios'
import { showToast } from './toast'
import FormData from 'form-data';
import qs from 'qs'

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
    if (token && config.url != "/api/user/refresh") {
      config.headers.Authorization = `Bearer ${token}`
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
    const { data } = response
    return data
  },
  async (error) => {
    if (error.response) {
      const { status, data } = error.response
      const originalRequest = error.config
      switch (status) {
        case 400:
          if (data.code === 'A0004') {
            window.location.href = "/#/login"
            showToast({
              title: "未登录,请先登录",
              color: 'danger'
            })
          } else if (data.code === 'A0005') {
            const newToken = await refreshAccessToken()
            if (newToken) {
              originalRequest.headers.Authorization = `Bearer ${newToken}`
              return instance(originalRequest)
            } else {
              window.location.href = "/#/login"
              showToast({
                title: "登录过期,请重新登录",
                color: 'danger'
              })
            }
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
            title: '拒绝访问',
            color: 'danger',
          })
          break

        default:
          showToast({
            title: '请求失败',
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
const refreshAccessToken = async () => {
  const refreshToken = localStorage.getItem('refreshToken')
  if (!refreshToken) return null

  try {
    // 使用 instance 实例发送请求并排除 Authorization 头部
    const response = await instance.get(`/api/user/refresh`, {
      headers: {
        Authorization: `Bearer ${refreshToken}`,
      }
    })
    // @ts-ignore
    const { accessToken } = response
    localStorage.setItem('token', accessToken)
    return accessToken
  } catch (error) {
    return null
  }
}

// 封装GET请求
const get = (url, params = {}, config = {}) => {
  return instance.get(url, { params, ...config }).catch((error) => {
    throw error
  })
}

// 封装POST请求
const post = (url, data = {}, config = {}, form = false) => {
  if (form) {
    const formData = new URLSearchParams();
    Object.keys(data).forEach(key => {
      formData.append(key, data[key]);
    });
    data = formData;
  }

  const headers = form ? {
    ...config.headers,
    'Content-Type': 'application/x-www-form-urlencoded',
  } : config.headers;

  const requestConfig = { ...config, headers };

  // 发送 POST 请求
  return instance.post(url, form ? data.toString() : data, requestConfig)
    .catch((error) => {
      // 捕获错误信息
      throw error;
    });
};


// 封装DELETE请求
const del = (url, data = {}, config = {}) => {
  return instance.delete(url, { data, ...config }).catch((error) => {
    // 在此处可以捕获到错误信息
    throw error
  })
}

// 导出axios实例和封装的请求方法
export { instance as axiosInstance, get, post, del }
