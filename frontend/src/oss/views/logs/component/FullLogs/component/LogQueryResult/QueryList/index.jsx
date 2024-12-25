import LogItem from './LogItem'
import { Empty, List } from 'antd'
import { useTranslation } from 'react-i18next' // 引入i18n

const QueryList = ({ logs, openContextModal = null, loading }) => {
  const { t } = useTranslation('oss/fullLogs') // 使用i18n

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
        <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description={t('queryList.noDataText')} />
      )}
    </div>
  )
}

export default QueryList
