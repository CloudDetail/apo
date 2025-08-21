/**
 * Copyright 2025 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import React, { useEffect, useState } from 'react'
import { useLocation } from 'react-router-dom'
import { Spin } from 'antd'
import FallbackPage from 'src/core/components/FallbackPage'
import { getRouterPermissionApi } from '../api/permission'
import { useTranslation } from 'react-i18next'

const checkRoutePermission = async (route: string) => {
  const authResult = await getRouterPermissionApi({ router: route });
  return authResult || route === '/user'
}

const AuthRouter = (WrappedComponent) => {
  return function GuardedComponent(props) {
    const [isAllowed, setIsAllowed] = useState(null)
    const location = useLocation()
    const { t } = useTranslation('common')

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

    // return isAllowed ? <WrappedComponent {...props} /> : <FallbackPage errorInfo={t('routeError')} />
    return <WrappedComponent {...props} />
  }
}

export default AuthRouter