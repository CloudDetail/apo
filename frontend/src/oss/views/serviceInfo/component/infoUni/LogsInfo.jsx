/**
 * Copyright 2024 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import React, { useState, useMemo } from 'react'
import BasicTable from 'src/core/components/Table/basicTable'
import DelayLineChart from 'src/core/components/Chart/DelayLineChart'
import Timeline from './TimeLine'
import { usePropsContext } from 'src/core/contexts/PropsContext'
import { getLogMetricsApi } from 'core/api/serviceInfo'
import { getStep } from 'src/core/utils/step'
import { selectSecondsTimeRange } from 'src/core/store/reducers/timeRangeReducer'
import { useSelector } from 'react-redux'
import { useDebounce } from 'react-use'
import { useTranslation } from 'react-i18next'

function LogsInfo() {
  const { t } = useTranslation('oss/serviceInfo')
  const [data, setData] = useState()
  const [loading, setLoading] = useState(false)
  const { serviceName, endpoint } = usePropsContext()
  const { startTime, endTime } = useSelector(selectSecondsTimeRange)

  const column = useMemo(
    () => [
      {
        title: t('logsInfo.instanceName'),
        accessor: 'name',
        customWidth: 150,
      },
      {
        title: t('logsInfo.logErrorCount'),
        accessor: 'logs',
        Cell: (props) => {
          const { value } = props
          return (
            <div className="h-52 w-80">
              <DelayLineChart data={value} timeRange={{ startTime, endTime }} type={'logs'} />
            </div>
          )
        },
      },
      {
        title: t('logsInfo.responseTimeP90'),
        accessor: 'latency',
        Cell: (props) => {
          const { value } = props
          return (
            <div className="h-52 w-80">
              <DelayLineChart data={value} timeRange={{ startTime, endTime }} type={'p90'} />
            </div>
          )
        },
      },
      {
        title: t('logsInfo.errorRate'),
        accessor: 'errorRate',
        Cell: (props) => {
          const { value } = props
          return (
            <div className="h-52 w-80">
              <DelayLineChart data={value} timeRange={{ startTime, endTime }} type={'errorRate'} />
            </div>
          )
        },
      },
      {
        title: t('logsInfo.logInfo'),
        accessor: 'logInfo',
        customWidth: 320,
        Cell: (props) => {
          const { value, row } = props
          return (
            <Timeline
              instance={row.original.name}
              nodeName={row.original.nodeName}
              pid={row.original.pid}
              containerId={row.original.containerId}
              type="logsInfo"
              instanceName={row.values.name}
              startTime={startTime}
              endTime={endTime}
            />
          )
        },
      },
    ],
    [startTime, endTime],
  )
  const getData = () => {
    if (startTime && endTime) {
      setLoading(true)
      getLogMetricsApi({
        startTime: startTime,
        endTime: endTime,
        service: serviceName,
        endpoint: endpoint,
        step: getStep(startTime, endTime),
      })
        .then((res) => {
          setData(res ?? [])
          setLoading(false)
        })
        .catch((error) => {
          setData([])
          setLoading(false)
        })
    }
  }
  // useEffect(() => {
  //   getData()
  // }, [serviceName, endpoint, startTime, endTime])
  //防抖避免跳转使用旧时间
  useDebounce(
    () => {
      getData()
    },
    300, // 延迟时间 300ms
    [startTime, endTime, serviceName, endpoint],
  )
  const tableProps = useMemo(() => {
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
  }, [data, serviceName, column])
  return (
    <>
      <div className="text-xs">{data && <BasicTable {...tableProps} />}</div>
    </>
  )
}
export default LogsInfo
