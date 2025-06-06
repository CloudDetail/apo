/**
 * Copyright 2024 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import { Flex, Popover, Button, Divider, Segmented } from 'antd'
import { LogoutOutlined, UserOutlined, SunOutlined, MoonOutlined } from '@ant-design/icons'
import { MdTune } from "react-icons/md";
import { useColorModes } from '@coreui/react'
import { VscColorMode } from "react-icons/vsc";
import { IoLanguageOutline } from "react-icons/io5";
import { useNavigate } from 'react-router-dom'
import { logoutApi, getUserInfoApi } from 'core/api/user'
import { HiUserCircle } from 'react-icons/hi'
import { useEffect } from 'react'
import { useUserContext } from '../contexts/UserContext'
import { useTranslation } from 'react-i18next'
import { notify } from '../utils/notify'
import i18next from 'i18next'
import { useDispatch, useSelector } from 'react-redux'

const UserToolBox = () => {
  const { user, dispatch: userDispatch } = useUserContext()
  const navigate = useNavigate()
  const { t, i18n } = useTranslation('core/userToolBox')

  const { theme } = useSelector((state) => state.settingReducer)
  const dispatch = useDispatch()

  const { setColorMode } = useColorModes('coreui-free-react-admin-template-theme')

  const toggleTheme = (value: 'light' | 'dark') => {
    setColorMode(value)
    dispatch({ type: 'setTheme', payload: value })
  }

  const toggleLanguage = (value: 'zh' | 'en') => {
    i18next
      .changeLanguage(value)
      .then(() => {
        dispatch({ type: 'setLanguage', payload: value })
      })
  }

  const content = (type: 'anonymous' | 'loggedIn') => (
    <>
      <Flex vertical className={'flex items-center w-36 rounded-lg z-50'}>
        <div className="w-full h-9 flex justify-center items-center gap-2">
          <VscColorMode className="text-base" title={t('colorMode')} />
          <Segmented
            defaultValue={theme}
            onChange={(value) => toggleTheme(value)}
            size="small"
            shape="round"
            options={[
              { value: 'light', icon: <SunOutlined />, className: 'w-8' },
              { value: 'dark', icon: <MoonOutlined />, className: 'w-8' },
            ]}
          />
        </div>
        <div className="w-full h-9 flex justify-center items-center gap-2">
          <IoLanguageOutline className="text-base" title={t('language')} />
          <Segmented
            defaultValue={i18n.language}
            onChange={(value) => toggleLanguage(value)}
            size="small"
            shape="round"
            options={[
              { value: 'zh', icon: 'ZH' },
              { value: 'en', icon: 'EN' },
            ]}
          />
        </div>
        { type === 'loggedIn' &&<>
        <Divider className='p-0 my-2' />
        <Flex
          vertical
          className="justify-center items-center w-full h-9 transition-colors hover:bg-[var(--ant-color-fill-tertiary)] active:bg-[var(--ant-color-fill-secondary)]"
          onClick={() => navigate('/user')}
        >
          <Flex className="w-2/3 justify-around p-2 cursor-pointer">
            <UserOutlined className="text-md" />
            <p className="text-md select-none my-2">{t('personalCenter')}</p>
          </Flex>
        </Flex>
        <Flex
          vertical
          className="justify-center items-center w-full h-9 transition-colors hover:bg-[var(--ant-color-fill-tertiary)] active:bg-[var(--ant-color-fill-secondary)]"
          onClick={logout}
        >
          <Flex className="w-2/3 justify-around p-2 cursor-pointer">
            <LogoutOutlined className="text-md" />
            <p className="text-md select-none my-2">{t('logout')}</p>
          </Flex>
        </Flex>
        </>}
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
      userDispatch({
        type: 'removeUser',
      })
      navigate('/login')
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
        // @ts-ignore
        userDispatch({
          type: 'setUser',
          payload: res,
        })
      })
      .catch((error) => {
        navigate('/login')
        console.error(error)
      })
  }

  useEffect(() => {
    getUserInfo()
  }, [])

  return (
    <>
      {user?.username !== 'anonymous' ? (
        <Popover content={content('loggedIn')}>
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
        <>
        <Popover content={content('anonymous')}>
          <Button type="text" icon={<MdTune />}></Button>
        </Popover>
        <Button type="link" onClick={() => navigate('/login')}>
          {t('login')}
        </Button>
        </>
      )}
    </>
  )
}

export default UserToolBox
