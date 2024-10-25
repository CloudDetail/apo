// src/api/request.js
import axios from 'axios'
import { useToast } from 'src/components/Toast/ToastContext'
import { showToast } from './toast'
import qs from 'qs'
// 创建axios实例
const instance = axios.create({
  baseURL: '', // 替换为你的API基础URL
  timeout: 30000,
  headers: { 'Content-Type': 'application/json' },
  paramsSerializer: (params) => qs.stringify(params, { arrayFormat: 'repeat' }),
})

// 请求拦截器
instance.interceptors.request.use(
  (config) => {
    // 在发送请求之前做些什么，比如添加token
    const token = localStorage.getItem('token')
    if (token) {
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
    // 对响应数据做点什么
    const { data } = response
    return data
  },
  (error) => {
    // 对响应错误做点什么
    if (error.response) {
      // 请求已发出，但服务器响应状态码不在2xx范围内
      const { status, data } = error.response
      switch (status) {
        case 401:
          //   message.error('未授权，请登录');
          // 可以跳转到登录页
          break
        case 403:
          showToast({
            title: '拒绝访问',
            color: 'danger',
          })
          break
        // case 404:
        //   message.error('请求地址出错');
        //   break;
        // case 500:
        //   message.error('服务器内部错误');
        //   break;
        default:
        
            showToast({
              title: '请求失败',
              message:data.message,
                color: 'danger',
              })
      }
    } else {
      // 一些错误是在设置请求的时候触发的
      showToast({
        title: error.message,
        color: 'danger',
      })
    //   message.error(error.message)
    }
    console.log(error)
    return Promise.reject(error)
  },
)

// 封装GET请求
const get = (url, params = {}, config = {}) => {
  return instance.get(url, { params, ...config }).catch((error) => {
    throw error
  })
}

// 封装POST请求
const post = (url, data = {}, config = {}) => {
  return instance.post(url, data, { ...config }).catch((error) => {
    // 在此处可以捕获到错误信息
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

// 导出axios实例和封装的请求方法
export { instance as axiosInstance, get, post, del }
