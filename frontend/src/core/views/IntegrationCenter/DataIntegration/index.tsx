import { Card } from 'antd'
import ClusterTable from './ClusterTable'

export default function DataIntegrationPage() {
  return (
    <Card style={{ height: 'calc(100vh - 100px)' }}>
      <ClusterTable />
    </Card>
  )
}
