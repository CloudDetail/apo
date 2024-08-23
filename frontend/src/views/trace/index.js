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
import { useSelector } from 'react-redux'
import LoadingSpinner from 'src/components/Spinner'

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

  const { startTime, endTime, service, instance, traceId, instanceOption, endpoint } = useSelector(
    (state) => state.urlParamsReducer,
  )
  const previousValues = useRef({
    startTime: null,
    endTime: null,
    service: '',
    instance: '',
    traceId: '',
    endpoint: '',
    pageIndex: 1,
    selectInstanceOption: {},
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

    previousValues.current = {
      startTime,
      endTime,
      service,
      instance,
      traceId,
      pageIndex,
      endpoint,
      selectInstanceOption,
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
  }, [startTime, endTime, service, instance, traceId, endpoint, pageIndex, instanceOption])
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
      <LoadingSpinner loading={loading} />
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
