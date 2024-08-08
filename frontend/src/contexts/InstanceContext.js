// contexts/InstanceContext.js

import React, { createContext, useContext, useReducer } from 'react'

// 初始化状态
const initialState = { instanceOption: {} }

// 创建 Reducer 函数
function reducer(state, action) {
  switch (action.type) {
    case 'setInstanceOption':
      console.log(action.payload)
      return { ...state, instanceOption: action.payload }
    default:
      throw new Error('Unknown action type')
  }
}

// 创建 Context
const InstanceContext = createContext({})

// 创建一个 Provider 组件
export function InstanceProvider({ children }) {
  const [instanceState, dispatch] = useReducer(reducer, initialState)
  return (
    <InstanceContext.Provider value={{ instanceState, dispatch }}>
      {children}
    </InstanceContext.Provider>
  )
}

// 创建一个自定义 Hook 方便使用 Context
export function useInstance() {
  const context = useContext(InstanceContext)
  if (context === undefined) {
    throw new Error('useInstance must be used within a InstanceProvider')
  }
  return context
}
