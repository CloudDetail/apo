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
const AppSidebar = () => {
  const { menuItems, user } = useUserContext()
  const { theme } = useSelector((state) => state.settingReducer)

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
      popupClassName: `submenu-with-parent-${menu.key}`,
      className: `menu-item-${menu.key}`
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
    setOpenKeys(openKeys)
  }
  useEffect(() => {
    const result = getItemKey(menuList)
    setSelectedKeys(result)
  }, [location.pathname, menuList])

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
      <style>
        {menuList.map((item) => `
          .submenu-with-parent-${item.key} .ant-menu-sub::before {
            content: '${item.label}';
            display: block;
            margin: 4px;
            padding: 20px 40px 20px 25px;
            color: var(--ant-color-text-tertiary);
            font-size: 14px ;
            font-weight: 500;
            border-bottom: 1px solid var(--ant-color-border);
          }
          /* 覆盖 Tooltip 主体样式 */
          .ant-tooltip-inner {
            background-color: var(--ant-color-bg-elevated) !important;
            color: var(--ant-color-text) !important;
            box-shadow: 0 2px 8px rgba(0,0,0,0.15) !important;
            position: relative !important;
            left: -6px;
          }

          /* 隐藏 Tooltip 箭头 */
          .ant-tooltip-arrow,
          .ant-tooltip-arrow-content {
            display: none !important;
          }

          /* 可选：调整 padding / 圆角 */
          .ant-tooltip-inner {
            padding: 6px 10px !important;
            border-radius: 4px !important;
          }
        `).join('')}
      </style>
      <Menu
        mode="inline"
        theme={theme}
        inlineCollapsed={true}
        items={menuList}
        onClick={onClick}
        selectedKeys={selectedKeys}
        openKeys={openKeys}
        onOpenChange={onOpenChange}
        className="sidebarMenu *:custom-scrollbar"
      ></Menu>
    </ConfigProvider>
  )
}

export default React.memo(AppSidebar)
