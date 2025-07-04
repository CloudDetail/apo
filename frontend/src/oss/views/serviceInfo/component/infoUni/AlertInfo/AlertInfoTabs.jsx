/**
 * Copyright 2024 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import { CTab, CTabContent, CTabList, CTabPanel, CTabs } from '@coreui/react'
import React, { useEffect, useState } from 'react'
import { useSelector } from 'react-redux'
import { getServiceAlertEventApi } from 'core/api/serviceInfo'
import { usePropsContext } from 'src/core/contexts/PropsContext'
import { selectProcessedTimeRange } from 'src/core/store/reducers/timeRangeReducer'
import AlertInfoTable from './AlertInfoTable'
import Empty from 'src/core/components/Empty/Empty'
import { useDebounce } from 'react-use'
import { useTranslation } from 'react-i18next'
import { RuleGroupMap } from 'src/constants'
import { useServiceInfoContext } from 'src/oss/contexts/ServiceInfoContext'

export default function AlertInfoTabs() {
  const setPanelsStatus = useServiceInfoContext((ctx) => ctx.setPanelsStatus)
  const openTab = useServiceInfoContext((ctx) => ctx.openTab)
  const { serviceName, endpoint, clusterIds } = usePropsContext()
  const { dataGroupId } = useSelector((state) => state.dataGroupReducer)
  const [loading, setLoading] = useState(true)
  const { startTime, endTime } = useSelector(selectProcessedTimeRange)
  const { t, i18n } = useTranslation('oss/serviceInfo')

  const [tabList, setTabList] = useState([])

  const prepareData = (result) => {
    let tabList = []
    Object.keys(result.events).forEach((group) => {
      if (RuleGroupMap[group] === undefined) {
        return
      }

      let num = 0
      const groupMap = result.events[group]

      const data = Object.keys(groupMap).map((item) => {
        num += groupMap[item].length
        return {
          name: item,
          list: groupMap[item],
        }
      })
      tabList.push({
        key: group,
        name: RuleGroupMap[group],
        data,
        num,
      })
    })
    setTabList(tabList)
  }

  const getAlertEvents = () => {
    if (startTime && endTime) {
      setLoading(true)
      getServiceAlertEventApi({
        startTime,
        endTime,
        service: serviceName,
        endpoint: endpoint,
        status: 'firing',
        groupId: dataGroupId,
        clusterIds,
      })
        .then((res) => {
          setLoading(false)
          prepareData(res)
          if (res?.status === 'critical') openTab('alert')
          setPanelsStatus('alert', res.status)
        })
        .catch((error) => {
          setTabList([])
          setPanelsStatus('alert', 'unknown')
          setLoading(false)
        })
    }
  }

  useDebounce(
    () => {
      getAlertEvents()
    },
    300,
    [serviceName, startTime, endTime, endpoint, dataGroupId, clusterIds],
  )

  useEffect(() => {
    // 监听语言变化，重新获取告警事件
    getAlertEvents()
  }, [i18n.language])

  return (
    <div>
      {tabList?.length > 0 && (
        <CTabs activeItemKey={tabList[0].key}>
          <CTabList variant="tabs">
            {tabList.map((tab) => (
              <CTab itemKey={tab.key}>{tab.name}</CTab>
            ))}
          </CTabList>
          <CTabContent>
            {tabList.map((tab) => (
              <CTabPanel className="p-3 text-xs" itemKey={tab.key}>
                <AlertInfoTable data={tab.data} />
              </CTabPanel>
            ))}
          </CTabContent>
        </CTabs>
      )}
      {tabList?.length === 0 && !loading && (
        <Empty context={t('alertInfo.alertInfoTabs.noAlertEvents')} />
      )}
    </div>
  )
}
