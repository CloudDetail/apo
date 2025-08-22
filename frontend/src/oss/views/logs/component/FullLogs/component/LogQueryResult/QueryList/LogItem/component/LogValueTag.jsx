/**
 * Copyright 2024 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import { Tag, Tooltip } from 'antd'
import React from 'react'
import { useSelector } from 'react-redux'
import LogTagDropDown from './LogTagDropdown'

// 内部Tag组件，能够接收isHighlighted属性
const TagWithHighlight = ({ objKey, value, isHighlighted = false }) => {
  const { theme } = useSelector((state) => state.settingReducer)
  
  // 根据主题定义高亮颜色
  const getHighlightStyle = () => {
    if (!isHighlighted) return {}
    
    if (theme === 'dark') {
      return {
        color: '#c3e88d',
        borderColor: '#c3e88d',
        backgroundColor: '#444444'
      }
    } else {
      // 亮色主题的高亮颜色 - 使用base16-light的字符串颜色
      return {
        color: '#f4bf75',
        borderColor: '#f4bf75',
        backgroundColor: '#faf8f0'
      }
    }
  }
  
  return (
    <Tooltip title={`${objKey} = "${value}"`} key={objKey}>
      <Tag
        className={`flex-shrink-0 inline-block max-w-[200px] overflow-hidden whitespace-nowrap text-ellipsis cursor-pointer`}
        style={getHighlightStyle()}
      >
        {value}
      </Tag>
    </Tooltip>
  )
}

// value作为tag内容
const LogValueTag = React.memo((props) => {
  const { objKey, value } = props

  return (
    <LogTagDropDown
      objKey={objKey}
      value={value}
      children={
        <TagWithHighlight objKey={objKey} value={value} />
      }
    />
  )
})
export default LogValueTag
