/**
 * Copyright 2024 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import { theme } from 'antd'
import React from 'react'

export default function Tag(props) {
  const { useToken } = theme
  const { token } = useToken()

  const { type = 'default', color, children } = props
  const typeColorMap = {
    // default: { color: '#a1a1a1', borderColor: '#fafafa' },
    // success: { color: '#6abe39', borderColor: '#274916', backgroundColor: '#162312' },
    // error: { color: '#e84749', borderColor: '#58181c', backgroundColor: '#2a1215' },
    // warning: { color: '#e89a3c', background: '#2b1d11', borderColor: '#593815' },
    default: { color: token.colorTextSecondary, borderColor: token.colorBorderSecondary },
    success: { color: token.colorSuccess, borderColor: token.colorSuccessBorder, backgroundColor: token.colorSuccessBg },
    error: { color: token.colorError, borderColor: token.colorErrorBorder, backgroundColor: token.colorErrorBg },
    warning: { color: token.colorWarning, background: token.colorWarningBg, borderColor: token.colorWarningBorder },
    // primary: {
    //   color: '#3c89e8',
    //   backgroundColor: '#111a2c',
    //   borderColor: '#15325b',
    // },
    primary: {
      color: token.colorPrimary,
      backgroundColor: token.colorPrimaryBg,
      borderColor: token.colorPrimaryBorder,
    },
  }
  return (
    <span
      style={{
        ...typeColorMap[type],
        border: '1px solid',
        borderColor: color ?? typeColorMap[type],
        padding: '2px 5px',
        borderRadius: 2,
        fontSize: 10,
      }}
    >
      {children}
    </span>
  )
}
