/**
 * Copyright 2024 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import {
  CCard,
  CCardBody,
  CCardHeader,
  CTab,
  CTabContent,
  CTabList,
  CTabPanel,
  CTabs,
} from '@coreui/react'
import React, { useState, useEffect } from 'react'
import DependentTable from './DependentTable'
import { usePropsContext } from 'src/core/contexts/PropsContext'
import TimelapseLineChart from './TimelapseLineChart'
import { useSelector } from 'react-redux'
import { selectSecondsTimeRange } from 'src/core/store/reducers/timeRangeReducer'
import { useTranslation } from 'react-i18next'
import i18next from 'i18next'

function DependentTabs() {
  const { serviceName, endpoint } = usePropsContext()
  const [serviceList, setServiceList] = useState([])
  const { startTime, endTime } = useSelector(selectSecondsTimeRange)
  const [activeItemKey, setActiveItemKey] = useState('timelapse')
  const { t } = useTranslation('oss/serviceInfo')
  const [language, setLanguage] = useState(i18next.language)

  return (
    <CCard className="mb-4 ml-1 h-[350px] p-2  w-3/5">
      <CCardHeader>
        {serviceName}
        {t('dependent.index.allDependencies')}
      </CCardHeader>
      <CCardBody className="text-xs overflow-hidden p-0">
        <CTabs
          activeItemKey={activeItemKey}
          className="w-full h-full overflow-hidden flex flex-col "
          onChange={(value) => setActiveItemKey(value)}
        >
          <CTabList variant="tabs" className="flex-grow-0 flex-shrink-0">
            <CTab itemKey="timelapse">{t('dependent.index.timelapseComparison')}</CTab>
            <CTab itemKey="table">{t('dependent.index.similarityRanking')}</CTab>
          </CTabList>
          <CTabContent className="h-full overflow-hidden flex-grow">
            <CTabPanel itemKey="timelapse" className="overflow-hidden h-full">
              {activeItemKey === 'timelapse' && (
                <TimelapseLineChart
                  endpoint={endpoint}
                  startTime={startTime}
                  endTime={endTime}
                  serviceName={serviceName}
                />
              )}
            </CTabPanel>
            <CTabPanel itemKey="table" className="h-full overflow-hidden">
              <DependentTable
                endpoint={endpoint}
                startTime={startTime}
                endTime={endTime}
                serviceName={serviceName}
              />
            </CTabPanel>
          </CTabContent>
        </CTabs>
      </CCardBody>
    </CCard>
  )
}

export default DependentTabs
