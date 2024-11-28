import LogItem from './LogItem'
import { Empty, List } from 'antd'
const QueryList = ({ logs, openContextModal = null, loading }) => {
  return (
    <div className="overflow-auto h-full">
      {logs?.length > 0 && (
        <List
          dataSource={logs}
          renderItem={(log) => (
            <List.Item>
              <LogItem log={log} openContextModal={openContextModal} />
            </List.Item>
          )}
        />
      )}
      {logs?.length === 0 && !loading && (
        <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="暂无日志数据" />
      )}
    </div>
  )
}

export default QueryList
