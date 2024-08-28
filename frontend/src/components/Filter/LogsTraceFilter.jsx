import { CFormInput } from '@coreui/react'
import React, { useState, useEffect, useMemo, useRef } from 'react'
import { useSearchParams } from 'react-router-dom'
import { getServiceInstancOptionsListApi, getServiceListApi } from 'src/api/service'
import DateTimeRangePickerCom from 'src/components/DateTime/DateTimeRangePickerCom'
import { CustomSelect } from 'src/components/Select'
import { getTimestampRange, timeRangeList } from 'src/store/reducers/timeRangeReducer'
import { ISOToTimestamp } from 'src/utils/time'
import { useDispatch } from 'react-redux'
import { Segmented, Tooltip } from 'antd'
import { swTraceIDToTraceID } from 'src/utils/trace'

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
  const [inputSWTraceId, setInputSWTraceId] = useState('')
  const [convertTraceId, setConvertSWTraceId] = useState('')
  const [traceType, setTraceType] = useState('TraceId')
  const [urlParam, setUrlParam] = useState({
    service: '',
    instance: '',
    traceId: '',
    endpoint: '',
    startTime: null,
    endTime: null,
  })
  const dispatch = useDispatch()
  const clearUrlParamsState = (value) => {
    dispatch({ type: 'clearUrlParamsState', payload: value })
  }

  const updateUrlParamsState = (params) => {
    changeUrlParams(params)
    dispatch({ type: 'setUrlParamsState', payload: params })
  }
  const onChangeService = (props) => {
    setSelectServiceName(props)
    if (!props) {
      setSelectInstance('')
      setInstanceList([])
      updateUrlParamsState({
        service: '',
        instance: '',
      })
    }
  }
  const onChangeInstance = (props) => {
    setSelectInstance(props)
    updateUrlParamsState({
      service: selectServiceName,
      instance: props,
    })
  }
  const onChangeTraceId = (event) => {
    setInputTraceId(event.target.value)
    updateUrlParamsState({
      traceId: event.target.value,
    })
  }
  const onChangeSWTraceId = (event) => {
    setInputSWTraceId(event.target.value)
    const convertTraceId = swTraceIDToTraceID(event.target.value)
    setConvertSWTraceId(convertTraceId)
    setInputTraceId(convertTraceId)
    updateUrlParamsState({
      traceId: convertTraceId,
    })
  }
  const onChangeEndpoint = (event) => {
    setInputEndpoint(event.target.value)
    updateUrlParamsState({
      endpoint: event.target.value,
    })
  }
  const getServiceListData = () => {
    getServiceListApi({ startTime, endTime })
      .then((res) => {
        setServiceList(res ?? [])

        let storeService = selectServiceName
        let storeInstance = selectInstance
        if (res.includes(urlParam.service)) {
          setSelectServiceName(urlParam.service)
          storeService = urlParam.service
        } else {
          setSelectServiceName('')
          setSelectInstance('')
          storeService = ''
          storeInstance = ''
        }

        if (!storeService) {
          updateUrlParamsState({
            startTime,
            endTime,
            service: storeService,
            instance: storeInstance,
            traceId: inputTraceId,
            endpoint: inputEndpoint,
          })
        } else if (selectServiceName === storeService) {
          getInstanceListData()
          return
        }
      })
      .catch((error) => {
        // console.log(error)
        setServiceList([])
      })
  }
  const getInstanceListData = () => {
    if (selectServiceName) {
      getServiceInstancOptionsListApi({
        startTime,
        endTime,
        service: selectServiceName,
      })
        .then((res) => {
          setInstanceList(res)
          // updateInstanceOption(res)
          let storeInstance = selectInstance
          if (res[urlParam.instance]) {
            setSelectInstance(urlParam.instance)
            storeInstance = urlParam.instance
          } else {
            setSelectInstance('')
            storeInstance = ''
          }
          updateUrlParamsState({
            startTime,
            endTime,
            service: selectServiceName,
            instance: storeInstance,
            traceId: inputTraceId,
            endpoint: inputEndpoint,
            instanceOption: res,
          })
        })
        .catch((error) => {
          // console.log(error)
          setInstanceList(null)
          // updateInstanceOption({})
        })
        .finally(() => {})
    }
  }
  useEffect(() => {
    const urlService = searchParams.get('service') ?? ''
    const urlInstance = searchParams.get('instance') ?? ''
    const urlTraceId = searchParams.get('traceId') ?? ''
    const urlEndpoint = searchParams.get('endpoint') ?? ''
    const urlFrom = searchParams.get(type + '-from')
    const urlTo = searchParams.get(type + '-to')
    console.log(
      'url参数改变',
      urlParam.service,
      urlService,
      urlInstance,
      urlTraceId,
      urlEndpoint,
      urlFrom,
      urlTo,
    )

    if (urlFrom && urlTo) {
      const urlTimeRange = timeRangeList.find((item) => item.from === urlFrom && item.to === urlTo)
      if (urlTimeRange) {
        //说明是快速范围，根据rangetype 获取当前开始结束时间戳
        const { startTime, endTime } = getTimestampRange(urlTimeRange.rangeType)
        // updateStateStartTime(startTime)
        // updateStateEndTime(endTime)
        setStartTime(startTime)
        setEndTime(endTime)
        // urlParam.startTime = startTime
        // urlParam.endTime = endTime
      } else {
        //说明可能是精确时间，先判断是不是可以转化成微妙时间戳
        const startTimestamp = ISOToTimestamp(urlFrom)
        const endTimestamp = ISOToTimestamp(urlTo)
        if (startTimestamp && endTimestamp) {
          setStartTime(startTimestamp)
          setEndTime(endTimestamp)
          // urlParam.startTime = startTimestamp
          // urlParam.endTime = endTimestamp
        }
      }
    }
    setInputTraceId(urlTraceId)
    setInputEndpoint(urlEndpoint)
    setUrlParam({
      ...urlParam,
      service: urlService,
      instance: urlInstance,
      traceId: urlTraceId,
      endpoint: urlEndpoint,
    })
  }, [searchParams])
  useEffect(() => {
    if (selectServiceName) {
      getInstanceListData()
    } else {
      // setInstanceList([])
      // onChangeInstance('')
    }
  }, [selectServiceName])
  useEffect(
    () => () => {
      clearUrlParamsState()
    },
    [],
  )
  useEffect(() => {
    let changeTime = false
    if (startTime !== urlParam.startTime) {
      changeTime = true
      urlParam.startTime = startTime
    }
    if (endTime !== urlParam.endTime) {
      changeTime = true
      urlParam.endTime = endTime
    }
    if (startTime && endTime) {
      // console.log(urlParam.service, selectServiceName)
      if (changeTime || urlParam.service !== selectServiceName) {
        getServiceListData()
      }
    }
  }, [startTime, endTime, urlParam])
  const changeUrlParams = (props) => {
    // console.log(props, urlParam)
    // const { service: storeService, instance: storeInstance } = props
    const params = new URLSearchParams(searchParams)
    let needChangeUrl = false
    if ('service' in props && props.service !== urlParam.service) {
      params.set('service', props.service || '')
      needChangeUrl = true
    }

    if ('instance' in props && props.instance !== urlParam.instance) {
      params.set('instance', props.instance || '')
      needChangeUrl = true
    }

    if ('endpoint' in props && props.endpoint !== urlParam.endpoint) {
      params.set('endpoint', props.endpoint || '')
      needChangeUrl = true
    }

    if ('traceId' in props && props.traceId !== urlParam.traceId) {
      params.set('traceId', props.traceId || '')
      needChangeUrl = true
    }
    // // console.log(props)
    // // console.log(service,instance)
    if (needChangeUrl) {
      // // console.log('state改变url')
      setSearchParams(params, { replace: true })
    }
  }
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
          {type === 'trace' ? (
            <Segmented options={['TraceId', 'SWTraceId']} onChange={setTraceType} />
          ) : (
            <span className="text-nowrap">TraceId：</span>
          )}
          ：
          {traceType === 'TraceId' ? (
            <CFormInput size="sm" value={inputTraceId} onChange={onChangeTraceId} />
          ) : (
            <Tooltip
              title={
                convertTraceId
                  ? '自动转换为TraceID：' + convertTraceId
                  : '输入SkyWalking的traceid将自动转换'
              }
            >
              <CFormInput size="sm" value={inputSWTraceId} onChange={onChangeSWTraceId} />
            </Tooltip>
          )}
        </div>
        {type === 'trace' && (
          <div className="flex flex-row items-center mr-5">
            <span className="text-nowrap">服务端点：</span>
            <CFormInput size="sm" value={inputEndpoint} onChange={onChangeEndpoint} />
          </div>
        )}
      </div>
      <DateTimeRangePickerCom type={type} />
    </div>
  )
})

export default LogsTraceFilter
