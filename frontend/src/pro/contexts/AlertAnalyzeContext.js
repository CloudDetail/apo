import React, { createContext, useContext, useEffect, useMemo, useReducer, useState } from 'react'
import { useSelector } from 'react-redux'
import { useDebounce, useUpdateEffect } from 'react-use'
import {
  getAlertDescendantApi,
  getAlertDescendantDeltaApi,
  getAlertEventImpactApi,
  getAlertDescendantTrendApi,
} from 'src/pro/api/alertAnalyze'
import { getServiceAlertEventsApi, getServiceRelationApi } from 'core/api/service'
import {
  selectProcessedTimeRange,
  selectSecondsTimeRange,
} from 'src/core/store/reducers/timeRangeReducer'
import { getStep } from 'src/core/utils/step'
import _ from 'lodash'
export const alertAnalyzeInitialState = {
  loading: false,
  alertEvents: [],
  pagination: {
    pageIndex: 1,
    pageSize: 10,
    total: 0,
  },
  selectedEventId: '',
  eventImpactList: [],
  //用于打开分析弹窗的业务入口
  selectedEndpointInfo: {
    service: '',
    endpoint: '',
  },
  //用于拓扑图上选择的节点
  selectedNodeInfo: {
    service: '',
    endpoint: '',
  },
  topologyData: null,
  //当前展示的异常信息类型
  anormalTypes: ['app', 'container', 'infra', 'network', 'error'],
  //用以时间区域图数据 [{timestamp:'', count:''}]
  timeLineChartData: [],
  //用以拓扑图节点信息map
  //{service+endpointkey:{
  // timestamp:{
  //   type:info }}}
  //当前的
  topologyNodesMap: null,
  //存储处理过的数据
  finalTopologyNodesMap: null,
  originTopologyNodesMap: null,

  //设置播放的当前时间:
  marklineTimestamp: 0,
  //设置切片时间：
  deltaStartTime: null,
  deltaEndTime: null,

  //是否播放
  isPlaying: false,
}
const AlertAnalyzeContext = createContext(alertAnalyzeInitialState)
function escapeId(id) {
  return id.replace(/[^a-zA-Z0-9-_]/g, '_')
}
function contactServiceEndpoint(service, endpoint) {
  return escapeId(service + '-' + endpoint)
}
export const useAlertAnalyzeContext = () => useContext(AlertAnalyzeContext)

