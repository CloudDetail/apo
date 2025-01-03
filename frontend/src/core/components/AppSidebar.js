/**
 * Copyright 2024 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import React, { useEffect, useState } from 'react'

// sidebar nav config
import { navIcon } from 'src/_nav'
import { ConfigProvider, Menu } from 'antd'
import { useLocation, useNavigate } from 'react-router-dom'
import { useUserContext } from '../contexts/UserContext'
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
  const location = useLocation()
  const navigate = useNavigate()
  const [selectedKeys, setSelectedKeys] = useState([])
  const [openKeys, setOpenKeys] = useState([])
  const [memoOpenKeys, setMemoOpenKeys] = useState(['logs', 'trace', 'alerts'])
  const [menuList, setMenuList] = useState([])

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
  return (
    <ConfigProvider
      theme={{
        components: {
          Menu: {
            itemHeight: 55,
            darkItemBg: '#1d222b',
          },
        },
      }}
    >
      <Menu
        mode="inline"
        theme="dark"
        inlineCollapsed={collapsed}
        items={menuList}
        onClick={onClick}
        selectedKeys={selectedKeys}
        openKeys={openKeys}
        onOpenChange={onOpenChange}
        className="sidebarMenu pb-20"
      ></Menu>
    </ConfigProvider>
  )
}

export default React.memo(AppSidebar)
