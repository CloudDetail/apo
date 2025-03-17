/**
 * Copyright 2025 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import { Button, Result } from 'antd'
import i18next, { t } from 'i18next'
import { useEffect, useRef, useState } from 'react'
import { workflowAnonymousLoginApi } from 'src/core/api/workflows'
import LoadingSpinner from 'src/core/components/Spinner'
import { useUserContext } from 'src/core/contexts/UserContext'

const WorkflowsIframe = ({ src }) => {
  const language = i18next.language
  const workflowRef = useRef(null)
  const intervalRef = useRef<any>(null)
  const { user } = useUserContext()

  const [difyToken, setDifyToken] = useState(localStorage.getItem('difyToken'))
  const [loading, setLoading] = useState(!difyToken)
  const [error, setError] = useState(false)

  useEffect(() => {
    if (user.username === 'anonymous') {
      loginDify(user)
    }
  }, [difyToken, user.username])

  useEffect(() => {
    if (error) return


  const sendMessageToB = (data) => {
    if (workflowRef.current) {
      workflowRef.current.contentWindow.postMessage(data, '*')
    }
  }

  const handleMessage = (event: MessageEvent) => {
    if (event.data === 'got' && intervalRef.current) {
      clearInterval(intervalRef.current)
      intervalRef.current = null
    }
  }

  let data;
  if (!difyToken) {
    data = { action: 'clear-token' }
  } else {
    data = {
      action: 'auto-login',
      data: {
        token: difyToken,
        refreshToken: localStorage.getItem('difyRefreshToken'),
      },
      src: src.slice(5),
    }
  }

  // 发送数据，直到 B 页面返回 'got'
  intervalRef.current = setInterval(() => sendMessageToB(data), 1000)
    window.addEventListener('message', handleMessage)

    return () => {
      if (intervalRef.current) {
        clearInterval(intervalRef.current)
      }
      window.removeEventListener('message', handleMessage)
    }
  }, [difyToken, error])

  const loginDify = async (user) => {
    setLoading(true) 
    setError(false)
    const language = i18next.language

    try {
      const res = await workflowAnonymousLoginApi({
        email: user.username + '@apo.com',
        language: language,
        remember_me: true,
      })

      if (res.result === 'success') {
        localStorage.setItem('difyToken', res.data.access_token)
        localStorage.setItem('difyRefreshToken', res.data.refresh_token)
        setDifyToken(res.data.access_token)
      } else {
        setError(true)
      }
    } catch (err) {
      console.error('Login error:', err)
      setError(true)
    } finally {
      setLoading(false)
    }
  }

  if (loading) {
    return <LoadingSpinner loading={loading} />
  }

  if (error) {
    return  <Result
    status="warning"
    title={t('oss/workflow:workflowError')}
  />
  }

  return difyToken ? (
    <iframe
      ref={workflowRef}
      src={'http://localhost:8083' + src}
      width="100%"
      height="100%"
      frameBorder={0}
    />
  ) : null
}

export default WorkflowsIframe
