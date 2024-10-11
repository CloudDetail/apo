import { Tag } from 'antd'
import React, { useEffect, useState } from 'react'
const LogItemDetail = ({ log }) => {
  return (
    <div className=" ">
      {Object.entries(log).map(([key, value]) => (
        <div className="flex">
          <div>
            <Tag className="cursor-pointer  text-gray-400">{key}</Tag>
          </div>
          ：{value}
        </div>
      ))}
    </div>
  )
}
export default LogItemDetail
