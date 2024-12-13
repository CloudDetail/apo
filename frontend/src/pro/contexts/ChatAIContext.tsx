import { ReactNode, useContext, useEffect, useMemo, useState } from 'react'
import { createContext, useContextSelector } from '@fluentui/react-context-selector'
import { useSelector } from 'react-redux'
import { useParams } from 'react-router-dom'
import { useDebounce, useStateList } from 'react-use'
import { selectSecondsTimeRange } from 'src/core/store/reducers/timeRangeReducer'
import ServiceTable from '../components/DataViewCom/ServiceTable'
import serviceImg from 'src/pro/assets/dataViewTypeExample/service.png'
import alertImg from 'src/pro/assets/dataViewTypeExample/alert.png'
import instanceImg from 'src/pro/assets/dataViewTypeExample/instance.png'
import errorInstanceImg from 'src/pro/assets/dataViewTypeExample/errorInstance.png'
import logsImg from 'src/pro/assets/dataViewTypeExample/logs.png'
import dashboardImg from 'src/pro/assets/dataViewTypeExample/dashboard.png'
import { getStep } from 'src/core/utils/step'
import streamFetch from '../util/streamFetch'
import SlowType from '../components/DataViewCom/SlowType'
import LogsTabs from '../components/DataViewCom/FaultSiteLogs/LogsTabs'
import { Button, Card } from 'antd'
import { anormalTypes } from '../constants'
import AnormalListCom from '../components/DataViewCom/AnormalListCom'
import FaultSiteLogsWrapper from '../components/DataViewCom/FaultSiteLogs'
import InstanceTable from '../components/DataViewCom/InstanceTable'
import ErrorInstanceTable from '../components/DataViewCom/ErrorInstanceTable'
import PolarisPng from 'src/pro/assets/9-1.png'
import MetricsDashboardCom from '../components/DataViewCom/MetricsDashboard'
import { showToast } from 'src/core/utils/toast'
// 定义上下文的数据类型
const dataViewType = {
  service: {
    title: '服务概览指标',
    exampleImg: serviceImg,
  },
  alert: {
    title: '告警事件',
    exampleImg: alertImg,
  },
  endpoint: {
    title: '服务端点实例',
    exampleImg: instanceImg,
  },
  error: {
    title: '错误实例',
    exampleImg: errorInstanceImg,
  },
  log: {
    title: '日志列表',
    exampleImg: logsImg,
  },
  dashboard: {
    title: '大盘',
    exampleImg: dashboardImg,
  },
  slow: {
    title: '北极星指标和延时报告',
    exampleImg: PolarisPng,
  },
}
const dataViewTypeList = Object.entries(dataViewType).map(([key, value]) => ({
  key: key,
  title: value.title,
}))
type DataViewType = typeof dataViewType
type DataViewTypeList = typeof dataViewTypeList
interface NodeData {
  service: string
  endpoint: string
  isRoot?: boolean
}
interface ChatAIType {
  // 分析的入口业务
  service: string | null
  endpoint: string | null
  //当前分析的根因节点
  causeNode: NodeData | null
  //疑似节点列表
  suspectedNodeList: NodeData[] | []
  //可选的查看
  dataViewType: DataViewType
  dataViewTypeList: DataViewTypeList
  //
  changeSuspectedNodeList: any
  //确认下一个需要分析的疑似节点
  selectedSuspectedNode: NodeData | null
  setSelectedSuspectedNode: any
  dialogue: any
  setDialogue: any
  isDone: boolean
  setIsDone: any
  //确认根因的框
  causeConfirmvisibility: boolean
  setCauseConfirmvisibilit: any
  controller: any
  setController: any
  // 当前流式传输由谁发起的
  streamType: 'init' | 'quiz' | 'analyze'
  setStreamType: any
  closeStream: any
  isAutoScroll: boolean
  setIsAutoScroll: any
}

// 创建上下文
const ChatAIContext = createContext<ChatAIType>({} as ChatAIType)

// 创建自定义 Hook
export const useChatAIContext = <T,>(selector: (context: ChatAIType) => T): T =>
  useContextSelector(ChatAIContext, selector)
