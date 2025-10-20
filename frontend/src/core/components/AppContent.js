/**
 * Copyright 2024 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import React, { Suspense } from 'react'
import { Navigate, Route, Routes } from 'react-router-dom'
import { CContainer } from '@coreui/react'

// routes config
import { useUserContext } from '../contexts/UserContext'
import { Spin } from 'antd'
import routes from 'src/routes'
const AppContent = () => {
  const { menuItems, routes: storeRoutes } = useUserContext()
  const getDefaultTo = () => {
    if (!menuItems || menuItems.length === 0) return '/'
    const redirectUrl = sessionStorage.getItem('urlBeforeLogin')

    if (redirectUrl !== null && storeRoutes?.includes(redirectUrl)) {
      window.location.href = redirectUrl
    }
    sessionStorage.removeItem('urlBeforeLogin')

    return menuItems[0]?.router?.to || menuItems[0]?.children?.[0]?.router?.to || '/'
  }
  return (
    <CContainer className="p-2 flex-1 h-0 overflow-auto" fluid>
      <Suspense fallback={<Spin size={'large'} />}>
        <Routes>
          {routes.map((route, idx) => {
            return (
              route.element && (
                <Route
                  key={idx}
                  path={route.path}
                  exact={route.exact}
                  name={route.name}
                  element={<route.element />}
                />
              )
            )
          })}
          <Route
            path="/"
            element={
              <Navigate
                to={getDefaultTo()}
                replace
              />
            }
          />
        </Routes>
      </Suspense>
    </CContainer>
  )
}

export default React.memo(AppContent)
