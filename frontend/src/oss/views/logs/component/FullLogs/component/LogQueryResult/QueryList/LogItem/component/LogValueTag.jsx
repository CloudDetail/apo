/**
 * Copyright 2024 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import { Tag, theme, Tooltip } from 'antd'
import React from 'react'
import LogTagDropDown from './LogTagDropdown'
// value作为tag内容
const LogValueTag = React.memo((props) => {
  const { objKey, value } = props
  const { useToken } = theme
  const { token } = useToken()

  return (
    <LogTagDropDown
      objKey={objKey}
      value={value}
      children={
        <Tooltip title={`${objKey} = "${value}"`} key={objKey}>
          <Tag
            className="flex-shrink-0 inline-block max-w-[200px] overflow-hidden whitespace-nowrap text-ellipsis cursor-pointer"
            style={{ color: token.colorTextSecondary }}
          >
            {value}
          </Tag>
        </Tooltip>
      }
    />
  )
})
export default LogValueTag
