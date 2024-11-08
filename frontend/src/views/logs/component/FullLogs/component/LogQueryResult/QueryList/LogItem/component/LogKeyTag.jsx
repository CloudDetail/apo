import { Tag } from 'antd'
import React from 'react'
import LogTagDropDown from './LogTagDropdown'

const LogKeyTag = (props) => {
  const { title, description } = props

  // 判断 title 和 description 是否为对象或数组，若是则转换为字符串
  const formatValue = (value) => (typeof value === 'object' ? JSON.stringify(value) : value)

  return (
    <div className="flex">
      <div>
        <Tag className="cursor-pointer text-gray-200">{formatValue(title)}</Tag>
      </div>
      ：
      <LogTagDropDown
        objKey={formatValue(title)}
        value={formatValue(description)}
        children={
          <span className="break-all hover:underline cursor-pointer">
            {formatValue(description)}
          </span>
        }
      />
    </div>
  )
}

export default LogKeyTag
