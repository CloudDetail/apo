import React, { useMemo, useState, useEffect, useRef } from 'react'
import { traceTableMock } from './mock'
import BasicTable from 'src/components/Table/basicTable'
import { getTimestampRange, timeRangeList } from 'src/store/reducers/timeRangeReducer'
import { convertTime, ISOToTimestamp } from 'src/utils/time'
import { useLocation, useSearchParams } from 'react-router-dom'
import { getTracePageListApi } from 'src/api/trace.js'
import StatusInfo from 'src/components/StatusInfo'
import { PropsProvider } from 'src/contexts/PropsContext'
import EndpointTableModal from './component/JaegerIframeModal'
import { useSelector } from 'react-redux'
import LoadingSpinner from 'src/components/Spinner'
import LogsTraceFilter from 'src/components/Filter/LogsTraceFilter'
import { DefaultTraceFilters } from 'src/constants'
import TraceErrorType from './component/TraceErrorType'

function Trace() {
  const [searchParams] = useSearchParams()
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
          getTraceData()
        } else {
          setPageIndex(1)
        }
      } else if (prev.pageIndex !== pageIndex) {
        getTraceData()
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
      title: '服务名',
      accessor: 'serviceName',
    },
    {
      title: '命名空间',
      accessor: 'labels',
      Cell: ({ value }) => {
        return value?.namespace ? value?.namespace : <span className="text-slate-400">N/A</span>
      },
    },
    {
      title: '实例名',
      accessor: 'instanceId',
    },
    {
      title: '服务端点',
      accessor: 'endpoint',
    },

    {
      title: '故障状态',
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
      title: '持续时间',
      accessor: 'duration',
      Cell: ({ value }) => {
        return convertTime(value, 'ms', 2) + 'ms'
      },
    },
    {
      title: '发生时间',
      accessor: 'timestamp',
      Cell: ({ value }) => {
        return convertTime(value, 'yyyy-mm-dd hh:mm:ss')
      },
    },
    {
      title: 'TraceId',
      accessor: 'traceId',
      Cell: (props) => {
        const { value } = props

        return (
          <a className=" cursor-pointer" onClick={() => openJeagerModal(value)}>
            {value}
          </a>
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
    if (faultTypeList?.includes('normal')) {
      if (faultTypeList.length === 2) {
        let type = faultTypeList.includes('slow') ? 'error' : 'slow'
        filters.push({
          ...DefaultTraceFilters[type],
          operation: 'IN',
          value: ['false'],
        })
      } else if (faultTypeList.length === 1) {
        filters.push({
          ...DefaultTraceFilters['error'],
          operation: 'IN',
          value: ['false'],
        })
        filters.push({
          ...DefaultTraceFilters['slow'],
          operation: 'IN',
          value: ['false'],
        })
      }
    } else {
      faultTypeList?.forEach((type) => {
        filters.push({
          ...DefaultTraceFilters[type],
          operation: 'IN',
          value: ['true'],
        })
      })
    }
    // if (isSlow) {
    //   filters.push({
    //     ...DefaultTraceFilters.isSlow,
    //     operation: 'IN',
    //     value: [isSlow.toString()],
    //   })
    // }
    // if (isError) {
    //   filters.push({
    //     ...DefaultTraceFilters.isError,
    //     operation: 'IN',
    //     value: [isError.toString()],
    //   })
    // }
    return filters
  }
  const getTraceData = () => {
    const { containerId, nodeName, pid } = instanceOption[instance] ?? {}
    setLoading(true)
    getTracePageListApi({
      startTime,
      endTime,
      service: service,
      // instance: instance,
      traceId: traceId,
      endpoint: endpoint,
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
  }
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
  }, [tracePageList])
  return (
    // <PropsProvider
    //   value={{
    //     startTime,
    //     endTime,
    //     service,
    //     instance,
    //     traceId,
    //     endpoint,
    //   }}
    // >
    <>
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
    </>
    // </PropsProvider>
  )
}
export default Trace
