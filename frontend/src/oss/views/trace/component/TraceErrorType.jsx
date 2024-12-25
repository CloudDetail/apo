/**
 * Copyright 2024 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import React from 'react'

export default function TraceErrorType({ type }) {
  const ErrorTypeMap = {
    error: {
      name: '错误故障',
      color: 'rgba(220,38,38,.85)',
      background: '#ef444433',
      border: 'rgb(75,85,99)',
    },
    slow: {
      name: '慢故障',
      color: 'rgba(245,158,11,1)',
      background: '#f59e0b33',
      border: 'rgb(75,85,99)',
    },
    normal: {
      name: '无异常',
      color: 'rgba(52,211,153,1)',
      background: '#10b98133',
      border: 'rgb(75,85,99)',
    },
  }

  return (
    <>
      {/* <div
        className="px-2 rounded-xl py-1 text-xs mr-1"
        style={{
          background: ErrorTypeMap[type].background,
          border: ErrorTypeMap[type].border,
          color: ErrorTypeMap[type].color,
        }}
      >
        <span>{ErrorTypeMap[type].name}</span>
      </div> */}
      <div className="px-2 rounded-xl py-1 text-xs border text-[#9ca3af] border-[#9ca3af]">
        {' '}
        <span
          className="rounded-full w-2 h-2 inline-block mr-2"
          style={{
            background: ErrorTypeMap[type].color,
            color: ErrorTypeMap[type].border,
          }}
        ></span>
        <span>{ErrorTypeMap[type].name}</span>
      </div>
    </>
  )
}
