/**
 * Copyright 2024 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

export const initialState = {
  user: { username: 'anonymous', userId: '', role: '', roleList: '' },
  token: { accesstoken: null, refreshToken: null },
  menuItems: [],
}

const userReducer = (state = initialState, action) => {
  switch (action.type) {
    case 'setUser':
      return { ...state, user: action.payload }
    case 'removeUser':
      return { user: 'anonymous', token: { accesstoken: null, refreshToken: null } }
    case 'setToken':
      console.log(action)
      return { ...state, token: action.payload }
    case 'removeToken':
      return { ...state, token: { accesstoken: null, refreshToken: null } }
    case 'setMenu':
      return { ...state, menuItems: action.payload }
    default:
      return state
  }
}

export default userReducer
