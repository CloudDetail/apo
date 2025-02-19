/**
 * Copyright 2025 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import { Card } from 'antd'
import ClusterTable from './ClusterTable'

export default function DataIntegrationPage() {
  return (
    <Card style={{ height: 'calc(100vh - 100px)' }}>
      <ClusterTable />
    </Card>
  )
}
