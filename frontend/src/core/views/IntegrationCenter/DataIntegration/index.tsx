/**
 * Copyright 2025 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import { Card } from 'antd'
import ClusterTable from './ClusterTable'
import CustomCard from 'src/core/components/Card/CustomCard'

export default function DataIntegrationPage() {
  return (
    <CustomCard>
      <ClusterTable />
    </CustomCard>
  )
}
