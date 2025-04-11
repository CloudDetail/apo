import React, { useEffect, useState } from 'react'
import { useLocation } from 'react-router-dom'
import { Spin } from 'antd'
import FallbackPage from 'src/core/components/FallbackPage'
import { getRouterPermissionApi } from '../api/permission'

const checkRoutePermission = async (route: string) => {
  const authResult = await getRouterPermissionApi({ router: route });
  return authResult
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