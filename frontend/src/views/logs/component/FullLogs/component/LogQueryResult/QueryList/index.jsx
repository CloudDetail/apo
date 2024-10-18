import React from 'react'
import LogItem from './LogItem'
import { Empty, List } from 'antd'

const QueryList = ({ logs, openContextModal = null }) => {
  return (
    <div className="overflow-y-auto h-full">
      {logs?.length > 0 ? (
        <List
          dataSource={logs}
          bordered={false}
          renderItem={(item, index) => (
            <List.Item key={index}>
              <LogItem log={item} openContextModal={openContextModal} />
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
