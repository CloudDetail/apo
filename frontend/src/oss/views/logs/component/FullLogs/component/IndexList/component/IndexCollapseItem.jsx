import { Collapse, List, Progress, Tag, Tooltip } from 'antd'
import React, { useEffect, useState } from 'react'
import { useSelector } from 'react-redux'
import { useSearchParams } from 'react-router-dom'
import { useLogsContext } from 'src/core/contexts/LogsContext'
import { selectProcessedTimeRange } from 'src/core/store/reducers/timeRangeReducer'
import { ISOToTimestamp } from 'src/core/utils/time'

const IndexCollapseItem = ({ field }) => {
  const {
    query = '',
    updateQuery,
    fieldIndexMap,
    getFieldIndexData,
    defaultFields,
    hiddenFields,
  } = useLogsContext()

  const [searchParams] = useSearchParams()
  const { startTime, endTime } = useSelector(selectProcessedTimeRange)
  const [loading, setLoading] = useState(false)
  const clickIndex = (index) => {
    let newQueryPart = '`' + field + '` =' + "'" + index.indexName + "'"
    // 检查 query 是否已经包含 newQueryPart
    if (!query.includes(newQueryPart)) {
      let newQuery = query
      if (newQuery.length > 0) {
        newQuery += ' AND '
      }
      newQuery += newQueryPart
      updateQuery(newQuery) // 更新查询
    }
  }
  useEffect(() => {
    if ((defaultFields.includes(field) || hiddenFields.includes(field)) && !fieldIndexMap[field]) {
      setLoading(true)

      getFieldIndexData({
        startTime,
        endTime,
        column: field,
      }).finally(() => {
        setLoading(false)
      })
    }
  }, [field, defaultFields, hiddenFields, fieldIndexMap])
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
