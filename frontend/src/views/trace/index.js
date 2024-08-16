import React, { useMemo, useState, useEffect, useRef } from 'react'
import { traceTableMock } from './mock'
import BasicTable from 'src/components/Table/basicTable'
import TraceFilters from './TraceFilters'
import { getTimestampRange, timeRangeList } from 'src/store/reducers/timeRangeReducer'
import { convertTime, ISOToTimestamp } from 'src/utils/time'
import { useLocation, useSearchParams } from 'react-router-dom'
import { getTracePageListApi } from 'src/api/trace.js'
import StatusInfo from 'src/components/StatusInfo'
import { PropsProvider } from 'src/contexts/PropsContext'
import EndpointTableModal from './component/JaegerIframeModal'
import { useUrlParams } from 'src/contexts/UrlParamsContext'

function TracePage() {
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
  const [loading, setLoading] = useState(true)
  const [total, setTotal] = useState(0)
  const [modalVisible, setModalVisible] = useState(false)
  // 传入modal的traceid
  const [selectTraceId, setSelectTraceId] = useState('')
  const { urlParamsState } = useUrlParams()
  const {
    startTime,
    endTime,
    service,
    instance,
    traceId,
    filtersLoaded,
    instanceOption,
    endpoint,
  } = urlParamsState
  const previousValues = useRef({
    startTime: null,
    endTime: null,
    service: '',
    instance: '',
    traceId: '',
    endpoint: '',
    pageIndex: 1,
    selectInstanceOption: {},
    filtersLoaded: false,
  })
  // useEffect(() => {
  //   const urlService = searchParams.get('service')
  //   const urlInstance = searchParams.get('instance')
  //   const urlTraceId = searchParams.get('traceId')
  //   const urlEndpoint = searchParams.get('endpoint')
  //   const urlFrom = searchParams.get('trace-from')
  //   const urlTo = searchParams.get('trace-to')
  //   let changeParams = false
  //   if (urlService !== service) {
  //     setService(urlService)
  //     changeParams = true
  //   }
  //   if (urlInstance !== instance) {
  //     setInstance(urlInstance)
  //     changeParams = true
  //   }
  //   if (urlTraceId !== traceId) {
  //     setTraceId(urlTraceId)
  //     changeParams = true
  //   }
  //   if (urlEndpoint !== endpoint) {
  //     setEndpoint(urlEndpoint)
  //     changeParams = true
  //   }
  //   if (urlFrom && urlTo && (urlFrom !== from || urlTo !== to)) {
  //     const urlTimeRange = timeRangeList.find((item) => item.from === urlFrom && item.to === urlTo)
  //     if (urlTimeRange) {
  //       //说明是快速范围，根据rangetype 获取当前开始结束时间戳
  //       const { startTime, endTime } = getTimestampRange(urlTimeRange.rangeType)
  //       setStartTime(startTime)
  //       setEndTime(endTime)
  //       changeParams = true
  //     } else {
  //       //说明可能是精确时间，先判断是不是可以转化成微妙时间戳
  //       const startTimestamp = ISOToTimestamp(urlFrom)
  //       const endTimestamp = ISOToTimestamp(urlTo)
  //       if (startTimestamp && endTimestamp) {
  //         setStartTime(startTimestamp)
  //         setEndTime(endTimestamp)
  //         changeParams = true
  //       }
  //     }
  //     setFrom(urlFrom)
  //     setTo(urlTo)
  //   }
  //   if (changeParams) {
  //     setPageIndex(1)
  //   }
  //   // console.log(window.location.href)
  // }, [searchParams])
  useEffect(() => {
    const prev = previousValues.current
    let paramsChange = false

    if (prev.startTime !== startTime) {
      console.log('startTime -> pre:', prev.startTime, 'now:', startTime)
      paramsChange = true
    }
    if (prev.endTime !== endTime) {
      console.log('endTime -> pre:', prev.endTime, 'now:', endTime)
      paramsChange = true
    }
    if (prev.service !== service) {
      console.log('service -> pre:', prev.service, 'now:', service)
      paramsChange = true
    }

    if (prev.traceId !== traceId) {
      console.log('traceId -> pre:', prev.traceId, 'now:', traceId)
      paramsChange = true
    }
    if (prev.endpoint !== endpoint) {
      console.log('endpoint -> pre:', prev.endpoint, 'now:', endpoint)
      paramsChange = true
    }
    const selectInstanceOption = instanceOption[instance]
    if (JSON.stringify(prev.selectInstanceOption) !== JSON.stringify(selectInstanceOption)) {
      console.log(
        'selectInstanceOption -> pre:',
        prev.selectInstanceOption,
        'now:',
        selectInstanceOption,
      )
      paramsChange = true
    }
    if (instance && !selectInstanceOption) {
      paramsChange = false
    }
    if (prev.filtersLoaded !== filtersLoaded) {
      paramsChange = true
    }
    console.log(
      '-----------',
      paramsChange,
      startTime,
      endTime,
      service,
      instance,
      traceId,
      endpoint,
      pageIndex,
      instanceOption,
      filtersLoaded,
    )

    previousValues.current = {
      startTime,
      endTime,
      service,
      instance,
      traceId,
      pageIndex,
      endpoint,
      selectInstanceOption,
      filtersLoaded,
    }
    if (startTime && endTime && filtersLoaded) {
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
  }, [startTime, endTime, service, instance, traceId, pageIndex, instanceOption, filtersLoaded])
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
      title: '实例名',
      accessor: 'instanceId',
    },
    {
      title: '服务端点',
      accessor: 'endpoint',
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
      title: '状态',
      accessor: 'isError',
      Cell: ({ value }) => {
        return <StatusInfo status={value ? 'critical' : 'normal'} />
      },
    },
    {
      title: 'TraceID',
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
  const getTraceData = () => {
    const { containerId, nodeName, pid } = instanceOption[instance] ?? {}
    getTracePageListApi({
      startTime,
      endTime,
      service: service,
      // instance: instance,
      traceId: traceId,
      endpoint: endpoint,
      pageNum: pageIndex,
      pageSize: 10,
      containerId,
      nodeName,
      pid,
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
      .catch(() => {
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
        pageSize: 10,
        pageIndex: pageIndex,
        pageCount: Math.ceil(total / 10),
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
      <div
        style={{ width: '100%', overflow: 'hidden', height: 'calc(100vh - 120px)' }}
        className="text-xs flex flex-col"
      >
        <div className="flex-shrink-0 flex-grow">
          <TraceFilters />
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
export default TracePage
