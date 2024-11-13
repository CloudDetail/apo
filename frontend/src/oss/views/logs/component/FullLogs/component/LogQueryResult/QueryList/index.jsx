import React from 'react'
import { Virtuoso } from 'react-virtuoso'
import LogItem from './LogItem'
import { Empty } from 'antd'
const QueryList = ({ logs, openContextModal = null }) => {
  return (
    <div className="overflow-y-auto h-full">
      {logs?.length > 0 ? (
        <Virtuoso
          style={{ height: '100%', width: '100%' }}
          data={logs}
          itemContent={(index) => (
            <div style={{ padding: '10px' }}>
              <LogItem log={logs[index]} openContextModal={openContextModal} />
            </div>
          )}
        />
      ) : (
        <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="暂无日志数据" />
      )}
    </div>
  )
}

export default QueryList
