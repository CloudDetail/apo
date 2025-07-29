import { Table } from 'antd'
import dayjs from 'dayjs'

const Logs = ({ args }: { args: any }) => {
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
      dataIndex: 'message',
    },
  ]

  return <Table scroll={{ y: 500 }} columns={column} dataSource={args} size="small" />
}

export default Logs
