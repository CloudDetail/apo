/**
 * Copyright 2024 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import React, { Suspense } from 'react'
import { Navigate, Route, Routes } from 'react-router-dom'
import { CContainer } from '@coreui/react'

// routes config
import routes from 'src/routes'
import { useUserContext } from '../contexts/UserContext'
import { Spin } from 'antd'

const AppContent = () => {
  const { menuItems } = useUserContext()
  const getDefaultTo = () => {
    if (!menuItems || menuItems.length === 0) return '/'
    return menuItems[0]?.router?.to || menuItems[0]?.children?.[0]?.router?.to || '/'
  }
  return (
    <CContainer className="px-2" fluid>
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
                to={import.meta.env.VITE_APP_VERSION === 'pro' ? '/alert-analyze' : getDefaultTo()}
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
