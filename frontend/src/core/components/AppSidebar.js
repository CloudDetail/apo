import React, { useEffect, useState } from 'react'

// sidebar nav config
import { _nav as navigation } from 'src/_nav'
import { ConfigProvider, Menu } from 'antd'
import { useLocation, useNavigate } from 'react-router-dom'

const AppSidebarMenuIcon = (menuItem) => {
  return (
    <div className="appSidebarMenuIcon">
      <div>{menuItem.icon}</div>
      <span className="text-xs ">
        {menuItem.abbreviation ? menuItem.abbreviation : menuItem.label}
      </span>
    </div>
  )
}
const AppSidebar = ({ collapsed }) => {
  const location = useLocation()
  const navigate = useNavigate()
  const [selectedKeys, setSelectedKeys] = useState([])
  const [openKeys, setOpenKeys] = useState([])
  const getItems = () => {
    return navigation.map((item) => ({ ...item, icon: AppSidebarMenuIcon(item) }))
  }
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
  useEffect(() => {
    const result = getItemKey(navigation)
    setSelectedKeys(result)
  }, [location.pathname])
  useEffect(() => {
    if (!collapsed) {
      setOpenKeys(['logs','manage'])
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
        inlineCollapsed={true}
        items={getItems()}
        onClick={onClick}
        selectedKeys={selectedKeys}
        openKeys={openKeys}
        className="pb-20"
      ></Menu>
    </ConfigProvider>
  )
}

export default React.memo(AppSidebar)
