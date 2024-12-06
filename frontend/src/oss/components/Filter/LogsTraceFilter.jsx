import React, { useState, useEffect } from 'react'
import { useSearchParams } from 'react-router-dom'
import { getServiceInstancOptionsListApi, getServiceListApi, getNamespacesApi } from 'core/api/service'
import DateTimeRangePickerCom from 'src/core/components/DateTime/DateTimeRangePickerCom'
import { CustomSelect } from 'src/core/components/Select'
import { getTimestampRange, timeRangeList } from 'src/core/store/reducers/timeRangeReducer'
import { ISOToTimestamp } from 'src/core/utils/time'
import { useDispatch } from 'react-redux'
import { Checkbox, Input, InputNumber, Segmented, Tooltip, Select } from 'antd'
import { swTraceIDToTraceID } from 'src/core/utils/trace'
import TraceErrorType from 'src/oss/views/trace/component/TraceErrorType'

const LogsTraceFilter = React.memo(({ type }) => {
  const [searchParams, setSearchParams] = useSearchParams()

  const [serviceList, setServiceList] = useState([])
  const [instanceList, setInstanceList] = useState([])
  const [namespaceList, setNamespaceList] = useState([])

  const [selectServiceName, setSelectServiceName] = useState('')
  const [selectInstance, setSelectInstance] = useState('')
  const [selectNamespace, setSelectNamespace] = useState(null)
  // 应该深入
  const [inputTraceId, setInputTraceId] = useState('')
  const [inputEndpoint, setInputEndpoint] = useState('')
  const [startTime, setStartTime] = useState(null)
  const [endTime, setEndTime] = useState(null)
  const [inputSWTraceId, setInputSWTraceId] = useState('')
  const [convertTraceId, setConvertSWTraceId] = useState('')
  // filter
  const [minDuration, setMinDuration] = useState(null)
  const [maxDuration, setMaxDuration] = useState(null)
  const [faultTypeList, setFaultTypeList] = useState([])
  const [traceType, setTraceType] = useState('TraceID')
  const [urlParam, setUrlParam] = useState({
    service: '',
    instance: '',
    traceId: '',
    endpoint: '',
    startTime: null,
    endTime: null,

    //filter
    namespace: '',
    minDuration: null,
    maxDuration: null,
    faultTypeList: null,
  })
  const options = [
    { label: <TraceErrorType type="slow" />, value: 'slow' },
    { label: <TraceErrorType type="error" />, value: 'error' },
    { label: <TraceErrorType type="normal" />, value: 'normal' },
  ]
  //trace more filter
  const [visible, setVisible] = useState(true)
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
  const onChangeMinDuration = (value) => {
    setMinDuration(value)
    updateUrlParamsState({
      minDuration: value,
    })
  }
  const onChangeMaxDuration = (value) => {
    setMaxDuration(value)
    updateUrlParamsState({
      maxDuration: value,
    })
  }
  const onChangeTypeList = (value) => {
    setFaultTypeList(value)
    updateUrlParamsState({
      faultTypeList: value,
    })
  }
  const onChangeNamespace = (event) => {
    setSelectNamespace(event)
    updateUrlParamsState({
      namespace: event,
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
            namespace:selectNamespace,
            minDuration,
            maxDuration,
          })
        } else if (selectServiceName === storeService) {
          getInstanceListData()
          return
        }
      })
      .catch((error) => {
        console.log(error)
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
            namespace: selectNamespace
          })
        })
        .catch((error) => {
          // console.log(error)
          setInstanceList(null)
          // updateInstanceOption({})
        })
        .finally(() => { })
    }
  }
  const getNamespaceList = () => {
    getNamespacesApi()
      .then((res) => {
        setNamespaceList(res.items.map((item) => ({
          label: <span>{item.metadata.name}</span>,
          value: item.metadata.name
        })))
      })
      .catch((error) => {
        console.error(error)
      })
      .finally(() => { })
  }
  useEffect(() => {
    const urlService = searchParams.get('service') ?? ''
    const urlInstance = searchParams.get('instance') ?? ''
    const urlTraceId = searchParams.get('traceId') ?? ''
    const urlEndpoint = searchParams.get('endpoint') ?? ''
    const urlFrom = searchParams.get(type + '-from')
    const urlTo = searchParams.get(type + '-to')
    const namespace = searchParams.get('namespace') == '' ? null : searchParams.get('namespace') ?? null
    const minDuration = searchParams.get('minDuration') ?? ''
    const maxDuration = searchParams.get('maxDuration') ?? ''
    const faultTypeList = searchParams.get('faultTypeList') ?? ''
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
    let faultTypeListValue = faultTypeList ? faultTypeList.split(',') : null
    setMinDuration(minDuration)
    setMaxDuration(maxDuration)
    setSelectNamespace(namespace)
    setFaultTypeList(faultTypeListValue)
    setUrlParam({
      ...urlParam,
      service: urlService,
      instance: urlInstance,
      traceId: urlTraceId,
      endpoint: urlEndpoint,
      namespace,
      minDuration,
      maxDuration,
      faultTypeList: faultTypeListValue,
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
  useEffect(() => {
    clearUrlParamsState()
  }, [])

  useEffect(() => {
    getNamespaceList()
  }, [])

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
    if ('namespace' in props && props.namespace !== urlParam.namespace) {
      params.set('namespace', props.namespace || '')
      needChangeUrl = true
    }
    if ('minDuration' in props && props.minDuration !== urlParam.minDuration) {
      params.set('minDuration', props.minDuration || '')
      needChangeUrl = true
    }
    if ('maxDuration' in props && props.maxDuration !== urlParam.maxDuration) {
      params.set('maxDuration', props.maxDuration || '')
      needChangeUrl = true
    }

    if ('faultTypeList' in props && props.faultTypeList.join(',') !== urlParam.faultTypeList) {
      params.set('faultTypeList', props.faultTypeList.join(','))
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
    <>
      <div className="flex flex-row my-2 justify-between">
        <div className="flex flex-row  flex-wrap">
          <div className="flex flex-row items-center mr-5 mt-2 min-w-[200px]">
            <span className="text-nowrap">命名空间：</span>
            <Select
              className='w-full'
              options={namespaceList}
              value={selectNamespace}
              placeholder="检索"
              onChange={onChangeNamespace}
              allowClear
              popupMatchSelectWidth={false}
            />
          </div>
          <div className="flex flex-row items-center mr-5 mt-2 w-[250px]">
            <span className="text-nowrap w-[60px]">服务名：</span>
            <div className="flex-1 w-0">
              <CustomSelect
                options={serviceList}
                value={selectServiceName}
                onChange={onChangeService}
                isClearable
              />
            </div>
          </div>
          <div className="flex flex-row items-center mr-5 mt-2 min-w-[200px]">
            <span className="text-nowrap">实例名：</span>
            <div className="flex-1">
              <CustomSelect
                options={Object.keys(instanceList)}
                value={selectInstance}
                onChange={onChangeInstance}
                isClearable
              />
            </div>
          </div>
        </div>
        <div className="flex-grow-0 flex-shrink-0 flex">
          <DateTimeRangePickerCom type={type} />
          {type === 'trace' && (
            <div
              onClick={() => setVisible(!visible)}
              className="flex flex-row items-center cursor-pointer"
            >
              {/* <span className=" font-bold mr-2">更多筛选器</span> <BsChevronDoubleDown size={20} /> */}
            </div>
          )}
        </div>
      </div>
      {type === 'trace' && (
        <>
          <div className="text-xs flex flex-row  flex-wrap w-full mb-2">
            <div className="flex flex-row items-center mr-5 mt-2">
              <span className="text-nowrap">响应时间：</span>
              <div className="flex-1 flex flex-row items-center">
                <div className="pr-2">
                  <InputNumber
                    addonBefore="MIN"
                    addonAfter="ms"
                    min={0}
                    value={minDuration}
                    onChange={onChangeMinDuration}
                    className=" w-[150px]"
                  />
                </div>
                至
                <div className="pl-2">
                  <InputNumber
                    addonBefore="MAX"
                    addonAfter="ms"
                    min={0}
                    value={maxDuration}
                    onChange={onChangeMaxDuration}
                    className=" w-[150px]"
                  />
                </div>
              </div>
            </div>
            {type === 'trace' && (
              <div className="flex flex-row items-center mr-5 mt-2 w-[150px]">
                <span className="text-nowrap ">服务端点：</span>
                <Input placeholder="检索" value={inputEndpoint} onChange={onChangeEndpoint} />
              </div>
            )}
          </div>
          <div className='flex'>
            <div className="flex flex-row items-center mr-5 mt-2 w-[300px] text-sm">
              {type === 'trace' ? (
                <Segmented options={['TraceID', 'SWTraceId']} onChange={setTraceType} />
              ) : (
                <span className="text-nowrap text-sm">TraceId：</span>
              )}
              ：
              {traceType === 'TraceID' ? (
                <Input placeholder="检索" value={inputTraceId} onChange={onChangeTraceId} />
              ) : (
                <Tooltip
                  title={
                    convertTraceId
                      ? '自动转换为TraceID：' + convertTraceId
                      : '输入SkyWalking的traceid将自动转换'
                  }
                >
                  <Input placeholder="检索" value={inputSWTraceId} onChange={onChangeSWTraceId} />
                </Tooltip>
              )}
            </div>
            <div className="flex flex-row items-center mr-5 mt-2">
              <span className="text-nowrap">故障状态：</span>
              <Checkbox.Group
                onChange={onChangeTypeList}
                options={options}
                value={faultTypeList}
              ></Checkbox.Group>
            </div>
          </div>
        </>
      )}
    </>
  )
})

export default LogsTraceFilter
