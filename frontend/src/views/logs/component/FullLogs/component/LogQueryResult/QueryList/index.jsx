import React, { useEffect, useState } from 'react'
import { AiFillCaretDown, AiFillCaretRight } from 'react-icons/ai'
import { useLogsContext } from 'src/contexts/LogsContext'
import LogItem from './LogItem'
import { Empty, List } from 'antd'

const QueryList = () => {
  const { logs } = useLogsContext()
  return (
    <div className="overflow-y-auto h-full">
      {logs?.length > 0 ? (
        <List
          // pagination={{ current: 1, pageSize: 10 }}
          dataSource={logs}
          bordered={false}
          renderItem={(item, index) => (
            <List.Item key={index}>
              <LogItem log={item} />
            </List.Item>
          )}
        />
      ) : (
        <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="暂无日志数据" />
      )}
    </div>
  )
}
export default QueryList
