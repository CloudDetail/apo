/**
 * Copyright 2024 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import { CTab, CTabContent, CTabList, CTabPanel, CTabs } from '@coreui/react'
import React, { useState } from 'react'
import { useLocation } from 'react-router-dom'
import AlertsRule from './AlertsRule'
import AlertsNotify from './AlertsNotify'
import { useTranslation } from 'react-i18next' // 引入i18n

function AlertsPage() {
  const location = useLocation()
  const [activeItemKey, setActiveItemKey] = useState('rule')
  const { t } = useTranslation('oss/alert') // 使用i18n
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
          <CTab itemKey="rule">{t('index.rule')}</CTab>
          <CTab itemKey="notify">{t('index.notify')}</CTab>
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
