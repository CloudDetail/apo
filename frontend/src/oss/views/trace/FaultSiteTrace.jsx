/**
 * Copyright 2024 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import React, { useMemo, useState, useEffect, useRef, useCallback } from 'react'
import { traceTableMock } from './mock'
import BasicTable from 'src/core/components/Table/basicTable'
import { getTimestampRange, timeRangeList } from 'src/core/store/reducers/timeRangeReducer'
import { convertTime, ISOToTimestamp, TimestampToISO } from 'src/core/utils/time'
import { useLocation, useSearchParams } from 'react-router-dom'
import { getTracePageListApi } from 'core/api/trace.js'
import EndpointTableModal from './component/JaegerIframeModal'
import { useSelector } from 'react-redux'
import LoadingSpinner from 'src/core/components/Spinner'
import LogsTraceFilter from 'src/oss/components/Filter/LogsTraceFilter'
import { DefaultTraceFilters } from 'src/constants'
import TraceErrorType from './component/TraceErrorType'
import { AiOutlineInfoCircle } from 'react-icons/ai'
import { Card, Tooltip, Button } from 'antd'
import { useTranslation } from 'react-i18next'
import { useDebouncedCallback } from 'use-debounce';

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

  const {
    startTime,
    endTime,
    service,
    instance,
    traceId,
    instanceOption,
    endpoint,
    namespace,
    minDuration,
    maxDuration,
    faultTypeList,
  } = useSelector((state) => state.urlParamsReducer)
  const previousValues = useRef({
    startTime: null,
    endTime: null,
    service: '',
    instance: '',
    traceId: '',
    endpoint: '',
    pageIndex: 1,
    selectInstanceOption: {},
    //filter
    namespace: '',
    faultTypeList: null,
    minDuration: '',
    maxDuration: '',
  })
  useEffect(() => {
    const prev = previousValues.current
    let paramsChange = false

    if (prev.startTime !== startTime) {
      paramsChange = true
    }
    if (prev.endTime !== endTime) {
      paramsChange = true
    }
    if (prev.service !== service) {
      paramsChange = true
    }
    if (prev.instance !== instance) {
      paramsChange = true
    }
    if (prev.traceId !== traceId) {
      paramsChange = true
    }
    if (prev.namespace !== namespace) {
      paramsChange = true
    }
    if (prev.minDuration !== minDuration) {
      paramsChange = true
    }
    if (prev.maxDuration !== maxDuration) {
      paramsChange = true
    }
    // console.log(prev.isError, isError)
    // if (prev.isError !== isError) {
    //   paramsChange = true
    // }
    // if (prev.isSlow !== isSlow) {
    //   paramsChange = true
    // }
    if (prev.faultTypeList !== faultTypeList) {
      paramsChange = true
    }
    if (prev.endpoint !== endpoint) {
      // console.log('endpoint -> pre:', prev.endpoint, 'now:', endpoint)
      paramsChange = true
    }
    const selectInstanceOption = instanceOption[instance]
    if (JSON.stringify(prev.selectInstanceOption) !== JSON.stringify(selectInstanceOption)) {
      // console.log(
      //   'selectInstanceOption -> pre:',
      //   prev.selectInstanceOption,
      //   'now:',
      //   selectInstanceOption,
      // )
      paramsChange = true
    }
    if (instance && !selectInstanceOption) {
      paramsChange = false
    }

    previousValues.current = {
      startTime,
      endTime,
      service,
      instance,
      traceId,
      pageIndex,
      endpoint,
      selectInstanceOption,
      namespace,
      minDuration,
      maxDuration,
      faultTypeList,
    }
    if (startTime && endTime) {
      if (paramsChange) {
        if (pageIndex === 1) {
          debouncedGetTraceData()
        } else {
          setPageIndex(1)
        }
      } else if (prev.pageIndex !== pageIndex) {
        debouncedGetTraceData()
      }
    }
  }, [
    startTime,
    endTime,
    service,
    instance,
    traceId,
    endpoint,
    pageIndex,
    instanceOption,
    namespace,
    minDuration,
    maxDuration,
    faultTypeList,
  ])
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
          <a className=" cursor-pointer text-blue-500" onClick={() => openJeagerModal(value)}>
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
        const { row } = props;
        const traceId = row.original.traceId;
        const serviceName = row.original.serviceName;
        const instanceId = row.original.instanceId;

        const formattedStartTime = TimestampToISO(startTime);
        const formattedEndTime = TimestampToISO(endTime);

        const targetUrl = `#/logs/fault-site?logs-from=${encodeURIComponent(formattedStartTime)}&logs-to=${encodeURIComponent(formattedEndTime)}&service=${encodeURIComponent(serviceName)}&instance=${encodeURIComponent(instanceId)}&traceId=${encodeURIComponent(traceId)}`
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
    if (namespace) {
      let filter = DefaultTraceFilters.namespace
      filter.operation = 'LIKE'
      filter.value = [namespace]
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

  if (faultTypeList?.length >= 1 && faultTypeList?.length <= 3) {
    // Helper function to create basic filter
    const createBasicFilter = (type, operator, value) => ({
      ...DefaultTraceFilters[type],
      operation: operator,
      value: [value]
    });

    // Helper function to create combined filters
    const createCombinedFilters = (logicalOperator, values, types = ['slow', 'error']) => {
      const subFilters = types.map((type, index) =>
        createBasicFilter(type, 'IN', values[index])
      );
      return { mergeSep: logicalOperator, subFilters };
    };

    // Filter creation operations
    const filterOperations = {
      // Handle single fault type selection
      single: {
        slowAndError: () => ['error', 'slow'].forEach(type =>
          filters.push(createBasicFilter(type, 'IN', 'true'))
        ),
        normal: () => ['error', 'slow'].forEach(type =>
          filters.push(createBasicFilter(type, 'IN', 'false'))
        ),
        slow: () => filters.push(createCombinedFilters('AND', ['true', 'false'])),
        error: () => filters.push(createCombinedFilters('AND', ['false', 'true']))
      },

      // Handle two fault type combinations
      double: {
        slowSlowAndError: () => filters.push(createBasicFilter('slow', 'IN', 'true')),
        errorSlowAndError: () => filters.push(createBasicFilter('error', 'IN', 'true')),
        errorNormal: () => filters.push(createBasicFilter('slow', 'IN', 'false')),
        slowNormal: () => filters.push(createBasicFilter('error', 'IN', 'false')),
        normalSlowAndError: () => {
        filters.push(createCombinedFilters('OR', ['false', 'true']));
        filters.push(createCombinedFilters('OR', ['true', 'false']));
        },
        errorSlow: () => {
        filters.push(createCombinedFilters('OR', ['false', 'false']));
        filters.push(createCombinedFilters('OR', ['true', 'true']));
        }
      },

      // Handle three fault type combinations (exclusion cases)
      triple: {
        noSlowAndError: () => filters.push(createCombinedFilters('OR', ['false', 'false'])),
        noNormal: () => filters.push(createCombinedFilters('OR', ['true', 'true'])),
        noSlow: () => filters.push(createCombinedFilters('OR', ['false', 'true'])),
        noError: () => filters.push(createCombinedFilters('OR', ['true', 'false']))
      }
    };

    // Determine operation based on selection count
    switch (faultTypeList.length) {
      case 1:
        filterOperations.single[faultTypeList[0]]?.();
        break;

      case 2: {
        const combinationKey = faultTypeList
          .sort() // Alphabetical sorting ensures consistent key for same combination
          .map((word, index) =>
            index === 0 ? word : word[0].toUpperCase() + word.slice(1) // Convert to camelCase
          )
          .join('');
        filterOperations.double[combinationKey]?.();
        break;
      }

      case 3: {
        const excludedType = ['slowAndError', 'normal', 'slow', 'error']
          .find(type => !faultTypeList.includes(type));
        const exclusionKey = `no${excludedType[0].toUpperCase()}${excludedType.slice(1)}`;
        filterOperations.triple[exclusionKey]?.();
        break;
      }

      default:
      console.warn('Unsupported fault type combination');
    }
  }

    return filters
  }
  const getTraceData = useCallback(() => {
    const { containerId, nodeName, pid } = instanceOption[instance] ?? {}
    setLoading(true)
    getTracePageListApi({
      startTime,
      endTime,
      service: service ? [service] : undefined,
      // instance: instance,
      traceId: traceId,
      endpoint: endpoint,
      namespace: namespace ? [namespace] : undefined,
      pageNum: pageIndex,
      pageSize: pageSize,
      containerId,
      nodeName,
      pid,
      filters: prepareFilter(),
    })
      .then((res) => {
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
    startTime,
    endTime,
    service,
    instance,
    traceId,
    endpoint,
    namespace,
    pageIndex,
    pageSize,
    instanceOption,
    minDuration,
    maxDuration,
    faultTypeList,
  ]);
  const debouncedGetTraceData = useDebouncedCallback(getTraceData, 500);
  const handleTableChange = (props) => {
    if (props.pageSize && props.pageIndex) {
      setPageSize(props.pageSize)
      setPageIndex(props.pageIndex)
    }
  }
  // useEffect(() => {
  //   if (startTime && endTime) {
  //     getTraceData()
  //   }
  // }, [startTime, endTime, service, instance, traceId, pageIndex, endpoint])
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
        pageCount: Math.ceil(total / pageSize),
      },
    }
  }, [tracePageList, column])
  return (
    <Card
      className="h-full flex flex-col overflow-hidden text-xs px-2"
      style={{ height: 'calc(100vh - 120px)' }}
      styles={{ body: { padding: '8px', height: '100%' } }}
    >
      <LoadingSpinner loading={loading} />
      <div className="text-xs flex flex-col h-full overflow-hidden">
        <div className="flex-shrink-0 flex-grow">
          <LogsTraceFilter type="trace" />
        </div>
        {traceTableMock && <BasicTable {...tableProps} />}
      </div>
      <EndpointTableModal
        traceId={selectTraceId}
        visible={modalVisible}
        closeModal={() => setModalVisible(false)}
      />
    </Card>
    // </PropsProvider>
  )
}
export default FaultSiteTrace
