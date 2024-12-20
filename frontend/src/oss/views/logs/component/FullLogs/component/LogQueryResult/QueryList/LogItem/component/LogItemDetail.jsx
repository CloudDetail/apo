import { Tag } from 'antd'
import React, { useEffect, useState } from 'react'
import { useLogsContext } from 'src/core/contexts/LogsContext'
import LogKeyTag from './LogKeyTag'
const LogItemDetail = ({ log }) => {
  const [contentInfo, setContentInfo] = useState({})
  const { tableInfo, displayFields } = useLogsContext()
  useEffect(() => {
    try {
      const obj = JSON.parse(log.content)
      // 验证是否为对象且非空对象
      if (typeof obj === 'object' && obj !== null && !Array.isArray(obj)) {
        setContentInfo(obj)
      } else {
        throw new Error('解析的内容不是一个有效的对象')
      }
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
        <LogKeyTag key={key} title={key} description={value} />
      ))}
    </div>
  )
}
export default LogItemDetail