// 创建 Provider 组件
export const ChatAIProvider: React.FC<{ children: ReactNode }> = ({ children }) => {
  const { startTime, endTime } = useSelector(selectSecondsTimeRange)
  const { service = null, endpoint = null } = useParams()
  const [analysisType, setAnalysisType] = useState([])
  const [dialogue, setDialogue] = useState([])
  const [causeNode, setCauseNode] = useState<NodeData | null>(null)
  const [suspectedNodeList, setSuspectedNodeList] = useState<NodeData[] | []>([
    {
      service: 'ts-train-service',
      endpoint: 'POST /api/v1/trainservice/trains/byNames',
      isRoot: true,
    },
    {
      service: 'ts-station-service',
      endpoint: 'POST /api/v1/stationservice/stations/idlist',
      isRoot: false,
    },
    {
      service: 'ts-price-service',
      endpoint: 'POST /api/v1/priceservice/prices/byRouteIdsAndTrainTypes',
      isRoot: false,
    },
  ])

  const [selectedSuspectedNode, setSelectedSuspectedNode] = useState<NodeData | null>(null)
  const [isDone, setIsDone] = useState<boolean>(true)
  const [causeConfirmvisibility, setCauseConfirmvisibilit] = useState(false)
  const [controller, setController] = useState<any>(null)
  const [streamType, setStreamType] = useState<'init' | 'quiz' | 'analyze'>('init')
  const [isAutoScroll, setIsAutoScroll] = useState(true) // 是否自动滚动

  const changeSuspectedNodeList = (nodeList: NodeData[] | []) => {
    let causeNode: NodeData | null = null
    const suspectedNodeList: NodeData[] | [] = []
    nodeList.map((node) => {
      if (node.isRoot) {
        causeNode = node
      }
      suspectedNodeList.push(node)
    })
    setCauseNode(causeNode)
    setSuspectedNodeList(suspectedNodeList)
  }
  const analysisNode = async (type) => {
    if (type?.length === 0) {
      showToast({
        title: '查看的数据类型不可为空',
        color: 'danger',
      })
      return
    }
    if (!selectedSuspectedNode?.service) {
      showToast({
        title: '请先选择根因节点',
        color: 'danger',
      })
      return
    }
    const controller = new AbortController()
    setController(controller)
    const step = getStep(startTime, endTime)
    setIsDone(false)
    setStreamType('analyze')
    setCauseConfirmvisibilit(false)
    const typeTitle = type.map((item) => dataViewType[item].title)
    setDialogue((prev) => {
      // 使用深拷贝初始化副本，避免未定义元素
      const updatedDialogue = [...prev]
      updatedDialogue.push({
        conversation: [
          {
            message: `疑似节点${selectedSuspectedNode?.service}(${selectedSuspectedNode?.endpoint})为根因节点，请给出关于此节点的${typeTitle.join('、')}数据`,
            role: 'human',
          },
          {
            message: `好的，以下是您确认的根因节点的相关数据`,
            role: 'ai',
          },
        ],
      })
      return updatedDialogue
    })
    setCauseConfirmvisibilit(false)
    setCauseNode(selectedSuspectedNode)
    try {
      const response = await streamFetch('/ai/api/nodeinfo', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        signal: controller.signal,
        body: JSON.stringify({
          endTime,
          service: selectedSuspectedNode?.service,
          endpoint: selectedSuspectedNode?.endpoint,
          entryService: service,
          entryEndpoint: endpoint,
          startTime,
          step,
          anormalTypes: anormalTypes.join(','),
          withTopology: true,
          type: type,
        }),
      })

      const reader = response.body.pipeThrough(new TextDecoderStream()).getReader()
      let buffer = '' // 用于存储未处理的数据
      while (true) {
        try {
          // 读取流数据
          const { value, done } = await reader.read()
          if (done) {
            setCauseConfirmvisibilit(true)
            setIsDone(true)
            break
          } // 流结束时退出循环

          // 确保 `value` 是字符串类型
          if (typeof value !== 'string') continue
          buffer += value // 将新读取的数据追加到缓冲区
          // 按 `\n\ndata:` 分割数据
          const parts = buffer.split('\n\ndata:')

          for (const [pIndex, part] of parts.entries()) {
            let newPart = pIndex === 0 ? `${part}` : `data:${part}`
            // 提取流中的数据
            const match = newPart.match(/data: (.+)/)
            if (!match) {
              continue
            }

            // 解析 JSON 数据
            let parsedData
            try {
              parsedData = JSON.parse(match[1])
            } catch (jsonError) {
              // console.error('JSON 解析失败:', jsonError, match[1])
              continue // 跳过解析失败的数据
            }
            buffer = ''
            // 处理内容
            if (parsedData?.components) {
              for (let i = 0; i < parsedData.components.length; i++) {
                await new Promise((resolve) => setTimeout(resolve, 200)) // 延迟 1 秒
                setDialogue((prev) => {
                  const updatedDialogue = [...prev]
                  const currentConversation =
                    updatedDialogue[updatedDialogue.length - 1].conversation[1]
                  currentConversation.components = [
                    ...(currentConversation.components || []),
                    parsedData.components[i],
                  ]
                  return updatedDialogue
                })
              }
            }
          }
        } catch (error) {
          console.log('读取或处理消息时发生错误:', error)
          setIsDone(true)
          break // 根据业务需求决定是否退出循环
        }
      }
    } catch {
      setIsDone(true)
      setCauseConfirmvisibilit(true)
      return
    }
  }

  const closeStream = () => {
    if (controller) {
      controller.abort()
      console.log('Fetch stream closed.')
    }
  }

  const infoValue = useMemo(
    () => ({
      service,
      endpoint,
    }),
    [service, endpoint],
  )
  const memoizedValue = useMemo(
    () => ({
      causeNode,
      dataViewType,
      dataViewTypeList,
      selectedSuspectedNode,
      analysisType,
      analysisNode,
    }),
    [service, endpoint, causeNode, selectedSuspectedNode, analysisType, analysisNode],
  )
  const suspectedValue = useMemo(() => ({ suspectedNodeList }), [suspectedNodeList])
  const dialogueValue = useMemo(() => ({ dialogue, setDialogue }), [dialogue])
  const isDoneValue = useMemo(() => ({ isDone, setIsDone }), [isDone])
  const controllerValue = useMemo(() => ({ controller, setController }), [controller])
  const streamTypeValue = useMemo(() => ({ streamType, setStreamType }), [streamType])
  const isAutoScrollValue = useMemo(() => ({ isAutoScroll, setIsAutoScroll }), [isAutoScroll])
  const causeConfirmvisibilityValue = useMemo(
    () => ({ causeConfirmvisibility, setCauseConfirmvisibilit }),
    [causeConfirmvisibility],
  )
  const methodValue = useMemo(
    () => ({
      changeSuspectedNodeList,
      setSelectedSuspectedNode,
      setAnalysisType,
    }),
    [],
  )

  const finalMValue = {
    ...infoValue,
    ...memoizedValue,
    ...methodValue,
    ...suspectedValue,
    ...dialogueValue,
    ...isDoneValue,
    ...causeConfirmvisibilityValue,
    ...controllerValue,
    ...streamTypeValue,
    ...isAutoScrollValue,
    closeStream,
  }
  return <ChatAIContext.Provider value={finalMValue}>{children}</ChatAIContext.Provider>
}
