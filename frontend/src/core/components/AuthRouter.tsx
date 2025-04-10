import React, { useEffect, useState } from 'react'
import { Navigate, useLocation } from 'react-router-dom'
import { Spin } from 'antd'
import FallbackPage from 'src/core/components/FallbackPage'

// Todo: 切换成真实的路由校验 API
const checkRoutePermission = (route: string) => {
  console.log("current route allowed: ", route)
  return true
}

const AuthRouter = (WrappedComponent) => {
  return function GuardedComponent(props) {
    const [isAllowed, setIsAllowed] = useState(null)
    const location = useLocation()

    useEffect(() => {
      const checkPermission = async () => {
        const hasPermission = await checkRoutePermission(location.pathname)
        setIsAllowed(hasPermission)
      }
      checkPermission()
    }, [location.pathname])

    if (isAllowed === null) {
      return (
        <div className="flex justify-center items-center h-screen">
          <Spin size="large" />
        </div>
      )
    }

    return isAllowed ? <WrappedComponent {...props} /> : <FallbackPage errorInfo='当前没有权限访问该路由' />
  }
}

export default AuthRouter