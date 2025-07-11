import React, { createContext, ReactNode, useContext, useEffect, useMemo, useState } from 'react'
import { useSelector } from 'react-redux'
import { useDebounce, useUpdateEffect } from 'react-use'
import { getAlertDescendantDeltaApi, getAlertDescendantTrendApi } from 'src/pro/api/alertAnalyze'
import { getServiceRelationApi } from 'core/api/service'
import {
  selectProcessedTimeRange,
  selectSecondsTimeRange,
} from 'src/core/store/reducers/timeRangeReducer'
import { getStep } from 'src/core/utils/step'
import _ from 'lodash'
import { anormalTypes as initAnormalTypes, anormalTypesOptions } from '../constants'
import { useSearchParams } from 'react-router-dom'
interface TopologyNode {
  depth: number
  endpoint: string
  service: string
}
type AnormalCountsMap = {
  [typeValue: string]: number
}

type TopologyNodeInfoMap = {
  [serviceEndpointKey: string]: AnormalCountsMap | undefined
}
interface AnormalCounts {
  anormalCounts: AnormalCountsMap
  endpoint: string
  serviceName: string
}

export interface SelectedNodeInfo {
  service: string
  endpoint: string
}
export interface TimelinePoint {
  timestamp: string
  count: string
}
export interface ServiceAnalyzeTypes extends SelectedNodeInfo {
  loading: boolean
  selectedEventId: string
  selectedNodeInfo: SelectedNodeInfo
  topologyData: any
  anormalTypes: string[]
  timeLineChartData: TimelinePoint[]
  topologyNodesMap: TopologyNodeInfoMap | null
  finalTopologyNodesMap: TopologyNodeInfoMap | null
  originTopologyNodesMap: TopologyNodeInfoMap | null
  marklineTimestamp: number
  deltaStartTime: number | null
  deltaEndTime: number | null
  isPlaying: boolean
}
export const serviceAnalyzeInitialState: ServiceAnalyzeTypes = {
  loading: false,
  selectedEventId: '',
  //Topology Selected Node Info
  selectedNodeInfo: {
    service: '',
    endpoint: '',
  },
  service: '',
  endpoint: '',
  topologyData: null,
  anormalTypes: initAnormalTypes,
  // [{timestamp:'', count:''}]
  timeLineChartData: [],
  //Topology Node Info map
  //{service+endpointkey:{
  // timestamp:{
  //   type:info }}}
  //current
  topologyNodesMap: null,
  //store processed data
  finalTopologyNodesMap: null,
  originTopologyNodesMap: null,

  //play start time
  marklineTimestamp: 0,
  //delta timerange：
  deltaStartTime: null,
  deltaEndTime: null,
  isPlaying: false,
}
const ServiceAnalyzeContext = createContext<ServiceAnalyzeTypes>(serviceAnalyzeInitialState)
function escapeId(id: string) {
  return id.replace(/[^a-zA-Z0-9-_]/g, '_')
}
function contactServiceEndpoint(service: string, endpoint: string) {
  return escapeId(service + '-' + endpoint)
}
export const useServiceAnalyzeContext = () => useContext(ServiceAnalyzeContext)

