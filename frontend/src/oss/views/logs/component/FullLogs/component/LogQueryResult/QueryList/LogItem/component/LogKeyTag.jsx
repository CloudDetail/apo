/**
 * Copyright 2024 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import { Tag, theme } from 'antd'
import React from 'react'
import LogTagDropDown from './LogTagDropdown'
import LogKeyTagValue from './LogKeyTagValue'

// 内部组件，能够接收isHighlighted属性
const KeyTagValueWithHighlight = ({ title, description, isHighlighted = false }) => (
  <LogKeyTagValue title={title} description={description} isHighlighted={isHighlighted} />
)

const LogKeyTag = (props) => {
  const { title, description } = props

  const { useToken } = theme
  const { token } = useToken()

  // 判断 title 和 description 是否为对象或数组，若是则转换为字符串
  const formatValue = (value) => (typeof value === 'object' ? JSON.stringify(value) : value)
  return (
    <div className="flex">
      <div>
        <Tag className="cursor-pointer" style={{ color: token.colorText }}>{formatValue(title)} :</Tag>
      </div>
      {description ? (
        <div className="flex-1">
          <LogTagDropDown
            objKey={formatValue(title)}
            value={formatValue(description)}
            trigger={['contextMenu']}
            children={<KeyTagValueWithHighlight title={title} description={description} />}
          />
        </div>
      ) : null}
    </div>
  )
}

export default LogKeyTag
