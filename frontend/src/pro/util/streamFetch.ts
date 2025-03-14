import i18next from 'i18next'
import { refreshAccessToken } from 'src/core/utils/request'
import { showToast } from 'src/core/utils/toast'

const streamFetch = async (url: string, options: any = {}) => {
  const token = localStorage.getItem('token') // 从本地存储获取 token
  const headers = {
    ...options.headers,
    'Content-Type': 'application/json',
    Authorization: token ? `Bearer ${token}` : undefined,
    'APO-Language':i18next.language
  }

  try {
    const response = await fetch(url, {
      ...options,
      headers,
    })

    if (!response.ok) {
      const errorData = await response.json()
      const { status } = response

      switch (status) {
        case 400:
          if (errorData.code === 'A0004') {
            window.location.href = '/#/login'
            showToast({
              title: '未登录,请先登录',
              color: 'danger',
            })
          } else if (errorData.code === 'A0005') {
            const newToken = await refreshAccessToken()
            if (newToken) {
              localStorage.setItem('token', newToken)
              return streamFetch(url, options) // 递归调用需返回结果
            } else {
              window.location.href = '/#/login'
              showToast({
                title: '登录过期,请重新登录',
                color: 'danger',
              })
            }
          } else {
            showToast({
              title: errorData.message || '请求错误',
              color: 'danger',
            })
          }
          break

        case 401:
          window.location.href = '/#/login'
          showToast({
            title: '认证失败，请重新登录',
            color: 'danger',
          })
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
            message: errorData.message || '未知错误',
            color: 'danger',
          })
          break
      }

      // throw new Error(`请求失败，状态码：${status}`)
    }

    // 处理成功响应
    return response
  } catch (error) {
    console.error('Fetch 请求出错:', error?.message || error)
    throw error
  }
}
export default streamFetch
