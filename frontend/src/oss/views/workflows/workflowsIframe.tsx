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
import { refreshDifyAccessToken, reloginDifyWithApoToken } from 'src/core/utils/request'

const WorkflowsIframe = ({ src }) => {
  const workflowRef = useRef(null)
  const { user } = useUserContext()
  const language = i18next.language
  const difyToken = useRef(localStorage.getItem('difyToken'))
  const refreshToken = useRef(localStorage.getItem('difyRefreshToken'))

  const [loading, setLoading] = useState(true)
  const [error, setError] = useState(false)
  const [iframeReloadKey, setIframeReloadKey] = useState(0)
  const retryCount = useRef(0)
  const initialized = useRef(false)
  const iframeRecoveryInProgress = useRef(false)

  const syncIframeTokens = useCallback((accessToken, difyRefreshToken) => {
    if (!accessToken || !difyRefreshToken)
      throw new Error('Dify token synchronization received incomplete credentials')
    localStorage.setItem('difyToken', accessToken)
    localStorage.setItem('difyRefreshToken', difyRefreshToken)
    difyToken.current = accessToken
    refreshToken.current = difyRefreshToken
    try {
      const iframeStorage = workflowRef.current?.contentWindow?.localStorage
      iframeStorage?.setItem('console_token', accessToken)
      iframeStorage?.setItem('refresh_token', difyRefreshToken)
    } catch (err) {
      console.warn('Unable to synchronize Dify iframe tokens:', err)
    }
  }, [])

  const loginDify = useCallback(async () => {
    setLoading(true)
    setError(false)

    try {
      if (user.username === 'anonymous') {
        const res = await workflowAnonymousLoginApi({
          email: `${user.username}@apo.com`,
          language,
          remember_me: true,
        })
        if (res.result !== 'success')
          throw new Error(res.message || 'Dify anonymous login failed')
        const { access_token, refresh_token } = res.data
        syncIframeTokens(access_token, refresh_token)
      }
      else {
        const { access_token, refresh_token } = await reloginDifyWithApoToken()
        syncIframeTokens(access_token, refresh_token)
      }
      retryCount.current = 0
      return true
    } catch (err) {
      console.error('Login error:', err)
      retryCount.current += 1
      setError(true)
      return false
    } finally {
      setLoading(false)
    }
  }, [user.username, language, syncIframeTokens])

  const refreshDifySession = useCallback(async (forceBlock = false, reloadIframe = false) => {
    const shouldBlock = forceBlock || !initialized.current
    let relogged = false
    if (shouldBlock)
      setLoading(true)
    setError(false)
    try {
      if (!difyToken.current || !refreshToken.current) {
        return await loginDify()
      }
      try {
        await refreshDifyAccessToken()
        syncIframeTokens(
          localStorage.getItem('difyToken'),
          localStorage.getItem('difyRefreshToken'),
        )
      } catch {
        const { access_token, refresh_token } = await reloginDifyWithApoToken()
        syncIframeTokens(access_token, refresh_token)
        relogged = true
      }
      retryCount.current = 0
      if ((relogged || reloadIframe) && initialized.current)
        setIframeReloadKey(key => key + 1)
      return true
    } catch (err) {
      console.error('Dify session recovery error:', err)
      retryCount.current += 1
      setError(true)
      return false
    } finally {
      initialized.current = true
      if (shouldBlock)
        setLoading(false)
    }
  }, [loginDify, syncIframeTokens])

  const recoverIframeLogin = useCallback(async () => {
    if (iframeRecoveryInProgress.current)
      return
    iframeRecoveryInProgress.current = true
    try {
      await refreshDifySession(true, true)
    } finally {
      iframeRecoveryInProgress.current = false
    }
  }, [refreshDifySession])

  const detectIframeLogin = useCallback(() => {
    try {
      if (workflowRef.current?.contentWindow?.location.pathname !== '/dify/signin')
        return
    } catch {
      return
    }
    // Dify may redirect only the iframe after its own refresh token expires.
    // Recover through APO and remount the original business page.
    recoverIframeLogin()
  }, [recoverIframeLogin])

  useEffect(() => {
    const handleMessage = (event) => {
      if (event.origin !== window.location.origin) {
        console.warn('CORS error: received message from unknown origin', event.origin)
        return
      }

      if (event.data?.type === 'refresh_token') {
        const { access_token, refresh_token } = event.data.data
        syncIframeTokens(access_token, refresh_token)
      }
    }

    window.addEventListener('message', handleMessage)
    return () => {
      window.removeEventListener('message', handleMessage)
    }
  }, [syncIframeTokens])

  useEffect(() => {
    refreshDifySession()
  }, [refreshDifySession])

  useEffect(() => {
    const timer = window.setInterval(refreshDifySession, 10 * 60 * 1000)
    const handleVisibilityChange = () => {
      if (document.visibilityState === 'visible')
        refreshDifySession(true)
    }
    document.addEventListener('visibilitychange', handleVisibilityChange)
    return () => {
      window.clearInterval(timer)
      document.removeEventListener('visibilitychange', handleVisibilityChange)
    }
  }, [refreshDifySession])

  useEffect(() => {
    const timer = window.setInterval(detectIframeLogin, 1000)
    return () => window.clearInterval(timer)
  }, [detectIframeLogin])

  useEffect(() => {
    if (!error)
      return
    const delay = Math.min(30000, 1000 * Math.pow(2, Math.min(retryCount.current, 5)))
    const timer = window.setTimeout(refreshDifySession, delay)
    return () => window.clearTimeout(timer)
  }, [error, refreshDifySession])

  if (loading) {
    return <LoadingSpinner loading={loading} />
  }

  if (error) {
    return <Result status="warning" title={t('oss/workflow:workflowError')} />
  }

  return (
    <iframe
      key={iframeReloadKey}
      ref={workflowRef}
      onLoad={detectIframeLogin}
      src={`${src}${src.includes('?') ? '&' : '?'}access_token=${difyToken.current}&refresh_token=${refreshToken.current}`}
      width="100%"
      height="100%"
      frameBorder={0}
    />
  )
}

export default WorkflowsIframe
