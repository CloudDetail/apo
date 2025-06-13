/**
 * Copyright 2024 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import React, { useEffect, useLayoutEffect, useRef, useState, useCallback } from 'react'

// sidebar nav config
import { navIcon } from 'src/_nav'
import { ConfigProvider, Menu } from 'antd'
import { useLocation, useNavigate } from 'react-router-dom'
import { useUserContext } from '../contexts/UserContext'
import { useSelector } from 'react-redux'
const AppSidebarMenuIcon = (menuItem) => {
  return (
    <div className="appSidebarMenuIcon">
      <div>{navIcon[menuItem.key]}</div>
      <span className="text-xs ">
        {menuItem.abbreviation ? menuItem.abbreviation : menuItem.label}
      </span>
    </div>
  )
}
const AppSidebar = ({ collapsed }) => {
  const { menuItems, user } = useUserContext()
  const { theme } = useSelector((state) => state.settingReducer)

  const location = useLocation()
  const navigate = useNavigate()
  const [selectedKeys, setSelectedKeys] = useState([])
  const [openKeys, setOpenKeys] = useState([])
  const [memoOpenKeys, setMemoOpenKeys] = useState(['logs', 'trace', 'alerts'])
  const [menuList, setMenuList] = useState([])
  const siderRef = useRef(null)

  function prepareMenu(menu) {
    return {
      key: menu.key,
      label: menu.label,
      abbreviation: menu.abbreviation,
      icon: AppSidebarMenuIcon(menu),
      to: menu.router?.to,
      children: menu.children?.map((child) => prepareMenu(child)),
    }
  }

  useEffect(() => {
    const items = menuItems?.length ? menuItems.map(prepareMenu) : []
    setMenuList(items)
  }, [menuItems])

  const onClick = ({ item, key, keyPath, domEvent }) => {
    navigate(item.props.to)
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
  const onOpenChange = (openKeys) => {
    if (!collapsed) {
      setOpenKeys(openKeys)
      setMemoOpenKeys(openKeys)
    }
  }
  useEffect(() => {
    const result = getItemKey(menuList)
    setSelectedKeys(result)
  }, [location.pathname, menuList])
  useEffect(() => {
    if (!collapsed) {
      setOpenKeys(memoOpenKeys)
    } else {
      setOpenKeys([])
    }
  }, [collapsed])
  const [menuVisible, setMenuVisible] = useState(false)

  const scrollSelectedIntoView = useCallback(() => {
    const selectedItem = siderRef.current?.menu?.list?.querySelector('.ant-menu-item-selected')
    if (selectedItem) {
      selectedItem.scrollIntoView({ behavior: 'auto', block: 'center' })
      setMenuVisible(true)
    }
  }, [])

  useLayoutEffect(() => {
    if (collapsed) {
      setMenuVisible(false)
      return
    }

    // Use ResizeObserver to detect when menu finishes expanding
    const resizeObserver = new ResizeObserver((entries) => {
      // Only scroll when width increases (menu expanding)
      const entry = entries[0]
      if (entry.contentRect.width > 70) { // 70 is collapsedWidth
        scrollSelectedIntoView()
      }
    })

    const menuElement = siderRef.current?.menu?.list
    if (menuElement) {
      resizeObserver.observe(menuElement)
    }

    return () => resizeObserver.disconnect()
  }, [collapsed, scrollSelectedIntoView])
  return (
    <ConfigProvider
      theme={{
        components: {
          Menu: {
            itemHeight: 55,
            itemBg: 'var(--color-sider)',
            itemSelectedBg: 'var(--ant-color-primary)',
            itemSelectedColor: 'var(--menu-selected-text-color)',
            subMenuItemSelectedColor: 'var(--menu-selected-text-color)',
          },
        },
      }}
    >
      <Menu
        ref={siderRef}
        mode="inline"
        theme={theme}
        inlineCollapsed={collapsed}
        items={menuList}
        onClick={onClick}
        selectedKeys={selectedKeys}
        openKeys={openKeys}
        onOpenChange={onOpenChange}
        className={`sidebarMenu pb-20 ${!menuVisible && !collapsed ? 'opacity-0' : 'opacity-100'}`}
        style={{ transition: 'opacity 0.1s ease-in-out' }}
      ></Menu>
    </ConfigProvider>
  )
}

export default React.memo(AppSidebar)
