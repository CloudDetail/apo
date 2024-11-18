import React, { useEffect, useRef, useState } from 'react'
import { NavLink, useLocation, useNavigate } from 'react-router-dom'
import { useSelector, useDispatch } from 'react-redux'
import logo from 'src/core/assets/brand/logo.svg'
import { CContainer, CHeader, CHeaderNav, useColorModes, CImage } from '@coreui/react'
import { AppBreadcrumb } from './index'
import { AppHeaderDropdown } from './header/index'
import DateTimeRangePicker from './DateTime/DateTimeRangePicker'
import routes from 'src/routes'
import CoachMask from './Mask/CoachMask'
import DateTimeCombine from './DateTime/DateTimeCombine'
import { Menu } from 'antd'
import { commercialNav } from 'src/_nav'
import ToolBox from './UserPage/component/ToolBox'
import { HiUserCircle } from "react-icons/hi";

// united / default
const AppHeader = ({ type = 'default' }) => {
  const location = useLocation()
  const navigate = useNavigate()
  const headerRef = useRef()
  const { colorMode, setColorMode } = useColorModes('coreui-free-react-admin-template-theme')
  const [toolVisibal, setToolVisibal] = useState(false)

  const dispatch = useDispatch()
  const sidebarShow = useSelector((state) => state.sidebarShow)
  const [selectedKeys, setSelectedKeys] = useState([])
  // 通过 ref 获取工具箱和触发按钮的 DOM 引用
  const toolBoxRef = useRef(null)
  const buttonRef = useRef(null)

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

  const checkRoute = () => {
    // 使用正则表达式替换动态参数（例如 :traceId）为通配符
    const currentRoute = routes.find((route) => {
      // 使用正则表达式替换动态参数（例如 :traceId）为通配符
      const routePattern = route.path.replace(/:\w+/g, '[^/]+') // 转换为 '/cause/report/[^/]+'
      const regex = new RegExp(`^${routePattern}$`) // 创建正则表达式
      return regex.test(location.pathname) // 使用正则测试
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

  useEffect(() => {
    const handleClickOutside = (event) => {
      if (toolBoxRef.current && !toolBoxRef.current.contains(event.target) &&
        buttonRef.current && !buttonRef.current.contains(event.target)) {
        setToolVisibal(false)
      }
    }

    // 监听点击事件
    document.addEventListener('click', handleClickOutside)

    // 清理事件监听器
    return () => {
      document.removeEventListener('click', handleClickOutside)
    }
  }, [])

  const vars = {
    // '--cui-header-bg': 'inherit',
    borderBottom: 0,
    zIndex: 998,
  }

  return (
    <CHeader position="sticky" className="mb-1 p-0" ref={headerRef} style={vars}>
      <div className="flex justify-between items-center w-full">
        {type === 'united' ? (
          <div className="flex items-center">
            <div className="h-[50px] flex overflow-hidden items-center mr-5">
              <CImage src={logo} className="w-[42px] sidebar-brand-narrow flex-shrink-0 mx-3" />
              <span className="flex-shrink-0 text-lg">向导式可观测平台</span>
            </div>
            <Menu
              mode="horizontal"
              theme="dark"
              items={commercialNav}
              onClick={onClick}
              selectedKeys={selectedKeys}
            ></Menu>
          </div>
        ) : (
          <CHeaderNav className="d-none d-md-flex  px-4 py-2 text-base flex items-center h-[50px] flex-grow">
            <AppBreadcrumb />
          </CHeaderNav>
        )}
        <CHeaderNav className="pr-4">
          {location.pathname === '/service/info' && <CoachMask />}
          {checkRoute() && <DateTimeCombine />}
        </CHeaderNav>
        <div className='relative h-8 w-8 flex justify-center items-center rounded-1 mr-5' ref={buttonRef} onClick={() => setToolVisibal(!toolVisibal)}>
          <HiUserCircle className='w-8 h-8' />
          <ToolBox visiable={toolVisibal} setVisiable={setToolVisibal} ref={toolBoxRef} />
        </div>
      </div>
    </CHeader>
  )
}

export default AppHeader
