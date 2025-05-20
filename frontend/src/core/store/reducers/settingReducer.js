/**
 * Copyright 2024 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

const initialState = {
  sidebarShow: true,
  theme: 'light',
  monacoPromqlConfig: {},
  language: 'zh',
  reactJsonTheme: 'shapeshifter:inverted'
}

const settingReducer = (state = initialState, { type, payload, ...rest }) => {
  switch (type) {
    case 'set':
      return { ...state, ...rest }
    case 'setMonacoPromqlConfig':
      return { ...state, monacoPromqlConfig: payload }
    case 'setLanguage':
      return { ...state, language: payload }
    case 'setTheme':
      return { ...state, theme: payload, reactJsonTheme: payload === 'light' ? 'shapeshifter:inverted' : 'brewer' }
    default:
      return state
  }
}

export default settingReducer
