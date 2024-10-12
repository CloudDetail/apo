import { Tag } from 'antd'
import React, { useEffect, useState } from 'react'
import { useLogsContext } from 'src/contexts/LogsContext'
const LogItemDetail = ({ log }) => {
  const [contentInfo, setContentInfo] = useState({})
  useEffect(() => {
    try {
      const obj = JSON.parse(log.content)
      setContentInfo(obj)
    } catch (error) {
      // console.error('JSON 解析失败:', error)
      setContentInfo({
        content: log.content,
      })
    }
  }, [log])
  return (
    <div className=" ">
      {Object.entries(contentInfo).map(([key, value]) => (
        <div className="flex">
          <div>
            <Tag className="cursor-pointer  text-gray-200">{key}</Tag>
          </div>
          ：{value}
        </div>
      ))}
    </div>
  )
}
export default LogItemDetail
