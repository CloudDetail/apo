import { Tag } from 'antd'
import React from 'react'

const LogTag = (props) => {
  const { title, description } = props
  return (
    <div className="flex">
      <div>
        <Tag className="cursor-pointer  text-gray-200">{title}</Tag>
      </div>
      ：<span className=" break-words">{description}</span>
    </div>
  )
}
export default LogTag
