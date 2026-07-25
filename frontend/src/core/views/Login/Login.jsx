/**
 * Copyright 2024 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import React, { useState, useEffect } from 'react'
import { Button, Input, Form, Flex, theme as antdTheme } from 'antd'
import bgLight from 'src/core/assets/brand/bg-light.png'
import bg from 'src/core/assets/brand/bg.jpg'
import { loginApi } from 'core/api/user'
import { useNavigate } from 'react-router-dom'
import { UserOutlined, LockOutlined } from '@ant-design/icons'
import { AiOutlineLoading } from 'react-icons/ai'
import style from './Login.module.css'
import { useTranslation } from 'react-i18next'
import { workflowLoginApi } from 'src/core/api/workflows'
import i18next from 'i18next'
import { notify } from 'src/core/utils/notify'
import { useSelector } from 'react-redux'
import { useDeployMode } from 'src/core/contexts/DeployModeContext'
import { clearAuthSession, clearDifySession, reloginDifyWithApoToken } from 'src/core/utils/request'

export default function Login() {
  const { getLogo, getTitle, deployMode } = useDeployMode()
  const navigate = useNavigate()
  const [form] = Form.useForm()
  const [remeberMe, setRemeberMe] = useState(true)
  const [loading, setLoading] = useState(false)
  const { t } = useTranslation('core/login')
  const language = i18next.language
  const { theme } = useSelector((state) => state.settingReducer)
  const { useToken } = antdTheme
  const { token } = useToken()
  const login = () => {
    if (loading) return
    form
      .validateFields()
      .then(async (values) => {
        try {
          setLoading(true)
          const { accessToken, refreshToken } = await loginApi(values)
          if (accessToken && refreshToken) {
            // Store the new APO credentials before calling Dify so the request
            // interceptor cannot reuse a token from an earlier session.
            clearAuthSession()
            window.localStorage.setItem('token', accessToken)
            window.localStorage.setItem('refreshToken', refreshToken)

            // Establish the separate Dify iframe session after APO login.
            if (deployMode !== 'syncause') {
              try {
                await loginDify(values)
              } catch (error) {
                clearDifySession()
                console.error('Dify password login failed; retrying with the APO session:', error)
                try {
                  await reloginDifyWithApoToken()
                } catch (reloginError) {
                  // APO remains usable. Dify-backed views will retry the same
                  // token exchange when the user opens them.
                  console.error('Dify automatic re-login temporarily unavailable:', reloginError)
                }
              }
            }

            remeberMe
              ? localStorage.setItem('username', values.username)
              : localStorage.removeItem('username')
            localStorage.setItem('remeberMe', String(remeberMe))
            navigate('/')
            notify({ message: t('index.loginSuccess'), type: 'success' })
          }
        } catch (error) {
          clearAuthSession()
          console.error(error)
        } finally {
          setLoading(false)
        }
      })
      .catch((errorInfo) => {
        console.log('验证失败:', errorInfo)
      })
  }

  const loginDify = async (values) => {
    const res = await workflowLoginApi({
      email: values.username + '@apo.com',
      password: values.password,
      language: language,
      remember_me: true,
    })
    if (res.result === 'success') {
      const { access_token, refresh_token } = res.data || {}
      if (!access_token || !refresh_token) {
        throw new Error('Dify login returned incomplete credentials')
      }
      window.localStorage.setItem('difyToken', access_token)
      window.localStorage.setItem('difyRefreshToken', refresh_token)
      return
    }
    throw new Error(res.message || 'Dify login failed')
  }
  useEffect(() => {
    form.setFieldValue('username', localStorage.getItem('username'))
    const savedRememberMe = localStorage.getItem('remeberMe')
    setRemeberMe(savedRememberMe ? JSON.parse(savedRememberMe) : false)

    const handleKeyDown = (e) => {
      if (e.key === 'Enter') {
        login()
      }
    }

    // 添加全局键盘事件监听器
    window.addEventListener('keydown', handleKeyDown)

    // 在组件卸载时移除事件监听器
    return () => {
      window.removeEventListener('keydown', handleKeyDown)
    }
  }, [])

  return (
    <Flex
      vertical
      className={style.loginBackground}
      style={{ backgroundImage: theme === 'light' ? `url(${bgLight})` : `url(${bg})` }}
    >
      <Flex
        vertical
        className="w-3/12 rounded-lg p-10 drop-shadow-xl"
        style={{
          backgroundColor: theme === 'light' ? token.colorFillSecondary : token.colorBgMask,
          color: token.colorTextSecondary,
        }}
      >
        <Flex className="w-full justify-center items-center select-none">
          <img src={getLogo()} className="w-12 mr-2" />
          <p className="text-2xl">{getTitle()}</p>
        </Flex>
        <Flex vertical className="w-full justify-center items-center mt-20">
          <Form form={form} className="w-full">
            <label className="text-xs">{t('index.username')}</label>
            <Form.Item
              name="username"
              rules={[{ required: true, message: t('index.enterUsername') }]}
            >
              <Input
                size="large"
                className="w-full"
                style={{ backgroundColor: token.colorFillSecondary }}
                prefix={<UserOutlined />}
              />
            </Form.Item>
            <label className="text-xs">{t('index.password')}</label>
            <Form.Item
              name="password"
              rules={[{ required: true, message: t('index.enterPassword') }]}
            >
              <Input.Password
                size="large"
                className="w-full"
                style={{ backgroundColor: token.colorFillSecondary }}
                prefix={<LockOutlined />}
              />
            </Form.Item>
          </Form>
          <Flex className="w-full justify-between items-start mt-14">
            <Button
              type="primary"
              size="large"
              disabled={loading}
              onClick={login}
              className="w-full border-none"
            >
              {loading ? <AiOutlineLoading className="animate-spin" /> : t('index.login')}
            </Button>
          </Flex>
        </Flex>
      </Flex>
    </Flex>
  )
}
