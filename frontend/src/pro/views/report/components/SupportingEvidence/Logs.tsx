import React from 'react'
import { Table } from 'antd'
import dayjs from 'dayjs'
const Logs = ({ logs }) => {
  const dataSource = logs.map((log, index) => ({
    ...log,
    key: `${log.timestamp || 'log'}-${log.body || log.message || index}`,
  }))
  const column = [
    {
      title: 'Date',
      dataIndex: 'timestamp',
      width: 150,
      render: (value) => {
        return dayjs(value / 1000).format('YYYY-MM-DD HH:mm:ss')
      },
    },
    {
      title: 'Message',
      dataIndex: 'body',
      render: (value, record) => value || record.message || '-',
    },
  ]

  return (
    <Table
      scroll={{ y: 500 }}
      columns={column}
      dataSource={dataSource}
      size="small"
    />
  )
}
export default Logs
