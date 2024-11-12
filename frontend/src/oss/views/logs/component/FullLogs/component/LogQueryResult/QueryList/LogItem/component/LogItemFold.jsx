import { Tag, Tooltip } from 'antd'
import React, { useEffect, useState } from 'react'
import LogTagDropDown from './LogTagDropdown'
import { useLogsContext } from 'src/core/contexts/LogsContext'
import LogValueTag from './LogValueTag'
const LogItemFold = ({ tags }) => {
  const { tableInfo } = useLogsContext()
  return (
    <div className="text-ellipsis text-wrap flex" style={{ display: '-webkit-box' }}>
      {Object.entries(tags).map(([key, value]) => {
        if (
          value !== '' && // 确保 value 存在且非空
          key !== (tableInfo?.timeField || 'timestamp') && // 排除与 timeField 相同的键
          typeof value !== 'object' // 确保 value 不是对象
        ) {
          return <LogValueTag key={key} objKey={key} value={value} />
        }
        return null // 不符合条件时返回 null
      })}
    </div>
  )
}
export default LogItemFold
