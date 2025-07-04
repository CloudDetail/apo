/**
 * Copyright 2024 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import { CCol, CRow } from '@coreui/react'
import React, { useState } from 'react'
import { useSelector } from 'react-redux'
import { useDebounce } from 'react-use'
import { getK8sEventApi } from 'core/api/serviceInfo'
import Empty from 'src/core/components/Empty/Empty'
import LoadingSpinner from 'src/core/components/Spinner'
import { usePropsContext } from 'src/core/contexts/PropsContext'
import { selectProcessedTimeRange } from 'src/core/store/reducers/timeRangeReducer'
import { useTranslation } from 'react-i18next'
import { useServiceInfoContext } from 'src/oss/contexts/ServiceInfoContext'

function K8sInfo() {
  const setPanelsStatus = useServiceInfoContext((ctx) => ctx.setPanelsStatus)
  const openTab = useServiceInfoContext((ctx) => ctx.openTab)

  const [data, setData] = useState({})
  const { serviceName, clusterIds } = usePropsContext()
  const { dataGroupId } = useSelector((state) => state.dataGroupReducer)
  const [colList, setColList] = useState([])
  const [loading, setLoading] = useState(false)
  const { startTime, endTime } = useSelector(selectProcessedTimeRange)
  const { t } = useTranslation('oss/serviceInfo')
  const mockColList = [
    {
      name: t('K8sInfo.appChangeFailed'),
      status: 'success',
      value: 0,
      weekValue: 0,
      monthValue: 0,
    },
    {
      name: t('K8sInfo.appScaling'),
      status: 'success',
      value: 2,
      weekValue: 664,
      monthValue: 881,
    },
    {
      name: t('K8sInfo.appScalingLimit'),
      status: 'success',
      value: 2,
      weekValue: 642,
      monthValue: 848,
    },
    {
      name: t('K8sInfo.isolationRemoval'),
      status: 'success',
      value: 0,
      weekValue: 0,
      monthValue: 0,
    },
    {
      name: t('K8sInfo.podStartFailed'),
      status: 'error',
      value: 21,
      weekValue: 3700,
      monthValue: 15447,
    },
    {
      name: t('K8sInfo.imagePullFailed'),
      status: 'success',
      value: 3,
      weekValue: 98,
      monthValue: 98,
    },
    {
      name: t('K8sInfo.podEvicted'),
      status: 'success',
      value: 0,
      weekValue: 0,
      monthValue: 0,
    },
    {
      name: t('K8sInfo.podOOM'),
      status: 'error',
      value: 24,
      weekValue: 4243,
      monthValue: 18242,
    },
    {
      name: t('K8sInfo.clusterResourceInsufficient'),
      status: 'success',
      value: 0,
      weekValue: 0,
      monthValue: 0,
    },
    {
      name: t('K8sInfo.nodeOOM'),
      status: 'success',
      value: 0,
      weekValue: 0,
      monthValue: 0,
    },
    {
      name: t('K8sInfo.nodeRestart'),
      status: 'success',
      value: 0,
      weekValue: 0,
      monthValue: 0,
    },
    {
      name: t('K8sInfo.nodeFDInsufficient'),
      status: 'success',
      value: 0,
      weekValue: 0,
      monthValue: 0,
    },
  ]
  const getData = () => {
    if (startTime && endTime) {
      setLoading(true)
      getK8sEventApi({
        startTime: startTime,
        endTime: endTime,
        service: serviceName,
        groupId: dataGroupId,
        clusterIds,
      })
        .then((res) => {
          setData(res.data ?? {})
          setLoading(false)
          if (res?.status === 'critical') openTab('k8s')
          setPanelsStatus('k8s', res.status)
        })
        .catch((error) => {
          setData({})
          setPanelsStatus('k8s', 'unknown')
          setLoading(false)
        })
    }
  }
  // useEffect(() => {
  //   getData()
  // }, [serviceName, startTime, endTime])
  //防抖避免跳转使用旧时间
  useDebounce(
    () => {
      getData()
    },
    300, // 延迟时间 300ms
    [startTime, endTime, serviceName, dataGroupId, clusterIds],
  )
  return (
    <>
      <div>
        <CRow xs={{ cols: 6 }}>
          {Object.keys(data).map((key) => {
            const item = data[key]
            return (
              <CCol key={key} className="text-center py-3.5">
                <div className="text-sm mb-2">{item.displayName}</div>
                <div
                  className="text-lg mb-2 text-[#467ffc] font-bold"
                  style={{ color: item.severity === 'Warning' ? '#dc2625' : '#467ffc' }}
                >
                  {item.counts.current ?? 0}
                </div>
                <div className="text-xs mb-1" style={{ color: 'rgba(248, 249, 250, 0.45)' }}>
                  {t('K8sInfo.timesIn7Days')}:{item.counts.lastWeek}
                </div>
                <div className="text-xs" style={{ color: 'rgba(248, 249, 250, 0.45)' }}>
                  {t('K8sInfo.timesIn30Days')}:{item.counts.lastMonth}
                </div>
              </CCol>
            )
          })}
          <LoadingSpinner loading={loading} />
        </CRow>
        {!loading && (!data || Object.keys(data).length === 0) && <Empty />}
      </div>
    </>
  )
}

export default K8sInfo
