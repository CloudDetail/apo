import React, { useEffect, useState, useMemo } from 'react'
import { useLogsContext } from 'src/core/contexts/LogsContext'
import LogKeyTag from './LogKeyTag'
const LogItemDetail = ({ log, contentVisibility }) => {
  const { content, logFields: fields } = log
  const [contentInfo, setContentInfo] = useState({})
  const { tableInfo, displayFields } = useLogsContext()

  //由tableName和type组成的唯一标识
  const tableId = `${tableInfo.tableName}_${tableInfo.type}`

  // 计算过滤后的 fields
  const filteredFields = useMemo(() => {
    return fields ? Object.entries(fields).filter(([key, value]) => displayFields[tableId]?.includes(key)) : [];
  }, [fields, displayFields]);

  useEffect(() => {
    try {
      const obj = JSON.parse(content)
      // 验证是否为对象且非空对象
      if (typeof obj === 'object' && obj !== null && !Array.isArray(obj)) {
        setContentInfo(obj)
      } else {
        throw new Error('解析的内容不是一个有效的对象')
      }
    } catch (error) {
      // console.error('JSON 解析失败:', error)
      setContentInfo({
        content: content,
      })
    }
  }, [log])
  return (<>
    {/* 渲染 fields */}
    <div className="text-ellipsis text-wrap flex flex-col overflow-hidden">
      {filteredFields.map(([key, value]) => (
        <LogKeyTag key={key} title={key} description={value} />
      ))}
    </div>
    <div>
      {contentVisibility && Object.entries(contentInfo).map(([key, value]) => (
        <LogKeyTag key={key} title={key} description={value} />
      ))}
    </div>
  </>)
}
export default LogItemDetail
