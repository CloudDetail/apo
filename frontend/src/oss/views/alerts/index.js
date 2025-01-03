/**
 * Copyright 2024 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import { CTab, CTabContent, CTabList, CTabPanel, CTabs } from '@coreui/react'
import React, { useState } from 'react'
import { useLocation } from 'react-router-dom'
import AlertsRule from './AlertsRule'
import AlertsNotify from './AlertsNotify'
function AlertsPage() {
  const location = useLocation()
  const [activeItemKey, setActiveItemKey] = useState('rule')
  return (
    <div
      style={{ width: '100%', overflow: 'hidden', height: 'calc(100vh - 100px)' }}
      className="text-xs"
    >
      <CTabs
        activeItemKey={activeItemKey}
        className="border-tab h-full flex flex-col"
        onChange={(key) => setActiveItemKey(key)}
      >
        <CTabList variant="tabs" className="flex-grow-0 flex-shrink-0 text-base">
          <CTab itemKey="rule">告警规则</CTab>
          <CTab itemKey="notify">告警通知</CTab>
        </CTabList>
        <CTabContent className="flex-grow flex-shrink overflow-hidden">
          <CTabPanel className="p-3 h-full " itemKey="rule">
            {activeItemKey === 'rule' && <AlertsRule />}
          </CTabPanel>
          <CTabPanel className="p-3 h-full" itemKey="notify">
            {activeItemKey === 'notify' && <AlertsNotify />}
            {/* <FullLogs logsList={logsList} /> */}
          </CTabPanel>
        </CTabContent>
      </CTabs>
    </div>
  )
}
export default AlertsPage
