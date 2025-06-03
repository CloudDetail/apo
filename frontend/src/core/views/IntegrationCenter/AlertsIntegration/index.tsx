/**
 * Copyright 2025 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import { Card, Divider, Flex } from 'antd'
import { AlertIntegrationProvider } from 'src/core/contexts/AlertIntegrationContext'
import IntegrationDrawer from './ConfigDrawer'
import AlertsDatasourceList from './AlertsDatasourceList'
import AlertsIntegrationTable from './AlertsIntegrationTable'
import { BasicCard } from 'src/core/components/Card/BasicCard'

const AlertsIntegrationPage = () => {
  return (
    <AlertIntegrationProvider>
      <IntegrationDrawer />
      <BasicCard>
        <Flex
          style={{
            // boxShadow: '0 0 10px rgba(0, 0, 0, 0.1)',
            // padding: 10,
            height: '100%'
          }}
        >
          <div className="h-full w-[310px] px-1 flex-shrink-0">
            <AlertsDatasourceList />
          </div>
          <Divider type="vertical" className="h-full text-[var(--ant-color-border)]" />
          <div className="px-4 flex-1 overflow-hidden">
            <AlertsIntegrationTable />
          </div>
        </Flex>
      </BasicCard>
    </AlertIntegrationProvider>
  )
}
export default AlertsIntegrationPage
