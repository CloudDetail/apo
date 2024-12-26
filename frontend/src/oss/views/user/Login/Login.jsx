import React, { useState, useEffect, useReducer } from 'react'
import { Button, Input, Form, Flex, Checkbox } from 'antd'
import { loginApi } from 'core/api/user'
import { useNavigate } from 'react-router-dom'
import { UserOutlined, LockOutlined } from '@ant-design/icons'
import { showToast } from 'core/utils/toast'
import { getUserInfoApi } from 'core/api/user'
import logo from 'core/assets/brand/logo.svg'
import { AiOutlineLoading } from 'react-icons/ai'
import style from './Login.module.css'
import userReducer, { initialState } from 'src/core/store/reducers/userReducer'
import { useUserContext } from 'src/core/contexts/UserContext'
import { useTranslation } from 'react-i18next'

export default function Login() {
  const { user, dispatchUser } = useUserContext()
  const navigate = useNavigate()
  const [form] = Form.useForm()
  const [remeberMe, setRemeberMe] = useState(true)
  const [loading, setLoading] = useState(false)
  const { t } = useTranslation('oss/login')

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
            showToast({ title: t('index.loginSuccess'), color: 'success' })
            remeberMe
              ? localStorage.setItem('username', values.username)
              : localStorage.removeItem('username')
            localStorage.setItem('remeberMe', String(remeberMe))
          }
          const user = await getUserInfoApi()
          // @ts-ignore
          dispatchUser({
            type: 'addUser',
            payload: JSON.stringify(user),
          })
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
    <Flex vertical className={style.loginBackground}>
      <Flex vertical className="w-3/12 bg-[rgba(0,0,0,0.4)] rounded-lg p-10 drop-shadow-xl">
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
                className="w-full bg-[rgba(17,18,23,0.5)] hover:bg-[rgba(17,18,23,0.5)]"
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
                className="w-full bg-[rgba(17,18,23,0.5)] hover:bg-[rgba(17,18,23,0.5)]"
                prefix={<LockOutlined />}
              />
            </Form.Item>
          </Form>
          <Flex className="w-full justify-between items-start mt-14">
            <Button
              size="large"
              disabled={loading}
              onClick={login}
              className="bg-[#455EEB] border-none w-full border-none"
            >
              {loading ? <AiOutlineLoading className="animate-spin" /> : t('index.login')}
            </Button>
          </Flex>
        </Flex>
      </Flex>
    </Flex>
  )
}