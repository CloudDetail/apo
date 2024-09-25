const initialState = {
  sidebarShow: true,
  theme: 'light',
  monacoPromqlConfig: {},
}

const settingReducer = (state = initialState, { type, payload, ...rest }) => {
  switch (type) {
    case 'set':
      return { ...state, ...rest }
    case 'setMonacoPromqlConfig':
      return { ...state, monacoPromqlConfig: payload }
    default:
      return state
  }
}

export default settingReducer
