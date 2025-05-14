/**
 * Copyright 2024 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import React, { useEffect, useRef, useState } from 'react'
import { useLocation, useNavigate } from 'react-router-dom'
import { useSelector, useDispatch } from 'react-redux'
import logo from 'src/core/assets/brand/logo.svg'
import { CHeader, CHeaderNav, useColorModes, CImage } from '@coreui/react'
import { AppBreadcrumb } from './index'
import routes from 'src/routes'
import CoachMask from './Mask/CoachMask'
import DateTimeCombine from './DateTime/DateTimeCombine'
import { commercialNav } from 'src/_nav'
import UserToolBox from './UserToolBox'
import { t } from 'i18next'
import ThemeSwitcher from './ThemeSwitcher'
import { theme } from 'antd'

const AppHeader = ({ type = 'default' }) => {
  const location = useLocation()
  const navigate = useNavigate()
  const headerRef = useRef()
  const { colorMode, setColorMode } = useColorModes('coreui-free-react-admin-template-theme')
  const [toolVisibal, setToolVisibal] = useState(false)
  const [username, setUsername] = useState('')
  const dispatch = useDispatch()
  const sidebarShow = useSelector((state) => state.sidebarShow)
  const [selectedKeys, setSelectedKeys] = useState([])

  const onClick = (to) => {
    navigate(to)
  }

  const getItemKey = (navList) => {
    let result = []
    navList.forEach((item) => {
      if (location.pathname.startsWith(item.to)) {
        result.push(item.key)
      }
      if (item.children) {
        result = result.concat(getItemKey(item.children))
      }
    })
    return result
  }

  const checkRoute = () => {
    const currentRoute = routes.find((route) => {
      const routePattern = route.path.replace(/:\w+/g, '[^/]+')
      const regex = new RegExp(`^${routePattern}$`)
      return regex.test(location.pathname)
    })
    return !currentRoute?.hideSystemTimeRangePicker
  }

  useEffect(() => {
    const result = getItemKey(commercialNav)
    setSelectedKeys(result)
  }, [location.pathname])

  useEffect(() => {
    document.addEventListener('scroll', () => {
      headerRef.current &&
        headerRef.current.classList.toggle('shadow-sm', document.documentElement.scrollTop > 0)
    })
  }, [])

  const vars = {
    borderBottom: 0,
    zIndex: 998,
  }
  const { useToken } = theme
  const { token } = useToken()
  return (
    <CHeader position="sticky" className="mb-1 p-0" ref={headerRef} style={vars}>
      <div className="flex justify-between items-center w-full">
        {type === 'united' ? (
          <div className="flex items-center">
            <div className="h-[50px] flex overflow-hidden items-center mr-5">
              <CImage src={logo} className="w-[42px] sidebar-brand-narrow flex-shrink-0 mx-3" />
              <span className="flex-shrink-0 text-lg">{t('common:apoTitle')}</span>
            </div>
            {commercialNav.map((item) => (
              <div
                onClick={() => onClick(item.to)}
                className="h-[50px] items-center px-3 flex justify-center text-sm cursor-pointer"
                style={{
                  backgroundColor: selectedKeys.includes(item.key)
                    ? token.colorPrimary
                    : token.colorBgBase,
                  color: selectedKeys.includes(item.key)
                    ? 'var(--menu-selected-text-color)'
                    : token.colorText,
                }}
              >
                <span className="pr-2">{item.icon}</span> {item.label}
              </div>
            ))}
          </div>
        ) : (
          <CHeaderNav className="d-none d-md-flex px-4 py-2 text-base flex items-center h-[50px] flex-grow">
            <AppBreadcrumb />
          </CHeaderNav>
        )}
        <CHeaderNav className="pr-4 flex items-center">
          {location.pathname === '/service/info' && <CoachMask />}
          {checkRoute() && <DateTimeCombine />}
          <ThemeSwitcher />
          <UserToolBox />
        </CHeaderNav>
      </div>
    </CHeader>
  )
}

export default AppHeader
