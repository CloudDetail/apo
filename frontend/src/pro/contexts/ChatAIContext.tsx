import { ReactNode, useContext, useEffect, useMemo, useState } from 'react'
import { createContext, useContextSelector } from '@fluentui/react-context-selector'
import { useSelector } from 'react-redux'
import { useParams } from 'react-router-dom'
import { useDebounce, useStateList } from 'react-use'
import { selectSecondsTimeRange } from 'src/core/store/reducers/timeRangeReducer'
import ServiceTable from '../components/DataViewCom/ServiceTable'
import serviceImgZH from 'src/pro/assets/dataViewTypeExample/zh/service.png'
import alertImgZH from 'src/pro/assets/dataViewTypeExample/zh/alert.png'
import instanceImgZH from 'src/pro/assets/dataViewTypeExample/zh/instance.png'
import errorInstanceImgZH from 'src/pro/assets/dataViewTypeExample/zh/errorInstance.png'
import logsImgZH from 'src/pro/assets/dataViewTypeExample/zh/logs.png'
import dashboardImgZH from 'src/pro/assets/dataViewTypeExample/zh/dashboard.png'

import serviceImgEN from 'src/pro/assets/dataViewTypeExample/en/service.png'
import alertImgEN from 'src/pro/assets/dataViewTypeExample/en/alert.png'
import instanceImgEN from 'src/pro/assets/dataViewTypeExample/en/instance.png'
import errorInstanceImgEN from 'src/pro/assets/dataViewTypeExample/en/errorInstance.png'
import logsImgEN from 'src/pro/assets/dataViewTypeExample/en/logs.png'
import dashboardImgEN from 'src/pro/assets/dataViewTypeExample/en/dashboard.png'
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
import PolarisPngEn from 'src/pro/assets/en/9-1.png'
import MetricsDashboardCom from '../components/DataViewCom/MetricsDashboard'
import { showToast } from 'src/core/utils/toast'
import GlossaryTable from 'src/core/components/PolarisMetrics/GlossaryTable'
import i18next, { t } from 'i18next'
const i18nKey = 'pro/chatAI:'
const currentLang = i18next.language
const imageMap = {
  en: {
    service: serviceImgEN,
    alert: alertImgEN,
    endpoint: instanceImgEN,
    error: errorInstanceImgEN,
    log: logsImgEN,
    dashboard: dashboardImgEN,
    slow: PolarisPngEn,
  },
  zh: {
    service: serviceImgZH,
    alert: alertImgZH,
    endpoint: instanceImgZH,
    error: errorInstanceImgZH,
    log: logsImgZH,
    dashboard: dashboardImgZH,
    slow: PolarisPng,
  },
}
// 定义上下文的数据类型
const dataViewType = {
  service: {
    title: t(`${i18nKey}dataViewType.serviceOverview`),
    exampleImg: imageMap[currentLang].service,
  },
  alert: {
    title: t(`${i18nKey}dataViewType.alertEvents`),
    exampleImg: imageMap[currentLang].alert,
  },
  endpoint: {
    title: t(`${i18nKey}dataViewType.serviceEndpointInstance`),
    exampleImg: imageMap[currentLang].endpoint,
  },
  error: {
    title: t(`${i18nKey}dataViewType.errorInstance`),
    exampleImg: imageMap[currentLang].error,
  },
  log: {
    title: t(`${i18nKey}dataViewType.logList`),
    exampleImg: imageMap[currentLang].log,
  },
  dashboard: {
    title: t(`${i18nKey}dataViewType.dashboard`),
    exampleImg: imageMap[currentLang].dashboard,
  },
  slow: {
    title: t(`${i18nKey}dataViewType.polarisMetrics`),
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
  const [suspectedNodeList, setSuspectedNodeList] = useState<NodeData[] | []>()

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
    setSelectedSuspectedNode(causeNode)
    setSuspectedNodeList(suspectedNodeList)
  }
  const analysisNode = async (type) => {
    if (type?.length === 0) {
      showToast({
        title: t(`${i18nKey}emptyDataType`),
        color: 'danger',
      })
      return
    }

    if (!selectedSuspectedNode?.service) {
      showToast({
        title: t(`${i18nKey}selectRootCauseNode`),
        color: 'danger',
      })
      return
    }
    const controller = new AbortController()
    setController(controller)
    const step = getStep(startTime, endTime)
    setIsDone(false)
    setIsAutoScroll(false)
    setStreamType('analyze')
    setCauseConfirmvisibilit(false)
    const typeTitle = type.map((item) => dataViewType[item].title)
    setDialogue((prev) => {
      // 使用深拷贝初始化副本，避免未定义元素
      const updatedDialogue = [...prev]
      updatedDialogue.push({
        conversation: [
          {
            message: t(`${i18nKey}confirmSuspectedNode`, {
              service: selectedSuspectedNode?.service,
              endpoint: selectedSuspectedNode?.endpoint,
              typeTitle: typeTitle.join('、'),
            }),
            role: 'human',
          },
          {
            message: t(`${i18nKey}aiResponse`),
            role: 'ai',
          },
        ],
      })
      return updatedDialogue
    })
    setCauseConfirmvisibilit(false)
    setCauseNode(selectedSuspectedNode)
    try {
      const language = i18next.language
      const timeZone = Intl.DateTimeFormat().resolvedOptions().timeZone
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
          language: language,
          timezone: timeZone,
        }),
      })

      const reader = response.body.pipeThrough(new TextDecoderStream()).getReader()
      let buffer = '' // 用于存储未处理的数据
      while (true) {
        try {
          // 读取流数据
          const { value, done } = await reader.read()
          if (done) {
            setIsAutoScroll(false)
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
                await new Promise((resolve) => setTimeout(resolve, 300))
                setDialogue((prev) => {
                  const updatedDialogue = [...prev]
                  const currentConversation =
                    updatedDialogue[updatedDialogue.length - 1].conversation[1]
                  currentConversation.components = [
                    ...(currentConversation.components || []),
                    parsedData.components[i],
                  ]
                  if (type.includes('slow')) {
                    updatedDialogue[updatedDialogue.length - 1].explanation = (
                      <>
                        {t(`${i18nKey}ebpfExplanation`)}
                        <GlossaryTable />
                      </>
                    )
                  }
                  return updatedDialogue
                })
              }
              await new Promise((resolve) => setTimeout(resolve, 300)) //刻意等待
            }
          }
        } catch (error) {
          setIsDone(true)
          break // 根据业务需求决定是否退出循环
        }
      }
    } catch {
      setIsDone(true)
      setIsAutoScroll(false)
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
