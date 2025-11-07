/**
 * Copyright 2024 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import React, { useEffect, useState } from 'react'

// sidebar nav config
import { navIcon } from 'src/_nav'
import { ConfigProvider, Divider, Menu } from 'antd'
import { useLocation, useNavigate } from 'react-router-dom'
import { useUserContext } from '../contexts/UserContext'
import { useSelector } from 'react-redux'
import styles from './AppSidebar.module.scss'

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

const agentKeys = import.meta.env.VITE_APP_CODE_VERSION === 'EE' ? ['agent', 'alerts', 'probe', 'integration'] : []

const AppSidebar = () => {
  const { menuItems, user } = useUserContext()
  const { theme } = useSelector((state) => state.settingReducer)

  const location = useLocation()
  const navigate = useNavigate()
  const [selectedKeys, setSelectedKeys] = useState([])
  const [openKeys, setOpenKeys] = useState([])
  const [menuList, setMenuList] = useState([])
  function prepareGroup(menu) {
    return [
      {
        type: 'group',
        key: menu.key + 'group',
        label: (
          <span
            className={`text-[var(${menu.children?.length > 0 ? '--ant-color-text-secondary' : '--ant-color-text'})]`}
          >
            {menu.label}
          </span>
        ),
        children: menu.children?.map((child) => ({
          key: child.key,
          label: <span className="text-[var(--ant-color-text)]">{child.label}</span>,
          to: child.router?.to,
        })),
      },
    ]
  }
  function prepareMenu(menu) {
    return {
      key: menu.key,
      label: menu.label,
      abbreviation: menu.abbreviation,
      icon: AppSidebarMenuIcon(menu),
      to: menu.router?.to,
      children: menu.children?.length > 0 && prepareGroup(menu),
      popupClassName: `submenu-with-parent-${menu.key}`,
      className: `menu-item-${menu.key}`,
    }
  }

  useEffect(() => {
    const items = menuItems?.length
      ? menuItems.filter((item) => !agentKeys.includes(item.key)).map(prepareMenu)
      : []
    setMenuList(items)
  }, [menuItems])

  const onClick = ({ item, key, keyPath, domEvent }) => {
    navigate(item.props.to)
  }
  const getItemKey = (navList) => {
    if (!navList) return []
    let result = []
    navList?.forEach((item) => {
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
    const agentResult = getItemKey(agentMenuList())
    setSelectedKeys([...result, ...agentResult])
  }, [location.pathname, menuList])
  const agentMenuList = () => {
    return agentKeys
      .map((key) => {
        const item = menuItems?.find((item) => item.key === key)
        return item ? prepareMenu(item) : null
      })
      .filter(Boolean)
  }

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
      <div className='h-full flex flex-col justify-between ' >
        <Menu
          mode="inline"
          theme={theme}
          inlineCollapsed={true}
          items={menuList}
          onClick={onClick}
          selectedKeys={selectedKeys}
          openKeys={openKeys}
          onOpenChange={onOpenChange}
          className={styles.sidebarMenu}
        ></Menu>

      </div>
    </ConfigProvider>
  )
}

export default React.memo(AppSidebar)
