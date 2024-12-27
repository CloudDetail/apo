import { CTab, CTabContent, CTabList, CTabPanel, CTabs } from '@coreui/react'
import React, { useState } from 'react'
import { useLocation } from 'react-router-dom'
import Empty from 'src/core/components/Empty/Empty'
import Trace from './Trace'
import FullTrace from './FullTrace'
import { useTranslation } from 'react-i18next' // 引入i18n

function LogsPage() {
  const location = useLocation()
  const [activeItemKey, setActiveItemKey] = useState('faultSite')
  const { t } = useTranslation('oss/trace') // 使用i18n
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
          <CTab itemKey="faultSite">{t('index.faultSite')}</CTab>
          <CTab itemKey="full">{t('index.full')}</CTab>
        </CTabList>
        <CTabContent className="flex-grow flex-shrink overflow-hidden">
          <CTabPanel className="p-3 h-full " itemKey="faultSite">
            {activeItemKey === 'faultSite' && <Trace />}
          </CTabPanel>
          <CTabPanel className="p-3 h-full" itemKey="full">
            {activeItemKey === 'full' && <FullTrace />}
            {/* <FullLogs logsList={logsList} /> */}
          </CTabPanel>
        </CTabContent>
      </CTabs>
    </div>
  )
}
export default LogsPage
