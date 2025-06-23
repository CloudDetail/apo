import React from 'react'
import { Table } from 'antd'
import dayjs from 'dayjs'
const Logs = ({ logs }) => {
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
      title: 'Massage',
      dataIndex: 'body',
    },
  ]

  return <Table scroll={{ y: 500 }} columns={column} dataSource={logs} size="small" />
}
export default Logs
