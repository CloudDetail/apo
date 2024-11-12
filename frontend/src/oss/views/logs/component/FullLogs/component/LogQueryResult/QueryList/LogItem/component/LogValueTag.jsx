import { Tag, Tooltip } from 'antd'
import React from 'react'
import LogTagDropDown from './LogTagDropdown'
// value作为tag内容
const LogValueTag = (props) => {
  const { objKey, value } = props
  return (
    <LogTagDropDown
      objKey={objKey}
      value={value}
      children={
        <Tooltip title={`${objKey} = "${value}"`} key={objKey}>
          <Tag className="flex-shrink-0 inline-block max-w-[200px] overflow-hidden whitespace-nowrap text-ellipsis cursor-pointer text-gray-400">
            {value}
          </Tag>
        </Tooltip>
      }
    />
  )
}
export default LogValueTag
