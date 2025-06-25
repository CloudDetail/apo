/**
 * Copyright 2024 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import { Flex, Popover, theme } from 'antd'
import { LogoutOutlined, UserOutlined } from '@ant-design/icons'
import { useNavigate } from 'react-router-dom'
import { logoutApi, getUserInfoApi } from 'core/api/user'
import { HiUserCircle } from 'react-icons/hi'
import { useEffect } from 'react'
import { useUserContext } from '../contexts/UserContext'
import { useTranslation } from 'react-i18next'
import { notify } from '../utils/notify'
import { redirectToLogin } from '../utils/redirectToLogin'

const UserToolBox = () => {
  const { user, dispatch } = useUserContext()
  const navigate = useNavigate()
  const { t } = useTranslation('core/userToolBox')

  const { useToken } = theme
  const { token } = useToken()

  const content = (
    <>
      <Flex vertical className={'flex items-center w-36 rounded-lg z-50'}>
        <Flex
          vertical
          className="justify-center items-center w-full h-9 transition-colors"
          onClick={() => navigate('/user')}
          onMouseEnter={(e) => {
            e.currentTarget.style.backgroundColor = token.colorFillSecondary
          }}
          onMouseOver={(e) => {
            e.currentTarget.style.backgroundColor = token.colorFillTertiary
          }}
          onMouseLeave={(e) => {
            e.currentTarget.style.backgroundColor = 'var(--ant-layout-color-bg-body)'
          }}
        >
          <Flex className="w-2/3 justify-around p-2 cursor-pointer">
            <UserOutlined className="text-md" />
            <p className="text-md select-none my-2">{t('personalCenter')}</p>
          </Flex>
        </Flex>
        <Flex
          vertical
          className="justify-center items-center w-full h-9 mt-2 transition-colors"
          onClick={logout}
          onMouseEnter={(e) => {
            e.currentTarget.style.backgroundColor = token.colorFillSecondary
          }}
          onMouseOver={(e) => {
            e.currentTarget.style.backgroundColor = token.colorFillTertiary
          }}
          onMouseLeave={(e) => {
            e.currentTarget.style.backgroundColor = 'var(--ant-layout-color-bg-body)'
          }}
        >
          <Flex className="w-2/3 justify-around p-2 cursor-pointer">
            <LogoutOutlined className="text-md" />
            <p className="text-md select-none my-2">{t('logout')}</p>
          </Flex>
        </Flex>
      </Flex>
    </>
  )

  //退出登录
  async function logout() {
    try {
      const params = {
        accessToken: localStorage.getItem('token'),
        refreshToken: localStorage.getItem('refreshToken'),
      }
      await logoutApi(params)
      localStorage.removeItem('token')
      localStorage.removeItem('refreshToken')
      localStorage.removeItem('difyToken')
      localStorage.removeItem('difyRefreshToken')
      // @ts-ignore
      dispatch({
        type: 'removeUser',
      })
      redirectToLogin(false)
      notify({
        message: t('logoutSuccess'),
        type: 'success',
      })
    } catch (error) {
      console.error(error)
    }
  }

  function getUserInfo() {
    getUserInfoApi()
      .then((res) => {
        console.log('res', res)
        // @ts-ignore
        dispatch({
          type: 'setUser',
          payload: res,
        })
      })
      .catch((error) => {
        redirectToLogin(true)
        console.error(error)
      })
  }

  useEffect(() => {
    getUserInfo()
  }, [])

  return (
    <>
      {user?.username !== 'anonymous' ? (
        <Popover content={content}>
          <div className="relative flex items-center select-none w-auto pl-2 pr-2 rounded-md cursor-pointer">
            <div>
              <HiUserCircle className="w-8 h-8" />
            </div>
            <div className="h-1/2 flex flex-col justify-center">
              <p className="text-base relative -top-0.5 m-2">{user?.username}</p>
            </div>
          </div>
        </Popover>
      ) : (
        <></>
        // <Button type="link" onClick={() => redirectToLogin(true)}>
        //   {t('login')}
        // </Button>
      )}
    </>
  )
}

export default UserToolBox
