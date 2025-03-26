/**
 * Copyright 2024 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import React, { useMemo, useState } from 'react'
import BasicTable from 'src/core/components/Table/basicTable'
import TempCell from 'src/core/components/Table/TempCell'
import { usePropsContext } from 'src/core/contexts/PropsContext'
import { selectSecondsTimeRange } from 'src/core/store/reducers/timeRangeReducer'
import { useSelector } from 'react-redux'
import { getStep } from 'src/core/utils/step'
import { getServiceSqlMetrics } from 'core/api/serviceInfo'
import LoadingSpinner from 'src/core/components/Spinner'
import { useDebounce, useUpdateEffect } from 'react-use'
import { useTranslation } from 'react-i18next'

export default function SqlMetrics() {
  const { t } = useTranslation('oss/serviceInfo')
  const [data, setData] = useState()
  const [status, setStatus] = useState()
  const { serviceName } = usePropsContext()
  const [loading, setLoading] = useState(false)
  const { startTime, endTime } = useSelector(selectSecondsTimeRange)
  const [pageIndex, setPageIndex] = useState(1)
  const [pageSize, setPageSize] = useState(5)
  const [total, setTotal] = useState(0)
  const column = useMemo(
    () => [
      {
        title: t('sqlMetrics.dbConnection'),
        accessor: 'dbUrl',
        Cell: ({ value }) => {
          return <>{value ? value : <span className="text-slate-400">N/A</span>}</>
        },
      },
      {
        title: t('sqlMetrics.dbOperation'),
        accessor: 'dbOperation',
        Cell: ({ value }) => {
          return <>{value ? value : <span className="text-slate-400">N/A</span>}</>
        },
      },

      {
        title: t('sqlMetrics.dbName'),
        accessor: 'dbName',
      },
      {
        title: t('sqlMetrics.avgResponseTime'),
        accessor: 'latency',
        minWidth: 200,
        Cell: (props) => {
          const { value } = props
          return <TempCell type="latency" data={value} timeRange={{ startTime, endTime }} />
        },
      },
      {
        title: t('sqlMetrics.errorRate'),
        accessor: 'errorRate',
        minWidth: 200,

        Cell: (props) => {
          const { value } = props
          return <TempCell type="errorRate" timeRange={{ startTime, endTime }} data={value} />
        },
      },
      {
        title: t('sqlMetrics.throughput'),
        accessor: 'tps',

        minWidth: 200,
        Cell: (props) => {
          const { value } = props
          return <TempCell type="tps" timeRange={{ startTime, endTime }} data={value} />
        },
      },

      {
        title: t('sqlMetrics.dbSource'),
        accessor: 'dbSystem',
      },
    ],
    [startTime, endTime],
  )
  const getData = () => {
    if (startTime && endTime) {
      setLoading(true)
      getServiceSqlMetrics({
        startTime: startTime,
        endTime: endTime,
        service: serviceName,
        sortBy: 'latency',
        step: getStep(startTime, endTime),
        currentPage: pageIndex,
        pageSize: pageSize,
      })
        .then((res) => {
          setTotal(res?.pagination.total)
          setData(res.sqlOperationDetails ?? [])
          // setStatus(res.status)
          setLoading(false)
          // handlePanelStatus(res.status)
        })
        .catch((error) => {
          setData([])
          // handlePanelStatus('unknown')
          setLoading(false)
        })
    }
  }
  // useEffect(() => {
  //   if (pageIndex === 1) {
  //     getData()
  //   } else {
  //     setPageIndex(1)
  //   }
  // }, [serviceName, startTime, endTime])

  useDebounce(
    () => {
      if (pageIndex === 1) {
        getData()
      } else {
        setPageIndex(1)
      }
    },
    300, // 延迟时间 300ms
    [startTime, endTime, serviceName],
  )
  useUpdateEffect(() => {
    getData()
  }, [pageIndex, pageSize])

  const handleTableChange = (pageIndex, pageSize) => {
    if (pageSize && pageIndex) {
      setPageSize(pageSize), setPageIndex(pageIndex)
    }
  }
  const tableProps = useMemo(() => {
    return {
      columns: column,
      data: data,
      showBorder: false,
      loading: false,
      onChange: handleTableChange,
      pagination: {
        pageSize: pageSize,
        pageIndex: pageIndex,
        total: total,
      },
      scrollY: 300,
    }
  }, [serviceName, data, column])
  return (
    <>
      <div className="text-xs relative">
        <LoadingSpinner loading={loading} />
        {data && <BasicTable {...tableProps} />}
      </div>
    </>
  )
}
