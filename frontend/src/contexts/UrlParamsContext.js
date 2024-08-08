import React, { createContext, useContext, useReducer } from 'react'

// 初始化状态
const initialState = {
  startTime: null,
  endTime: null,
  service: '',
  instance: '',
  traceId: '',
  endpoint: '',
  filtersLoaded: false,
  instanceOption: {},
}

// 创建 Reducer 函数
function reducer(state, action) {
  switch (action.type) {
    case 'setStartTime':
      return { ...state, startTime: action.payload }
    case 'setEndTime':
      return { ...state, endTime: action.payload }
    case 'setService':
      return { ...state, service: action.payload }
    case 'setInstance':
      return { ...state, instance: action.payload }
    case 'setTraceId':
      return { ...state, traceId: action.payload }
    case 'setEndpoint':
      return { ...state, endpoint: action.payload }
    case 'setInstanceOption':
      return { ...state, instanceOption: action.payload }
    case 'setFiltersLoaded':
      return { ...state, filtersLoaded: action.payload }
    case 'setUrlParamsState':
      console.log(action.payload)
      return { ...state, ...action.payload }
    case 'clearUrlParamsState':
      return { ...initialState }
    default:
      throw new Error('Unknown action type')
  }
}

// 创建 Context
const UrlParamsContext = createContext({})

// 创建一个 Provider 组件
export function UrlParamsProvider({ children }) {
  const [urlParamsState, dispatch] = useReducer(reducer, initialState)
  return (
    <UrlParamsContext.Provider value={{ urlParamsState, dispatch }}>
      {children}
    </UrlParamsContext.Provider>
  )
}

// 创建一个自定义 Hook 方便使用 Context
export function useUrlParams() {
  const context = useContext(UrlParamsContext)
  if (context === undefined) {
    throw new Error('useInstance must be used within a InstanceProvider')
  }
  return context
}
