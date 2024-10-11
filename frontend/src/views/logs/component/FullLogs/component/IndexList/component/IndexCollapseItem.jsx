import { Collapse, List, Progress, Tag, Tooltip } from 'antd'
import React, { useEffect, useState } from 'react'
import { useSearchParams } from 'react-router-dom'
import { useLogsContext } from 'src/contexts/LogsContext'
import { ISOToTimestamp } from 'src/utils/time'

const IndexCollapseItem = ({ field }) => {
  const { query, updateQuery, fieldIndexMap, getFieldIndexData } = useLogsContext()

  const [searchParams] = useSearchParams()
  const startTime = ISOToTimestamp(searchParams.get('log-from'))
  const endTime = ISOToTimestamp(searchParams.get('log-to'))
  const [loading, setLoading] = useState(false)
  const clickIndex = (index) => {
    let newQueryPart = '`' + field + '` =' + "'" + index.indexName + "'"
    // 检查 query 是否已经包含 newQueryPart
    if (!query.includes(newQueryPart)) {
      let newQuery = query
      if (newQuery.length > 0) {
        newQuery += ' And '
      }
      newQuery += newQueryPart
      updateQuery(newQuery) // 更新查询
    }
  }
  useEffect(() => {
    if (!fieldIndexMap[field]) {
      setLoading(true)

      getFieldIndexData({
        startTime,
        endTime,
        column: field,
      }).finally(() => {
        setLoading(false)
      })
    }
  }, [field, fieldIndexMap])
  return (
    <List
      loading={loading}
      dataSource={fieldIndexMap[field]}
      bordered={false}
      className="h-full overflow-y-auto w-full pl-5"
      renderItem={(item) => (
        <List.Item key={item} onClick={() => clickIndex(item)} className=" cursor-pointer">
          <div>
            <Tooltip title={item.indexName}>
              <Tag className="max-w-[120px] overflow-hidden whitespace-nowrap text-ellipsis cursor-pointer text-gray-400">
                {item.indexName}
              </Tag>
            </Tooltip>
            <Progress
              percent={parseFloat(item.percent.toFixed(2))}
              size={{ width: 160 }}
              status="normal"
            />
          </div>
        </List.Item>
      )}
    />
  )
}
export default IndexCollapseItem
