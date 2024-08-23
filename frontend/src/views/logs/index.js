import { CTab, CTabContent, CTabList, CTabPanel, CTabs } from '@coreui/react'
import React, { useState } from 'react'
import { useLocation } from 'react-router-dom'
import FaultSiteLogs from './FaultSiteLogs'
import FullLogs from './FullLogs'
import Empty from 'src/components/Empty/Empty'
function LogsPage() {
  const location = useLocation()
  const [activeItemKey, setActiveItemKey] = useState('faultSite')
  return (
    <div
      style={{ width: '100%', overflow: 'hidden', height: 'calc(100vh - 150px)' }}
      className="text-xs"
    >
      <CTabs
        activeItemKey={activeItemKey}
        className="border-tab h-full flex flex-col"
        onChange={(key) => setActiveItemKey(key)}
      >
        <CTabList variant="tabs" className="flex-grow-0 flex-shrink-0 text-base">
          <CTab itemKey="faultSite">故障现场日志</CTab>
          <CTab itemKey="full">全量日志</CTab>
        </CTabList>
        <CTabContent className="flex-grow flex-shrink overflow-hidden">
          <CTabPanel className="p-3 h-full " itemKey="faultSite">
            {activeItemKey === 'faultSite' && <FaultSiteLogs />}
          </CTabPanel>
          <CTabPanel className="p-3 h-full" itemKey="full">
            {activeItemKey === 'full' && <Empty context="敬请期待" />}
            {/* <FullLogs logsList={logsList} /> */}
          </CTabPanel>
        </CTabContent>
      </CTabs>
    </div>
  )
}
export default LogsPage
