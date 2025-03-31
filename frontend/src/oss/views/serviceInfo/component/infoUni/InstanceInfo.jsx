/**
 * Copyright 2024 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import React, { useMemo, useState, useEffect } from 'react'
import BasicTable from 'src/core/components/Table/basicTable'
import StatusInfo from 'src/core/components/StatusInfo'
import TempCell from 'src/core/components/Table/TempCell'
import { usePropsContext } from 'src/core/contexts/PropsContext'
import { getServiceInstancesApi } from 'core/api/serviceInfo'
import { selectSecondsTimeRange } from 'src/core/store/reducers/timeRangeReducer'
import { useSelector } from 'react-redux'
import { getStep } from 'src/core/utils/step'
import { convertTime } from 'src/core/utils/time'
import { useDebounce } from 'react-use'
import { useTranslation } from 'react-i18next'
import { useServiceInfoContext } from 'src/oss/contexts/ServiceInfoContext'

function InstanceInfo() {
  const { t } = useTranslation('oss/serviceInfo')
  const setPanelsStatus = useServiceInfoContext((ctx) => ctx.setPanelsStatus)
  const setDashboardVariable = useServiceInfoContext((ctx) => ctx.setDashboardVariable)
  const openTab = useServiceInfoContext((ctx) => ctx.openTab)

  const [data, setData] = useState()
  const { serviceName, endpoint } = usePropsContext()
  const [loading, setLoading] = useState(false)
  const { startTime, endTime } = useSelector(selectSecondsTimeRange)
  const column = useMemo(
    () => [
      {
        title: t('instanceInfo.instanceName'),
        accessor: 'name',
        customWidth: 150,
      },
      {
        title: t('instanceInfo.avgResponseTime'),
        accessor: 'latency',
        minWidth: 150,
        Cell: (props) => {
          const { value } = props
          return <TempCell type="latency" data={value} timeRange={{ startTime, endTime }} />
        },
      },
      {
        title: t('instanceInfo.errorRate'),
        accessor: 'errorRate',
        minWidth: 150,
        Cell: (props) => {
          const { value } = props
          return <TempCell type="errorRate" timeRange={{ startTime, endTime }} data={value} />
        },
      },
      {
        title: t('instanceInfo.throughput'),
        accessor: 'tps',
        minWidth: 150,
        Cell: (props) => {
          const { value } = props
          return <TempCell type="tps" timeRange={{ startTime, endTime }} data={value} />
        },
      },
      {
        title: t('instanceInfo.logErrorCount'),
        accessor: 'logs',
        minWidth: 150,
        Cell: (props) => {
          const { value } = props
          return <TempCell type="logs" timeRange={{ startTime, endTime }} data={value} />
        },
      },
      {
        title: t('instanceInfo.infrastructureStatus'),
        accessor: 'infrastructureStatus',
        Cell: (props) => {
          const { value, row, column } = props
          const alertReason = row.original?.alertReason?.[column.id]
          return (
            <>
              <StatusInfo status={value} alertReason={alertReason} title={column.title} />
            </>
          )
        },
      },
      {
        title: t('instanceInfo.networkQualityStatus'),
        accessor: 'netStatus',
        Cell: (props) => {
          const { value, row, column } = props
          const alertReason = row.original?.alertReason?.[column.id]
          return (
            <>
              <StatusInfo status={value} alertReason={alertReason} title={column.title} />
            </>
          )
        },
      },
      {
        title: t('instanceInfo.k8sEventStatus'),
        accessor: 'k8sStatus',
        Cell: (props) => {
          const { value, row, column } = props
          const alertReason = row.original?.alertReason?.[column.id]
          return (
            <>
              <StatusInfo status={value} alertReason={alertReason} title={column.title} />
            </>
          )
        },
      },

      {
        title: t('instanceInfo.nodeInfo'),
        accessor: 'nodeName',
        minWidth: 150,
        Cell: (props) => {
          return (
            <div>
              <div className="flex ">
                <span className="text-gray-400 flex-shrink-0 flex-grow-0">
                  {t('instanceInfo.nodeName')}：
                </span>
                {props.value}
              </div>
              <div className="flex">
                <span className="text-gray-400 flex-shrink-0 flex-grow-0">
                  {t('instanceInfo.nodeIP')}：
                </span>
                {props.row.original.nodeIP}
              </div>
            </div>
          )
        },
      },
      {
        title: t('instanceInfo.lastDeploymentOrRestartTime'),
        accessor: `timestamp`,
        Cell: (props) => {
          const { value } = props
          return (
            <>
              {value !== null ? (
                convertTime(value, 'yyyy-mm-dd hh:mm:ss')
              ) : (
                <span className="text-slate-400">N/A</span>
              )}
            </>
          )
        },
      },
    ],
    [startTime, endTime],
  )
  const getData = () => {
    if (startTime && endTime) {
      setLoading(true)
      getServiceInstancesApi({
        startTime: startTime,
        endTime: endTime,
        serviceName: serviceName,
        endpoint: endpoint,
        step: getStep(startTime, endTime),
      })
        .then((res) => {
          setData(res.data ?? [])
          setLoading(false)
          if (res?.status === 'critical') openTab('instance')

          setPanelsStatus('instance', res.status)
        })
        .catch((error) => {
          setData([])
          setPanelsStatus('instance', 'unknown')
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
    [startTime, endTime, serviceName, endpoint],
  )
  useEffect(() => {
    const namespaceList = [
      ...new Set(data?.map((obj) => obj.namespace).filter((namespace) => namespace !== '')),
    ]
    const podList = [...new Set(data?.map((obj) => obj.name).filter((name) => name !== ''))]
    setDashboardVariable({ namespaceList, podList, service: serviceName })
  }, [data])

  const tableProps = useMemo(() => {
    console.log('1212')
    return {
      columns: column,
      data: data,
      showBorder: false,
      loading: false,
      pagination: {
        pageSize: 5,
        pageIndex: 1,
        total: data?.length || 0,
      },
      scrollY: 300,
    }
  }, [data, column])
  return (
    <>
      <div className="text-xs">{data && <BasicTable {...tableProps} />}</div>
    </>
  )
}

export default React.memo(InstanceInfo)
