/**
 * Copyright 2024 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import React, { useState, useEffect } from 'react'
import { Button, Input, Form, Checkbox } from 'antd'
import { loginApi } from 'core/api/user'
import { useNavigate } from 'react-router-dom'
import { UserOutlined, LockOutlined } from '@ant-design/icons'
import { showToast } from 'core/utils/toast'
import logo from 'core/assets/brand/logo.svg'
import { AiOutlineLoading } from 'react-icons/ai'
import style from './Login.module.css'
import { useUserContext } from 'src/core/contexts/UserContext'

export default function Login() {
  const navigate = useNavigate()
  const [form] = Form.useForm()
  const [remeberMe, setRemeberMe] = useState(true)
  const [loading, setLoading] = useState(false)

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
            navigate('/')
            showToast({ title: '登录成功', color: 'success' })
            remeberMe
              ? localStorage.setItem('username', values.username)
              : localStorage.removeItem('username')
            localStorage.setItem('remeberMe', String(remeberMe))
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
    <div className={style.loginBackground}>
      <div className="w-3/12 bg-[rgba(0,0,0,0.4)] rounded-lg p-10 drop-shadow-xl flex flex-col">
        <div className="w-full flex justify-center items-center select-none">
          <img src={logo} className="w-12 mr-2" />
          <p className="text-2xl">向导式可观测平台</p>
        </div>
        <div className="w-full flex flex-col justify-center items-center mt-20">
          <Form form={form} className="w-full" layout="vertical">
            <Form.Item
              name="username"
              label={<label className="text-xs">用户名</label>}
              rules={[{ required: true, message: '请输入用户名' }]}
            >
              <Input
                size="large"
                className="w-full bg-[rgba(17,18,23,0.5)] hover:bg-[rgba(17,18,23,0.5)]"
                prefix={<UserOutlined />}
              />
            </Form.Item>

            <Form.Item
              name="password"
              label={<label className="text-xs">密码</label>}
              rules={[{ required: true, message: '请输入密码' }]}
            >
              <Input.Password
                size="large"
                className="w-full bg-[rgba(17,18,23,0.5)] hover:bg-[rgba(17,18,23,0.5)]"
                prefix={<LockOutlined />}
              />
            </Form.Item>
          </Form>
          <div className="w-full flex justify-between items-start mt-14">
            <Button
              size="large"
              disabled={loading}
              onClick={login}
              className="bg-[#455EEB] w-full border-none"
            >
              {loading ? <AiOutlineLoading className="animate-spin" /> : '登录'}
            </Button>
          </div>
        </div>
      </div>
    </div>
  )
}
