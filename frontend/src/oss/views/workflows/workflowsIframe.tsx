/**
 * Copyright 2025 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import { Result } from 'antd'
import i18next, { t } from 'i18next'
import { useEffect, useRef, useState, useCallback } from 'react'
import { workflowAnonymousLoginApi } from 'src/core/api/workflows'
import LoadingSpinner from 'src/core/components/Spinner'
import { useUserContext } from 'src/core/contexts/UserContext'

const WorkflowsIframe = ({ src }) => {
  const workflowRef = useRef(null)
  const { user } = useUserContext()
  const language = i18next.language
  const difyToken = useRef(localStorage.getItem('difyToken'))
  const refreshToken = useRef(localStorage.getItem('difyRefreshToken'))

  const [loading, setLoading] = useState(!difyToken || !refreshToken)
  const [error, setError] = useState(false)

  const loginDify = useCallback(async () => {
    setLoading(true)
    setError(false)

    try {
      const res = await workflowAnonymousLoginApi({
        email: `${user.username}@apo.com`,
        language,
        remember_me: true,
      })

      if (res.result === 'success') {
        const { access_token, refresh_token } = res.data
        localStorage.setItem('difyToken', access_token)
        localStorage.setItem('difyRefreshToken', refresh_token)
        difyToken.current = access_token
        refreshToken.current = refresh_token
      } else {
        setError(true)
      }
    } catch (err) {
      console.error('Login error:', err)
      setError(true)
    } finally {
      setLoading(false)
    }
  }, [user.username, language])

  useEffect(() => {
    const handleMessage = (event) => {
      if (event.origin !== window.location.origin) {
        console.warn('CORS error: received message from unknown origin', event.origin)
        return
      }

      if (event.data?.type === 'refresh_token') {
        const { access_token, refresh_token } = event.data.data
        localStorage.setItem('difyToken', access_token)
        localStorage.setItem('difyRefreshToken', refresh_token)
        // setDifyToken(access_token)
        // setRefreshToken(refresh_token)
      }
    }

    window.addEventListener('message', handleMessage)
    return () => {
      window.removeEventListener('message', handleMessage)
    }
  }, [])

  useEffect(() => {
    if ((!difyToken || !refreshToken) && user.username === 'anonymous') {
      loginDify()
    }
  }, [difyToken, refreshToken, user.username, loginDify])

  if (loading) {
    return <LoadingSpinner loading={loading} />
  }

  if (error) {
    return <Result status="warning" title={t('oss/workflow:workflowError')} />
  }

  return (
    <iframe
      ref={workflowRef}
      src={`${src}${src.includes('?') ? '&' : '?'}access_token=${difyToken}&refresh_token=${refreshToken}`}
      width="100%"
      height="100%"
      frameBorder={0}
    />
  )
}

export default WorkflowsIframe