export const AlertAnalyzeProvider = ({ children }) => {
  const { startTime, endTime } = useSelector(selectProcessedTimeRange)
  const { startTime: stepStartTime, endTime: stepEndTime } = useSelector(selectSecondsTimeRange)
  //state
  const [loading, setLoading] = useState(false)
  const [alertEvents, setAlertEvents] = useState([])
  const [pagination, setPagination] = useState({
    pageIndex: 1,
    pageSize: 10,
    total: 0,
  })
  const [eventImpactList, setEventImpactList] = useState([])
  //节点层级
  const [topologyLevels, setTopologyLevels] = useState([])
  //用于下钻分析的service和endpoint
  const [selectedEndpointInfo, setSelectedEndpointInfo] = useState({
    service: '',
    endpoint: '',
  })
  //用于拓扑图上选择的节点
  const [selectedNodeInfo, setSelectedNodeInfo] = useState({
    service: '',
    endpoint: '',
  })
  const [topologyData, setTopologyData] = useState(null)
  //当前选择的异常信息类型
  const [anormalTypes, setAnormalTypes] = useState([
    'app',
    'container',
    'infra',
    'network',
    'error',
    'appInstance',
  ])
  const [timeLineChartData, setTimeLineChartData] = useState([])
  //当前的拓扑图节点信息map
  const [topologyNodesMap, setTopologyNodesMap] = useState(null)
  //存储处理过的数据前、后拓扑图节点信息map
  const [finalTopologyNodesMap, setFinalTopologyNodesMap] = useState(null)
  const [midwayTopologyNodesMap, setMidwayTopologyNodesMap] = useState(null)
  const [originTopologyNodesMap, setOriginTopologyNodesMap] = useState(null)

  // 切片最终告警事件列表
  const [finalAnormalEvents, setFinalAnormalEvents] = useState([])
  // 切片变化列表
  const [deltaAnormalEvents, setDeltaAnormalEvents] = useState([])
  //设置播放的当前时间点:
  const [marklineTimestamp, setMarklineTimestamp] = useState(0)
  //设置切片时间：
  const [deltaStartTime, setDeltaStartTime] = useState(0)
  const [deltaEndTime, setDeltaEndTime] = useState(0)
  const [isPlaying, setIsPlaying] = useState(false)
  // 是否按过一键分析了
  const [isAnalyze, setIsAnalyze] = useState(false)

  const anormalTypesOptions = [
    {
      value: 'unknown',
      key: 0,
      label: '未知',
      hide: true,
    },
    {
      value: 'app',
      key: 1,
      label: '应用接口异常类型',
    },
    {
      value: 'container',
      key: 2,
      label: '容器异常类型',
    },
    {
      value: 'infra',
      key: 3,
      label: '基础设施异常类型',
    },
    {
      value: 'network',
      key: 4,
      label: '网络异常类型',
    },
    {
      value: 'error',
      key: 5,
      label: 'JAVA Exception',
    },
    {
      value: 'appInstance',
      key: 6,
      label: '应用实例异常类型',
    },
  ]

  const prepareTopologyNode = (data) => {
    let topologyNodesMap = {}
    data?.map((node) => {
      let nodeId = contactServiceEndpoint(node.serviceName, node.endpoint)
      // 初始化 topologyNodesMap[nodeId] 如果它还不存在
      if (!topologyNodesMap[nodeId]) {
        topologyNodesMap[nodeId] = {}
      }
      topologyNodesMap[nodeId] = node.anormalCounts
    })

    return topologyNodesMap
  }

  // 获取告警列表
  const fetchData = async () => {
    setLoading(true)
    try {
      const params = {
        startTime: startTime,
        endTime: endTime,
        currentPage: pagination.pageIndex,
        pageSize: pagination.pageSize,
        status: 'firing',
        // service: 'ts-gateway-service',
      }

      const [res] = await Promise.all([
        getServiceAlertEventsApi(params),
        // getLogRuleApi({ tableName: 'test_logs', dataBase: 'default' }),
      ])
      setAlertEvents(res.events || [])
      setPagination({
        pageIndex: pagination.pageIndex,
        pageSize: pagination.pageSize,
        total: res.pagination.total,
      })
    } catch (error) {
      console.error('请求出错:', error)

      setAlertEvents([])
      setPagination({
        pageIndex: pagination.pageIndex,
        pageSize: pagination.pageSize,
        total: res.pagination.total,
      })
    } finally {
      setLoading(false)
    }
  }
  // 获取告警的影响面分析
  const getAlertEventImpact = (onlyUpdateEvents = false) => {
    setLoading(true)
    const step = getStep(stepStartTime, stepEndTime)
    getAlertEventImpactApi({
      startTime: stepStartTime,
      endTime: stepEndTime,
      step,
      currentPage: pagination.pageIndex,
      pageSize: pagination.pageSize,
    })
      .then((res) => {
        if (!onlyUpdateEvents) setEventImpactList(res.data || [])
        setAlertEvents(res.alertEvents || [])
      })
      .catch((error) => {
        if (!onlyUpdateEvents) setEventImpactList([])
        setAlertEvents([])
      })
      .finally(() => {
        setLoading(false)
      })
  }
  // 获取拓扑图
  const getServiceRelation = async () => {
    getServiceRelationApi({
      startTime,
      endTime,
      service: selectedEndpointInfo.service,
      endpoint: selectedEndpointInfo.endpoint,
      entryService: selectedEndpointInfo.service,
      entryEndpoint: selectedEndpointInfo.endpoint,
      withTopologyLevel: true,
      removeClientCall: true,
    }).then((res) => {
      setTopologyData(res)
      let result = []
      res.topologyLevels.map((node) => {
        if (!result[node.depth]) {
          result[node.depth] = []
        }
        result[node.depth].push(contactServiceEndpoint(node.service, node.endpoint))
      })
      setTopologyLevels(result)
    })
  }

  const initAlertDescendantTrend = async () => {
    const step = getStep(stepStartTime, stepEndTime)
    getAlertDescendantTrendApi({
      startTime: stepStartTime,
      endTime: stepEndTime,
      service: selectedEndpointInfo.service,
      endpoint: selectedEndpointInfo.endpoint,
      step,
      anormalTypes: anormalTypes.join(','),
    }).then((res) => {
      const chartData = res.anormalCount?.chartData || []
      setTimeLineChartData(chartData)
      let maxcountTimestamp = Number(
        Object.keys(chartData)
          .sort((a, b) => b - a) // 按降序排序
          .reduce((max, key) => (chartData[key] > chartData[max] ? key : max)),
      )
      if (maxcountTimestamp - step < stepStartTime) {
        maxcountTimestamp = stepStartTime + step
      }
      setDeltaStartTime(maxcountTimestamp - step)
      setDeltaEndTime(maxcountTimestamp)
      setMarklineTimestamp(maxcountTimestamp)
    })
  }
  const initAlertDescendantDelta = async () => {
    const step = getStep(stepStartTime, stepEndTime)
    setLoading(true)
    getAlertDescendantDeltaApi({
      startTime: stepStartTime,
      endTime: stepEndTime,
      service: selectedEndpointInfo.service,
      endpoint: selectedEndpointInfo.endpoint,
      deltaStartTime: stepEndTime - step,
      deltaEndTime: stepEndTime,
      step,
      anormalTypes: anormalTypes.join(','),
    })
      .then((res) => {
        //默认初始化，用步长最后一个时间段作为切片

        const finalTopologyNodesMap = prepareTopologyNode(res.finalAnormalCounts)
        const midwayTopologyNodesMap = prepareTopologyNode(res.alertTriggeredCounts)
        const originTopologyNodesMap = prepareTopologyNode(res.originAnormalCounts)

        setTopologyNodesMap(finalTopologyNodesMap)

        setOriginTopologyNodesMap(originTopologyNodesMap)
        setMidwayTopologyNodesMap(midwayTopologyNodesMap)
        setFinalTopologyNodesMap(finalTopologyNodesMap)
        setFinalAnormalEvents(res.finalAnormalEvents || [])
        setDeltaAnormalEvents(res.deltaAnormalEvents || [])
      })
      .catch(() => {
        setTopologyNodesMap(null)
        setOriginTopologyNodesMap(null)
        setMidwayTopologyNodesMap(null)
        setFinalTopologyNodesMap(null)
        setFinalAnormalEvents([])
        setDeltaAnormalEvents([])
      })
      .finally(() => {
        setLoading(false)
      })
  }
  //仅改变切片时间或筛选类型
  const changeAlertDescendantDelta = _.debounce(({ deltaStartTime, deltaEndTime }) => {
    //默认初始化，用步长最后一个时间段作为切片
    setDeltaStartTime(deltaStartTime)
    setDeltaEndTime(deltaEndTime)
    setLoading(true)
    setMarklineTimestamp(deltaEndTime)
    const step = getStep(stepStartTime, stepEndTime)
    if (selectedEndpointInfo.service && selectedEndpointInfo.endpoint) {
      getAlertDescendantDeltaApi({
        startTime: stepStartTime,
        endTime: stepEndTime,
        service: selectedEndpointInfo.service,
        endpoint: selectedEndpointInfo.endpoint,
        deltaStartTime: deltaStartTime,
        deltaEndTime: deltaEndTime,
        step,
        anormalTypes: anormalTypes.join(','),
      })
        .then((res) => {
          const finalTopologyNodesMap = prepareTopologyNode(res.finalAnormalCounts)
          const midwayTopologyNodesMap = prepareTopologyNode(res.alertTriggeredCounts)
          const originTopologyNodesMap = prepareTopologyNode(res.originAnormalCounts)
          setTopologyNodesMap(finalTopologyNodesMap)
          setOriginTopologyNodesMap(originTopologyNodesMap)
          setMidwayTopologyNodesMap(midwayTopologyNodesMap)
          setFinalTopologyNodesMap(finalTopologyNodesMap)
          setFinalAnormalEvents(res.finalAnormalEvents)
          setDeltaAnormalEvents(res.deltaAnormalEvents || [])
        })
        .catch(() => {
          setTopologyNodesMap(null)
          setOriginTopologyNodesMap(null)
          setFinalTopologyNodesMap(null)
          setFinalAnormalEvents([])
          setDeltaAnormalEvents([])
        })
        .finally(() => {
          setLoading(false)
        })
    }
  }, 500)
  //仅类型变更导致的图变更
  const updateAlertDescendantTrend = async () => {
    const step = getStep(stepStartTime, stepEndTime)
    getAlertDescendantTrendApi({
      startTime: stepStartTime,
      endTime: stepEndTime,
      service: selectedEndpointInfo.service,
      endpoint: selectedEndpointInfo.endpoint,
      step,
      anormalTypes: anormalTypes.join(','),
    }).then((res) => {
      const chartData = res.anormalCount?.chartData || []
      setTimeLineChartData(chartData)
    })
  }

  const handlePlay = () => {
    let currentTimestamp = deltaStartTime
    setMarklineTimestamp(currentTimestamp)
    let index = topologyLevels.length - 1
    let step = (deltaEndTime - deltaStartTime) / topologyLevels.length
    let originMap = { ...originTopologyNodesMap }
    let midMap = midwayTopologyNodesMap
    let finalMap = finalTopologyNodesMap
    setTopologyNodesMap({ ...originMap })

    setTimeout(() => {
      setIsPlaying(true)

      const playToMidInterval = setInterval(() => {
        if (index < 0) {
          clearInterval(playToMidInterval)
          index = topologyLevels.length - 1
          // 第一个 setInterval 完成后，开始执行第二个
          const playToFinalInterval = setInterval(() => {
            if (index < 0) {
              clearInterval(playToFinalInterval)
              setIsPlaying(false)
            } else {
              topologyLevels[index].forEach((nodeId) => {
                originMap[nodeId] = finalMap[nodeId]
              })
              setTopologyNodesMap({ ...originMap })

              index -= 1
              currentTimestamp += step
              if (currentTimestamp > deltaEndTime) {
                currentTimestamp = deltaEndTime
              }
              setMarklineTimestamp(currentTimestamp)
            }
          }, 1000)
        } else {
          topologyLevels[index].forEach((nodeId) => {
            originMap[nodeId] = midMap[nodeId]
          })
          setTopologyNodesMap({ ...originMap })

          index -= 1
          currentTimestamp += step
          if (currentTimestamp > deltaEndTime) {
            currentTimestamp = deltaEndTime
          }
          setMarklineTimestamp(currentTimestamp)
        }
      }, 1000)
    }, 100)
  }

  useUpdateEffect(() => {
    if (startTime && endTime) {
      setLoading(true)
      if (isAnalyze) {
        getAlertEventImpact(true)
      } else {
        fetchData({ startTime, endTime })
      }
    }
  }, [pagination.pageIndex, pagination.pageSize])
  useEffect(() => {
    if (isAnalyze) getAlertEventImpact()
  }, [isAnalyze])
  //防抖避免跳转使用旧时间
  useDebounce(
    () => {
      if (startTime && endTime) {
        setIsAnalyze(false)
        setTopologyData(null)
        if (pagination.pageIndex === 1) {
          fetchData({ startTime, endTime })
        } else {
          setPagination({
            ...pagination,
            pageIndex: 1,
          })
        }
      }
    },
    300, // 延迟时间 300ms
    [startTime, endTime],
  )

  const updateState = async () => {
    setLoading(true)
    if (selectedEndpointInfo?.endpoint) {
      await getServiceRelation()
      await initAlertDescendantTrend()
      await initAlertDescendantDelta()
    }
    setLoading(false)
  }

  useUpdateEffect(() => {
    updateState()
  }, [selectedEndpointInfo.endpoint, selectedEndpointInfo.service, startTime, endTime])
  useUpdateEffect(() => {
    changeAlertDescendantDelta({
      deltaStartTime: deltaStartTime,
      deltaEndTime: deltaEndTime,
    })
    updateAlertDescendantTrend()
  }, [anormalTypes])
  const memoizedValue = useMemo(
    () => ({
      pagination,
      loading,
      alertEvents,
      eventImpactList,
      selectedEndpointInfo,
      topologyData,
      anormalTypes,
      timeLineChartData,
      marklineTimestamp,
      topologyNodesMap,
      selectedNodeInfo,
      deltaStartTime,
      deltaEndTime,
      isPlaying,
      anormalTypesOptions,

      finalAnormalEvents,
      deltaAnormalEvents,
      isAnalyze,

      fetchData,
      handlePlay,
      changeAlertDescendantDelta,

      updatePagination: (pagination) => setPagination({ ...pagination, ...pagination }),
      setLoading,
      setEventImpactList,
      setSelectedEndpointInfo,
      setTopologyData,
      setAnormalTypes,
      setMarklineTimestamp,
      setTopologyNodesMap,
      setSelectedNodeInfo,
      setDeltaTimerange: (timerange) => {
        setDeltaStartTime(timerange.deltaStartTime)
        setDeltaEndTime(timerange.deltaEndTime)
      },
      setIsPlaying,
      prepareTopologyNode,
      setIsAnalyze,
    }),
    [
      pagination,
      loading,
      alertEvents,
      eventImpactList,
      selectedEndpointInfo,
      topologyData,
      anormalTypes,
      timeLineChartData,
      marklineTimestamp,
      topologyNodesMap,
      selectedNodeInfo,
      deltaStartTime,
      deltaEndTime,
      isPlaying,
      finalAnormalEvents,
      deltaAnormalEvents,
      isAnalyze,
    ],
  )

  return (
    <AlertAnalyzeContext.Provider value={memoizedValue}>{children}</AlertAnalyzeContext.Provider>
  )
}
