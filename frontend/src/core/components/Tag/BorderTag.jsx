/**
 * Copyright 2024 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import { theme } from 'antd'
import React from 'react'

export default function BorderTag({ type, name }) {
  const ErrorTypeMap = {
    //红 错误
    error: {
      color: 'rgba(220,38,38,.85)',
      background: '#ef444433',
      border: 'rgb(75,85,99)',
    },
    //黄 警告
    warning: {
      color: 'rgba(245,158,11,1)',
      background: '#f59e0b33',
      border: 'rgb(75,85,99)',
    },
    //绿 正确
    success: {
      color: 'rgba(52,211,153,1)',
      background: '#10b98133',
      border: 'rgb(75,85,99)',
    },
    //紫
    magenta: {
      color: '#EB2F96',
      background: '#10b98133',
      border: 'rgb(75,85,99)',
    },
  }
  const { useToken } = theme
  const { token } = useToken()
  return (
    <>
      <div className="px-2 rounded-xl py-1 text-xs border text-[#9ca3af] border-[#9ca3af]">
        {' '}
        <span
          className="rounded-full w-2 h-2 inline-block mr-2"
          style={{
            background: ErrorTypeMap[type].color,
            color: ErrorTypeMap[type].border,
          }}
        ></span>
        <span style={{ color: token.colorTextSecondary }}>{name}</span>
      </div>
    </>
  )
}