export const ServiceAnalyzeProvider: React.FC<{ children: ReactNode }> = ({ children }) => {
  const { startTime, endTime } = useSelector(selectProcessedTimeRange)
  const { startTime: stepStartTime, endTime: stepEndTime } = useSelector(selectSecondsTimeRange)
  const [loading, setLoading] = useState(false)
  const [topologyLevels, setTopologyLevels] = useState<string[][]>([])
  const [selectedNodeInfo, setSelectedNodeInfo] = useState({
    service: '',
    endpoint: '',
  })
  const [topologyData, setTopologyData] = useState(null)
  const [anormalTypes, setAnormalTypes] = useState(initAnormalTypes)
  const [timeLineChartData, setTimeLineChartData] = useState([])
  const [topologyNodesMap, setTopologyNodesMap] = useState<TopologyNodeInfoMap | null>(null)
  const [finalTopologyNodesMap, setFinalTopologyNodesMap] = useState<TopologyNodeInfoMap | null>(
    null,
  )
  const [midwayTopologyNodesMap, setMidwayTopologyNodesMap] = useState<TopologyNodeInfoMap | null>(
    null,
  )
  const [originTopologyNodesMap, setOriginTopologyNodesMap] = useState<TopologyNodeInfoMap | null>(
    null,
  )

  const [finalAnormalEvents, setFinalAnormalEvents] = useState<any[]>([])
  const [deltaAnormalEvents, setDeltaAnormalEvents] = useState<any[]>([])
  const [marklineTimestamp, setMarklineTimestamp] = useState(0)
  const [deltaStartTime, setDeltaStartTime] = useState(0)
  const [deltaEndTime, setDeltaEndTime] = useState(0)
  const [isPlaying, setIsPlaying] = useState(false)
  const [searchParams] = useSearchParams()
  const [service, setService] = useState<string | null>(null)
  const [endpoint, setEndpoint] = useState<string | null>(null)
  const [clusterIds, setClusterIds] = useState<string[]>([])
  const { dataGroupId } = useSelector((state) => state.dataGroupReducer)
  const prepareTopologyNode = (data: AnormalCounts[]) => {
    let topologyNodesMap: TopologyNodeInfoMap = {}
    data?.map((node: AnormalCounts) => {
      let nodeId: string = contactServiceEndpoint(node.serviceName, node.endpoint)
      // init topologyNodesMap[nodeId] when is undefind
      if (!topologyNodesMap[nodeId]) {
        topologyNodesMap[nodeId] = {}
      }
      topologyNodesMap[nodeId] = node.anormalCounts
    })

    return topologyNodesMap
  }

  // 获取拓扑图
  const getServiceRelation = async () => {
    const res = await getServiceRelationApi({
      startTime,
      endTime,
      service: service!,
      endpoint: endpoint!,
      entryService: service!,
      entryEndpoint: endpoint!,
      withTopologyLevel: true,
      removeClientCall: true,
      clusterIds,
      groupId: dataGroupId,
    })
    setTopologyData(res)
    let result: string[][] = []
    res.topologyLevels.map((node: TopologyNode) => {
      if (!result[node.depth]) {
        result[node.depth] = []
      }
      result[node.depth].push(contactServiceEndpoint(node.service, node.endpoint))
    })
    setTopologyLevels(result)
  }

  const initAlertDescendantTrend = async () => {
    const step = getStep(stepStartTime, stepEndTime)
    const res = await getAlertDescendantTrendApi({
      startTime: stepStartTime,
      endTime: stepEndTime,
      service: service,
      endpoint: endpoint,
      step,
      anormalTypes: anormalTypes.join(','),
      clusterIds,
      groupId: dataGroupId,
    })
    type ChartDataMap = Record<string, number>

    const chartData: ChartDataMap = res.anormalCount?.chartData || {}
    setTimeLineChartData(chartData)

    const chartKeys = Object.keys(chartData)
    if (chartKeys.length > 0) {
      const maxCountTimestamp = chartKeys.reduce(
        (maxKey, currentKey) => (chartData[currentKey] > chartData[maxKey] ? currentKey : maxKey),
        chartKeys[0],
      )

      let maxcountTimestamp = Number(maxCountTimestamp)
      if (maxcountTimestamp - step < stepStartTime) {
        maxcountTimestamp = stepStartTime + step
      }
      setDeltaStartTime(maxcountTimestamp - step)
      setDeltaEndTime(maxcountTimestamp)
      setMarklineTimestamp(maxcountTimestamp)
    }
  }
  const initAlertDescendantDelta = async () => {
    try {
      const step = getStep(stepStartTime, stepEndTime)
      setLoading(true)

      const res = await getAlertDescendantDeltaApi({
        startTime: stepStartTime,
        endTime: stepEndTime,
        service,
        endpoint,
        deltaStartTime: stepEndTime - step,
        deltaEndTime: stepEndTime,
        step,
        anormalTypes: anormalTypes.join(','),
        clusterIds,
        groupId: dataGroupId,
      })

      const finalTopologyNodesMap = prepareTopologyNode(res.finalAnormalCounts)
      const midwayTopologyNodesMap = prepareTopologyNode(res.alertTriggeredCounts)
      const originTopologyNodesMap = prepareTopologyNode(res.originAnormalCounts)

      setTopologyNodesMap(finalTopologyNodesMap)
      setOriginTopologyNodesMap(originTopologyNodesMap)
      setMidwayTopologyNodesMap(midwayTopologyNodesMap)
      setFinalTopologyNodesMap(finalTopologyNodesMap)

      setFinalAnormalEvents(res.finalAnormalEvents || [])
      setDeltaAnormalEvents(res.deltaAnormalEvents || [])
    } catch {
      setTopologyNodesMap(null)
      setOriginTopologyNodesMap(null)
      setMidwayTopologyNodesMap(null)
      setFinalTopologyNodesMap(null)
      setFinalAnormalEvents([])
      setDeltaAnormalEvents([])
    }
  }

  //仅改变切片时间或筛选类型
  const changeAlertDescendantDelta = _.debounce(({ deltaStartTime, deltaEndTime }) => {
    //默认初始化，用步长最后一个时间段作为切片
    setDeltaStartTime(deltaStartTime)
    setDeltaEndTime(deltaEndTime)
    setLoading(true)
    setMarklineTimestamp(deltaEndTime)
    const step = getStep(stepStartTime, stepEndTime)
    if (service && endpoint) {
      getAlertDescendantDeltaApi({
        startTime: stepStartTime,
        endTime: stepEndTime,
        service: service,
        endpoint: endpoint,
        deltaStartTime: deltaStartTime,
        deltaEndTime: deltaEndTime,
        step,
        anormalTypes: anormalTypes.join(','),
        clusterIds,
        groupId: dataGroupId,
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
      service: service,
      endpoint: endpoint,
      step,
      anormalTypes: anormalTypes.join(','),
      clusterIds,
      groupId: dataGroupId,
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
                originMap[nodeId] = finalMap?.[nodeId] || undefined
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
            originMap[nodeId] = midMap?.[nodeId] || undefined
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

  //防抖避免跳转使用旧时间
  useDebounce(
    async () => {
      if (startTime && endTime && service && endpoint && dataGroupId !== null) {
        setLoading(true)
        await updateState()
      }
    },
    300, // 延迟时间 300ms
    [startTime, endTime, service, endpoint, clusterIds, dataGroupId],
  )
  useEffect(() => {
    setService(searchParams.get('service'))
    setEndpoint(searchParams.get('endpoint'))
    const clusterIds = searchParams.getAll('clusterIds')
    setClusterIds(clusterIds.length > 0 ? clusterIds : undefined)
  }, [searchParams])

  const updateState = async () => {
    if (!endpoint) return

    setLoading(true)

    try {
      await Promise.all([
        getServiceRelation(),
        initAlertDescendantTrend(),
        initAlertDescendantDelta(),
      ])
    } catch (error) {
      console.log(error)
    } finally {
      setLoading(false)
    }
  }

  useUpdateEffect(() => {
    changeAlertDescendantDelta({
      deltaStartTime: deltaStartTime,
      deltaEndTime: deltaEndTime,
    })
    updateAlertDescendantTrend()
  }, [anormalTypes])
  const memoizedValue = useMemo(
    () => ({
      clusterIds,
      loading,
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
      handlePlay,
      changeAlertDescendantDelta,

      setLoading,
      setTopologyData,
      setAnormalTypes,
      setMarklineTimestamp,
      setTopologyNodesMap,
      setSelectedNodeInfo,
      setDeltaTimerange: (timerange: { deltaStartTime: number; deltaEndTime: number }) => {
        setDeltaStartTime(timerange.deltaStartTime)
        setDeltaEndTime(timerange.deltaEndTime)
      },
      setIsPlaying,
      prepareTopologyNode,
      service,
      endpoint,
    }),
    [
      loading,
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
      service,
      endpoint,
      clusterIds,
    ],
  )

  return (
    <ServiceAnalyzeContext.Provider value={memoizedValue}>
      {children}
    </ServiceAnalyzeContext.Provider>
  )
}
