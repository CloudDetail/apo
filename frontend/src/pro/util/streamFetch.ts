import i18next, { t } from 'i18next'
import { notify } from 'src/core/utils/notify'
import { refreshAccessToken } from 'src/core/utils/request'

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
            notify({
              message: t('core/login:request.notLoggedIn'),
              type: 'error',
            })
          } else if (errorData.code === 'A0005') {
            const newToken = await refreshAccessToken()
            if (newToken) {
              localStorage.setItem('token', newToken)
              return streamFetch(url, options)
            } else {
              window.location.href = '/#/login'
              notify({
                message: t('core/login:request.loginExpired'),
                type: 'error',
              })
            }
          } else {
            notify({
              message: errorData.message || t('core/login:request.requestFailed'),
              type: 'error',
            })
          }
          break

        case 401:
          window.location.href = '/#/login'
          notify({
            message: t('core/login:request.notLoggedIn'),
            type: 'error',
          })
          break

        case 403:
          notify({
            message: t('core/login:request.accessDenied'),
            type: 'error',
          })
          break

        default:
          notify({
            message: errorData.message || t('core/login:request.requestFailed'),
            type: 'error',
          })
          break
      }
    }

    return response
  } catch (error) {
    console.error('Fetch 请求出错:', error?.message || error)
    throw error
  }
}
export default streamFetch
