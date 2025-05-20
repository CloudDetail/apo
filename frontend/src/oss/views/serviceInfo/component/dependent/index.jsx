/**
 * Copyright 2024 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import React from 'react'
import DependentTable from './DependentTable'
import { usePropsContext } from 'src/core/contexts/PropsContext'
import TimelapseLineChart from './TimelapseLineChart'
import { useSelector } from 'react-redux'
import { selectSecondsTimeRange } from 'src/core/store/reducers/timeRangeReducer'
import { useTranslation } from 'react-i18next'
import { Card, Tabs } from 'antd'
import styles from './index.module.scss'
function DependentTabs() {
  const { serviceName, endpoint } = usePropsContext()
  const { startTime, endTime } = useSelector(selectSecondsTimeRange)
  const { t } = useTranslation('oss/serviceInfo')
  const items = [
    {
      key: 'timelapse',
      label: t('dependent.index.timelapseComparison'),
      children: (
        <TimelapseLineChart
          endpoint={endpoint}
          startTime={startTime}
          endTime={endTime}
          serviceName={serviceName}
        />
      ),
      style: {
        height: '100%',
      },
    },
    {
      key: 'table',
      label: t('dependent.index.similarityRanking'),
      children: (
        <DependentTable
          endpoint={endpoint}
          startTime={startTime}
          endTime={endTime}
          serviceName={serviceName}
        />
      ),
    },
  ]
  return (
    <Card
      size="small"
      title={
        <>
          {serviceName}
          {t('dependent.index.allDependencies')}
        </>
      }
      className="mb-4 ml-1 h-[350px] w-3/5 whitespace-normal flex flex-col"
      classNames={{
        body: 'h-0 flex-1 p-0',
        title: 'whitespace-normal',
      }}
      styles={{
        title: {
          whiteSpace: 'normal',
        },
      }}
    >
      <div className="h-full w-full">
        <Tabs
          type="card"
          size="small"
          items={items}
          className={styles.tabs}
          tabBarStyle={{ margin: 0 }}
        />
      </div>
    </Card>
  )
}

export default DependentTabs
