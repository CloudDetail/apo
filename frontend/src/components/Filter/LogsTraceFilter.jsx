import { CFormInput } from '@coreui/react'
import React, { useState, useEffect } from 'react'
import { useSelector } from 'react-redux'
import { useSearchParams } from 'react-router-dom'
import {
  getServiceInstanceListApi,
  getServiceInstancOptionsListApi,
  getServiceListApi,
} from 'src/api/service'
import DateTimeRangePickerCom from 'src/components/DateTime/DateTimeRangePickerCom'
import { usePropsContext } from 'src/contexts/PropsContext'
import { CustomSelect } from 'src/components/Select'
import {
  getTimestampRange,
  selectProcessedTimeRange,
  timeRangeList,
} from 'src/store/reducers/timeRangeReducer'
import { ISOToTimestamp } from 'src/utils/time'
import { useInstance } from 'src/contexts/InstanceContext'
import { useUrlParams } from 'src/contexts/UrlParamsContext'

const LogsTraceFilter = React.memo(({ type }) => {
  const [searchParams, setSearchParams] = useSearchParams()

  const [serviceList, setServiceList] = useState([])
  const [instanceList, setInstanceList] = useState([])

  const [selectServiceName, setSelectServiceName] = useState('')
  const [selectInstance, setSelectInstance] = useState('')
  // 应该深入
  const [inputTraceId, setInputTraceId] = useState('')
  const [inputEndpoint, setInputEndpoint] = useState('')
  const [startTime, setStartTime] = useState(null)
  const [endTime, setEndTime] = useState(null)
  const { urlParamsState, dispatch } = useUrlParams()
  const [isLoaded, setIsLoaded] = useState(false)
  // const { startTime, endTime, service, instance, traceId, endpoint } = urlParamsState

  const setInstanceOption = (value) => {
    dispatch({ type: 'setInstanceOption', payload: value })
  }
  const updateUrlParamsState = (params) => {
    dispatch({ type: 'setUrlParamsState', payload: params })
  }
  const updateFilterLoaded = (value) => {
    dispatch({ type: 'setFiltersLoaded', payload: value })
  }
  const clearUrlParamsState = (value) => {
    dispatch({ type: 'clearUrlParamsState', payload: value })
  }

  const onChangeService = (props) => {
    setInstanceList([])
    onChangeInstance('')
    setSelectServiceName(props ?? '')
    const params = new URLSearchParams(searchParams)
    params.set('service', props)
    params.set('instance', '')
    setSearchParams(params)
    if (!props) {
      setIsLoaded(true)
    } else {
      setIsLoaded(false)
    }
  }
  const onChangeInstance = (props) => {
    // console.log(props)
    setSelectInstance(props)
    const params = new URLSearchParams(searchParams)
    params.set('instance', props)
    setSearchParams(params)
    setIsLoaded(true)
  }
  const onChangeTraceId = (event) => {
    setInputTraceId(event.target.value)
    const params = new URLSearchParams(searchParams)
    params.set('traceId', event.target.value)
    setSearchParams(params)
  }
  const onChangeEndpoint = (event) => {
    setInputEndpoint(event.target.value)
    const params = new URLSearchParams(searchParams)
    params.set('endpoint', event.target.value)
    setSearchParams(params)
  }

  const getServiceListData = () => {
    getServiceListApi({ startTime, endTime })
      .then((res) => {
        setServiceList(res ?? [])
        console.log(res, selectServiceName, selectInstance)
        if (!selectServiceName) {
          setIsLoaded(true)
        } else {
          if (!res.includes(selectServiceName)) {
            onChangeService('')
          } else {
            if (!selectInstance) {
              setIsLoaded(true)
            }
          }
        }
      })
      .catch((error) => {
        console.log(error)
        setServiceList([])
        setSelectServiceName('')
        setIsLoaded(true)
      })
  }

  const getInstanceListData = () => {
    getServiceInstancOptionsListApi({
      startTime,
      endTime,
      service: selectServiceName,
    })
      .then((res) => {
        setInstanceList(res)
        setInstanceOption(res)
        if (!res[selectInstance]) {
          onChangeInstance('')
        }
      })
      .catch((error) => {
        console.log(error)
        setInstanceList(null)
        setInstanceOption(null)
        setSelectInstance('')
      })
      .finally(() => {
        setIsLoaded(true)
      })
  }
  useEffect(() => {
    const urlService = searchParams.get('service') ?? ''
    const urlInstance = searchParams.get('instance') ?? ''
    const urlTraceId = searchParams.get('traceId') ?? ''
    const urlEndpoint = searchParams.get('endpoint') ?? ''
    const urlFrom = searchParams.get(type + '-from')
    const urlTo = searchParams.get(type + '-to')
    // console.log('url参数改变', urlService, urlInstance, urlTraceId, urlEndpoint, urlFrom, urlTo)

    let needLoadFilter = false
    if (urlService !== selectServiceName) {
      // updateStateService(urlService)
      // console.log(urlService, selectServiceName)
      setSelectServiceName(urlService || '')

      needLoadFilter = true
    }
    if (urlInstance !== selectInstance) {
      // updateStateInstance(urlInstance)
      setSelectInstance(urlInstance)

      needLoadFilter = true
    }
    if (urlTraceId !== inputTraceId) {
      setInputTraceId(urlTraceId)
      // updateStateTraceId(urlTraceId)
    }
    if (urlEndpoint !== inputEndpoint) {
      setInputEndpoint(urlEndpoint)
      // updateStateEndpoint(urlEndpoint)
    }
    if (urlFrom && urlTo) {
      const urlTimeRange = timeRangeList.find((item) => item.from === urlFrom && item.to === urlTo)
      if (urlTimeRange) {
        //说明是快速范围，根据rangetype 获取当前开始结束时间戳
        const { startTime, endTime } = getTimestampRange(urlTimeRange.rangeType)
        // updateStateStartTime(startTime)
        // updateStateEndTime(endTime)
        setStartTime(startTime)
        setEndTime(endTime)

        needLoadFilter = true
      } else {
        //说明可能是精确时间，先判断是不是可以转化成微妙时间戳
        const startTimestamp = ISOToTimestamp(urlFrom)
        const endTimestamp = ISOToTimestamp(urlTo)
        if (startTimestamp && endTimestamp) {
          setStartTime(startTimestamp)
          setEndTime(endTimestamp)
          needLoadFilter = true
        }
      }
    }
    if (needLoadFilter) setIsLoaded(false)
  }, [searchParams])

  useEffect(() => {
    // console.log('本地的', isLoaded, selectServiceName, selectInstance, inputTraceId, inputEndpoint)
    updateUrlParamsState({
      startTime,
      endTime,
      service: selectServiceName,
      instance: selectInstance,
      traceId: inputTraceId,
      endpoint: inputEndpoint,
      filtersLoaded: isLoaded,
    })
  }, [isLoaded, selectServiceName, selectInstance, inputTraceId, inputEndpoint])

  useEffect(() => {
    if (startTime && endTime) {
      // console.log('时间更新了', startTime, endTime, selectServiceName, selectInstance)
      setIsLoaded(false)
      getServiceListData()
    }
  }, [startTime, endTime])

  useEffect(() => {
    // console.log(selectServiceName, startTime, endTime)
    if (selectServiceName) {
      if (selectInstance) {
        setIsLoaded(false)
      }
      if (startTime && endTime) getInstanceListData()
    }
  }, [selectServiceName, startTime, endTime])
  useEffect(
    () => () => {
      clearUrlParamsState()
    },
    [],
  )
  return (
    <div className="flex flex-row my-2 justify-between">
      <div className="flex flex-row">
        <div className="flex flex-row items-center mr-5">
          <span className="text-nowrap">服务名：</span>
          <CustomSelect
            options={serviceList}
            value={selectServiceName}
            onChange={onChangeService}
            isClearable
          />
        </div>
        <div className="flex flex-row items-center mr-5">
          <span className="text-nowrap">实例名：</span>
          <CustomSelect
            options={Object.keys(instanceList)}
            value={selectInstance}
            onChange={onChangeInstance}
            isClearable
          />
        </div>
        <div className="flex flex-row items-center mr-5">
          <span className="text-nowrap">TraceId：</span>
          <CFormInput
            size="sm"
            value={inputTraceId}
            onChange={onChangeTraceId}
          />
        </div>
        {type === 'trace' && (
          <div className="flex flex-row items-center mr-5">
            <span className="text-nowrap">服务端点：</span>
            <CFormInput
              size="sm"
              value={inputEndpoint}
              onChange={onChangeEndpoint}
            />
          </div>
        )}
      </div>
      <DateTimeRangePickerCom type={type} />
    </div>
  )
})

export default LogsTraceFilter
