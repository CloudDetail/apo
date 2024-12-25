/**
 * Copyright 2024 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

export const initialState = {
  user: { username: 'anonymous', userId: '' },
  token: { token: null, refreshToken: null },
  menuItems: [],
}

const userReducer = (state = initialState, action) => {
  switch (action.type) {
    case 'setUser':
      return { ...state, user: action.payload }
    case 'removeUser':
      return { user: 'anonymous', token: { token: null, refreshToken: null } }
    case 'setMenu':
      return { ...state, menuItems: action.payload }
    default:
      return state
  }
}

export default userReducer
