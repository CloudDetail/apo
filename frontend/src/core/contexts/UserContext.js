/**
 * Copyright 2024 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import React, { createContext, useContext, useEffect, useMemo, useReducer, useState } from 'react'
import userReducer, { initialState } from '../store/reducers/userReducer'
import { getUserPermissionApi } from '../api/permission'
import { useTranslation } from 'react-i18next'

const UserContext = createContext({})

export const useUserContext = () => useContext(UserContext)

export const UserProvider = ({ children }) => {
  const [state, dispatch] = useReducer(userReducer, initialState)
  const { i18n } = useTranslation()
  const { user, menuItems } = state

  const getUserPermission = () => {
    // getUserPermissionApi(state.user?.userId).then((res) => {
    getUserPermissionApi({ userId: user.userId, language: i18n.language }).then((res) => {
      dispatch({ type: 'setMenu', payload: res?.menuItem || [] })
    })
  }
  useEffect(() => {
    if (user.userId) getUserPermission()
  }, [user.userId, i18n.language])

  const value = {
    user: user,
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
