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
import logo from 'core/assets/brand/logo.svg'
import { AiOutlineLoading } from 'react-icons/ai'
import style from './Login.module.css'
import { useUserContext } from 'src/core/contexts/UserContext'
import { useTranslation } from 'react-i18next'
import { workflowLoginApi } from 'src/core/api/workflows'
import i18next from 'i18next'
import { notify } from 'src/core/utils/notify'
import { useSelector } from 'react-redux'

export default function Login() {
  const { user, dispatchUser } = useUserContext()
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
            window.localStorage.setItem('token', accessToken)
            window.localStorage.setItem('refreshToken', refreshToken)
            localStorage.removeItem('difyToken')
            localStorage.removeItem('difyRefreshToken')
            const redirectUrl = sessionStorage.getItem('urlBeforeLogin')

            notify({ message: t('index.loginSuccess'), type: 'success' })
            remeberMe
              ? localStorage.setItem('username', values.username)
              : localStorage.removeItem('username')
            localStorage.setItem('remeberMe', String(remeberMe))

            await loginDify(values)
            if (redirectUrl !== null) {
              window.location.href = redirectUrl
              sessionStorage.removeItem('urlBeforeLogin')
            } else {
              navigate('/')
            }
          }
        } catch (error) {
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
      window.localStorage.setItem('difyToken', res.data.access_token)
      window.localStorage.setItem('difyRefreshToken', res.data.refresh_token)
    }
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
          color: token.colorTextSecondary
        }}
      >
        <Flex className="w-full justify-center items-center select-none">
          <img src={logo} className="w-12 mr-2" />
          <p className="text-2xl">{t('index.title')}</p>
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
                style={{ backgroundColor: token.colorFillSecondary}}
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
                style={{ backgroundColor: token.colorFillSecondary}}
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
