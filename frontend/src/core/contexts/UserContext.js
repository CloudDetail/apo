/**
 * Copyright 2024 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import React, { createContext, useContext, useEffect } from 'react'
import { getUserPermissionApi } from '../api/permission'
import { useTranslation } from 'react-i18next'
import { useDispatch, useSelector } from 'react-redux'

const UserContext = createContext({})

export const useUserContext = () => useContext(UserContext)

export const UserProvider = ({ children }) => {
  const dispatch = useDispatch()
  const state = useSelector((state) => state.userReducer)
  const { i18n } = useTranslation()
  const { user, menuItems, dataGroupList } = state

  const getUserPermission = () => {
    // getUserPermissionApi(state.user?.userId).then((res) => {
    getUserPermissionApi({ userId: user.userId, language: i18n.language }).then((res) => {
      dispatch({ type: 'setMenu', payload: res?.menuItem || [] })
    })
  }

  useEffect(() => {
    if (user.userId) {
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
