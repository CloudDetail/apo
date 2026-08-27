/**
 * Copyright 2024 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import React, { createContext, useContext, useEffect } from 'react'
import { getUserPermissionApi } from '../api/permission'
import { useTranslation } from 'react-i18next'
import { useDispatch, useSelector } from 'react-redux'
import { replaceBuiltInDashboardMenuRoutes } from '../components/AppSidebarBehavior.mjs'
import {
  isSystemHealthPreviewRoute,
  normalizeSystemHealthMenu,
} from 'src/core/utils/systemHealthRoute'
import { isProbeManagementPreviewRoute } from 'src/core/utils/probeManagementRoute'

const UserContext = createContext({})

export const useUserContext = () => useContext(UserContext)

export const UserProvider = ({ children }) => {
  const dispatch = useDispatch()
  const state = useSelector((state) => state.userReducer)
  const { i18n } = useTranslation()
  const { user, menuItems, dataGroupList } = state

  const getUserPermission = () => {
    // getUserPermissionApi(state.user?.userId).then((res) => {
    return getUserPermissionApi({ userId: user.userId, language: i18n.language }).then((res) => {
      dispatch({
        type: 'setMenu',
        payload: normalizeSystemHealthMenu(replaceBuiltInDashboardMenuRoutes(res?.menuItem || [])),
      })
      dispatch({ type: 'setRoutes', payload: res?.routes || [] })
    })
  }

  useEffect(() => {
    if (user.userId && !isSystemHealthPreviewRoute() && !isProbeManagementPreviewRoute()) {
      getUserPermission()
      // getUserDataGroup()
    }
  }, [user.userId, i18n.language])

  const value = {
    user: user,
    dataGroupList: dataGroupList,
    dispatch: dispatch,
    menuItems: menuItems,
    getUserPermission,
  }

  return (
    <>
      <UserContext.Provider value={value}>{children}</UserContext.Provider>
    </>
  )
}
