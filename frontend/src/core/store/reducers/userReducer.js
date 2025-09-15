/**
 * Copyright 2024 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

export const initialState = {
  user: { username: 'anonymous', userId: '', role: '', roleList: '' },
  token: { accesstoken: null, refreshToken: null },
  menuItems: [],
  dataGroupList: [],
  routes: []
}

const userReducer = (state = initialState, action) => {
  switch (action.type) {
    case 'setUser':
      return { ...state, user: action.payload }
    case 'removeUser':
      return { user: 'anonymous', token: { accesstoken: null, refreshToken: null } }
    case 'setMenu':
      return { ...state, menuItems: action.payload }
    case 'setRoutes':
      return { ...state, routes: action.payload }
    case 'setDataGroupList':
      return { ...state, dataGroupList: action.payload }
    default:
      return state
  }
}

export default userReducer
