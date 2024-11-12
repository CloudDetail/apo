import { Collapse, Empty } from 'antd'
import React, { useEffect, useState } from 'react'
import HighlightCode from 'src/core/components/HighlightCode/HighlightCode'
import LoadingSpinner from 'src/core/components/Spinner'

const ErrorLog = ({ logMap, logCount, loading, node }) => {
  const [items, setItems] = useState([])
  const [activeKey, setActiveKey] = useState([])
  useEffect(() => {
    let items = []
    let activeKey = []
    logMap?.map((log: any) => {
      if (log?.name && log?.id !== undefined && log?.values.length > 0) {
        if (activeKey.length <= 0) {
          activeKey = log.id
        }
        items.push({
          key: log.id,
          label: log.name,
          children: <HighlightCode rows={log.values} title={log.name} />,
        })
      }
    })
    setItems(items)
    setActiveKey(activeKey)
  }, [logMap])
  return logMap && logCount > 0 ? (
    <React.Fragment>
      <label style={{ marginBottom: '5px' }}>相关日志</label>
      <Collapse items={items} activeKey={activeKey}></Collapse>
    </React.Fragment>
  ) : loading ? (
    <LoadingSpinner loading={loading} />
  ) : node.isProfiled || node.hasOwnProperty('logs') ? (
    <Empty />
  ) : (
    <Empty description="未采集日志" />
  )
}
export default ErrorLog
