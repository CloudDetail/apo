/**
 * Copyright 2024 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import React, { createContext, useContext, useEffect, useMemo, useReducer, useState } from 'react'
import userReducer, { initialState } from '../store/reducers/userReducer'
import { getUserPermissionApi } from '../api/permission'

const UserContext = createContext({})

export const useUserContext = () => useContext(UserContext)

export const UserProvider = ({ children }) => {
  const [state, dispatch] = useReducer(userReducer, initialState)
  const { user, menuItems } = state

  const getUserPermission = () => {
    // getUserPermissionApi(state.user?.userId).then((res) => {
    getUserPermissionApi({ userId: user.userId }).then((res) => {
      dispatch({ type: 'setMenu', payload: res?.menuItem || [] })
    })
  }
  useEffect(() => {
    if (user.userId) getUserPermission()
  }, [user.userId])

  const value = {
    user: user,
    dispatchUser: dispatch,
    menuItems: menuItems,
    getUserPermission,
  }

  return (
    <>
      <UserContext.Provider value={value}>{children}</UserContext.Provider>
    </>
  )
}
