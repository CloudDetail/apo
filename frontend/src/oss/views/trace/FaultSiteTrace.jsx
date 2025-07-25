/**
 * Copyright 2024 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import React, { useMemo, useState, useCallback } from 'react'
import { traceTableMock } from './mock'
import BasicTable from 'src/core/components/Table/basicTable'
import { convertTime, TimestampToISO } from 'src/core/utils/time'
import { getTracePageListApi } from 'core/api/trace.js'
import EndpointTableModal from './component/JaegerIframeModal'
import LoadingSpinner from 'src/core/components/Spinner'
import LogsTraceFilter from 'src/oss/components/Filter/LogsTraceFilter'
import { DefaultTraceFilters } from 'src/constants'
import TraceErrorType from './component/TraceErrorType'
import { AiOutlineInfoCircle } from 'react-icons/ai'
import { Tooltip, Button } from 'antd'
import { useTranslation } from 'react-i18next'
import { BasicCard } from 'src/core/components/Card/BasicCard'
import { useLogsTraceFilterContext } from 'src/oss/contexts/LogsTraceFilterContext'
import { useDebounce } from 'react-use'
import { useSelector } from 'react-redux'

function FaultSiteTrace() {
  const { t } = useTranslation('oss/trace')
  // const [startTime, setStartTime] = useState(null)
  const [tracePageList, setTracePageList] = useState([])
  // const [endTime, setEndTime] = useState(null)
  // const [service, setService] = useState(null)
  // const [instance, setInstance] = useState(null)
  // const [traceId, setTraceId] = useState(null)
  // const [endpoint, setEndpoint] = useState(null)
  const [from, setFrom] = useState(null)
  const [to, setTo] = useState(null)
  const [pageIndex, setPageIndex] = useState(1)
  const [pageSize, setPageSize] = useState(10)
  const [loading, setLoading] = useState(false)
  const [total, setTotal] = useState(0)
  const [modalVisible, setModalVisible] = useState(false)
  // 传入modal的traceid
  const [selectTraceId, setSelectTraceId] = useState('')
  const { dataGroupId } = useSelector((state) => state.dataGroupReducer)
  const {
    clusterIds,
    services,
    instance,
    traceId,
    namespaces,
    startTime,
    endTime,
    minDuration,
    maxDuration,
    faultTypeList,
    endpoint,
    isFilterDone,
  } = useLogsTraceFilterContext((ctx) => ctx)
  const openJeagerModal = (traceId) => {
    setSelectTraceId(traceId)
    setModalVisible(true)
  }
  const column = [
    {
      title: t('trace.serviceName'),
      accessor: 'serviceName',
    },
    {
      title: t('trace.namespace'),
      accessor: 'labels',
      Cell: ({ value }) => {
        return value?.namespace ? value?.namespace : <span className="text-slate-400">N/A</span>
      },
    },
    {
      title: t('trace.instanceName'),
      accessor: 'instanceId',
    },
    {
      title: t('trace.endpoint'),
      accessor: 'endpoint',
    },

    {
      title: (
        <Tooltip
          title={
            <div>
              <div className="text-[#D3D3D3]">{t('trace.slowFault')}</div>
              <div className="text-[#D3D3D3] mt-2">{t('trace.errorFault')}</div>
              <div className="text-[#D3D3D3] mt-2">{t('trace.noFault')}</div>
            </div>
          }
        >
          <div className="flex flex-row justify-center items-center">
            {t('trace.faultStatus')}
            <AiOutlineInfoCircle size={16} className="ml-2" />
          </div>
        </Tooltip>
      ),
      accessor: 'flags',
      Cell: ({ value }) => {
        let typeList = []
        if (value.is_slow) {
          typeList.push('slow')
        }
        if (value.is_error) {
          typeList.push('error')
        }
        if (typeList.length === 0) {
          typeList.push('normal')
        }
        return typeList.map((type) => <TraceErrorType type={type} key={type} />)
      },
    },
    {
      title: t('trace.responseTime'),
      accessor: 'duration',
      Cell: ({ value }) => {
        return convertTime(value, 'ms', 2) + 'ms'
      },
    },
    {
      title: t('trace.occurTime'),
      accessor: 'timestamp',
      Cell: ({ value }) => {
        return convertTime(value, 'yyyy-mm-dd hh:mm:ss')
      },
    },
    {
      title: t('trace.traceId'),
      accessor: 'traceId',
      Cell: (props) => {
        const { value } = props

        return (
          <a
            className=" cursor-pointer text-[var(--ant-color-link)]"
            onClick={() => openJeagerModal(value)}
          >
            {value}
          </a>
        )
      },
    },
    {
      title: t('trace.operation'),
      accessor: 'action',
      // minWidth: 140,
      Cell: (props) => {
        const { row } = props
        const { traceId, serviceName, instanceId } = row.original

        const formattedStartTime = TimestampToISO(startTime)
        const formattedEndTime = TimestampToISO(endTime)
        const clusterIdsParam = clusterIds
          ? `&clusterIds=${encodeURIComponent(Array.isArray(clusterIds) ? clusterIds.join(',') : clusterIds)}`
          : ''
        const targetUrl = `#/logs/fault-site?logs-from=${encodeURIComponent(formattedStartTime)}&logs-to=${encodeURIComponent(formattedEndTime)}&service=${encodeURIComponent(serviceName)}&instance=${encodeURIComponent(instanceId)}&traceId=${encodeURIComponent(traceId)}${clusterIdsParam}&groupId=${dataGroupId}`
        return (
          <div className="flex flex-col">
            <Button
              onClick={() => window.open(targetUrl)}
              className="my-1"
              variant="outlined"
              color="primary"
            >
              {t('trace.viewLogs')}
            </Button>
          </div>
        )
      },
    },
  ]
  const prepareFilter = () => {
    let filters = []
    if (namespaces?.length > 0) {
      let filter = DefaultTraceFilters.namespace
      filter.operation = 'LIKE'
      filter.value = namespaces
      filters.push(filter)
    }
    let duration = DefaultTraceFilters.duration
    if (minDuration) {
      filters.push({
        ...duration,
        operation: 'GREATER_THAN',
        value: [(minDuration * 1000).toString()],
      })
    }
    if (maxDuration) {
      filters.push({
        ...duration,
        operation: 'LESS_THAN',
        value: [(maxDuration * 1000).toString()],
      })
    }

    // The selection of fault types is achieved through nested sub-filters
    const createSingleOption = (type, operator = 'AND') => {
      const subFilters = []
      for (let key in type) {
        if (key === 'slow' || key === 'error') {
          subFilters.push({
            ...DefaultTraceFilters[key],
            operation: 'EQUAL',
            value: [type[key]],
          })
        }
      }

      return { mergeSep: operator, subFilters }
    }

    const subFilters = []
    faultTypeList?.forEach((faultType) => {
      switch (faultType) {
        case 'slow':
          subFilters.push(createSingleOption({ slow: 'true', error: 'false' }))
          break
        case 'error':
          subFilters.push(createSingleOption({ slow: 'false', error: 'true' }))
          break
        case 'normal':
          subFilters.push(createSingleOption({ slow: 'false', error: 'false' }))
          break
        case 'slowAndError':
          subFilters.push(createSingleOption({ slow: 'true', error: 'true' }))
          break
      }
    })

    filters.push({
      mergeSep: 'OR',
      subFilters,
    })

    return filters
  }
  const filterUndefinedOrEmpty = (obj) => {
    return Object.fromEntries(
      Object.entries(obj).filter(([_, value]) => {
        if (value === undefined || value === null) return false
        if (typeof value === 'string' && value.trim() === '') return false
        if (Array.isArray(value) && value.length === 0) return false
        return true
      }),
    )
  }
  const getTraceData = useCallback(() => {
    setLoading(true)
    const { containerId, node: nodeName, pid, id: instanceId } = instance?.[0] ?? {}
    const queryParams = filterUndefinedOrEmpty({
      startTime,
      endTime,
      service: services,
      // instance: instance,
      traceId: traceId,
      endpoint: endpoint,
      namespace: namespaces,
      pageNum: pageIndex,
      pageSize: pageSize,
      containerId,
      nodeName,
      clusterIds,
      pid,
      filters: prepareFilter(),
      groupId: dataGroupId,
    })
    getTracePageListApi(queryParams)
      .then((res) => {
        const totalPages = Math.ceil(res.pagination.total / res.pagination.pageSize)
        if (pageIndex > totalPages && totalPages > 0) {
          setPageIndex(totalPages)
          return
        }
        setLoading(false)
        setTracePageList(res?.list ?? [])
        setTotal(res?.pagination.total)
        if (sessionStorage.getItem('openJaegerModalAfterLoad')) {
          openJeagerModal(true)
          setSelectTraceId(traceId)
          sessionStorage.removeItem('openJaegerModalAfterLoad')
        }
        //
      })
      .catch((error) => {
        console.log(error)
        setTracePageList([])
        setLoading(false)
      })
  }, [
    clusterIds,
    services,
    instance,
    traceId,
    namespaces,
    startTime,
    endTime,
    minDuration,
    maxDuration,
    faultTypeList,
    endpoint,
    pageIndex,
    pageSize,
    dataGroupId,
  ])
  const handleTableChange = (pageIndex, pageSize) => {
    if (pageSize && pageIndex) {
      setPageSize(pageSize), setPageIndex(pageIndex)
    }
  }
  // useEffect(() => {
  //   if (startTime && endTime) {
  //     getTraceData()
  //   }
  // }, [startTime, endTime, service, instance, traceId, pageIndex, endpoint])
  useDebounce(
    () => {
      if (startTime && endTime && isFilterDone && dataGroupId !== null) {
        getTraceData()
      }
    },
    300,
    [
      clusterIds,
      services,
      instance,
      traceId,
      namespaces,
      startTime,
      endTime,
      isFilterDone,
      faultTypeList,
      endpoint,
      pageIndex,
      minDuration,
      maxDuration,
      dataGroupId,
    ],
  )
  const tableProps = useMemo(() => {
    return {
      columns: column,
      data: tracePageList,
      showBorder: false,
      loading: false,
      onChange: handleTableChange,
      pagination: {
        pageSize: pageSize,
        pageIndex: pageIndex,
        total: total,
      },
    }
  }, [tracePageList, column])
  return (
    <BasicCard>
      <LoadingSpinner loading={loading} />

      <BasicCard.Header>
        <div className="w-full flex-shrink-0 flex-grow mb-2">
          <LogsTraceFilter type="trace" />
        </div>
      </BasicCard.Header>

      <BasicCard.Table>{traceTableMock && <BasicTable {...tableProps} />}</BasicCard.Table>

      <EndpointTableModal
        traceId={selectTraceId}
        visible={modalVisible}
        closeModal={() => setModalVisible(false)}
      />
    </BasicCard>
    // </PropsProvider>
  )
}
export default FaultSiteTrace
