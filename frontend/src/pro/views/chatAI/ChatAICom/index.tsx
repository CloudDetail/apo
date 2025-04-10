import { useEffect, useRef, useState } from 'react'
import { useSelector } from 'react-redux'
import { useDebounce } from 'react-use'
import { selectSecondsTimeRange } from 'src/core/store/reducers/timeRangeReducer'
import { useChatAIContext } from 'src/pro/contexts/ChatAIContext'
import { Button, Divider, Tooltip } from 'antd'
import ChartInput from './ChartInput'
import ConversationCom from './Conversation/ConversationCom'
import { getStep } from 'src/core/utils/step'
import HintCom from './HintCom'
import LoadingSpinner from 'src/core/components/Spinner'
import GlassCard from '../../../components/GlassCard'
import CauseConfirm from './CauseConfirm'
import streamFetch from 'src/pro/util/streamFetch'
import React from 'react'
import { anormalTypes } from 'src/pro/constants'
import { FaLocationCrosshairs } from 'react-icons/fa6'
import { Explanations } from './Explanations'
import { useTranslation } from 'react-i18next'
import CustomCard from 'src/core/components/Card/CustomCard'
const ChatAICom = () => {
  const { t } = useTranslation('pro/chatAI')
  const { startTime, endTime } = useSelector(selectSecondsTimeRange)
  //上下文
  const service = useChatAIContext((ctx) => ctx.service)
  const endpoint = useChatAIContext((ctx) => ctx.endpoint)
  const changeSuspectedNodeList = useChatAIContext((ctx) => ctx.changeSuspectedNodeList)
  const dialogue = useChatAIContext((ctx) => ctx.dialogue)
  const setDialogue = useChatAIContext((ctx) => ctx.setDialogue)
  const isDone = useChatAIContext((ctx) => ctx.isDone)
  const setIsDone = useChatAIContext((ctx) => ctx.setIsDone)
  const controller = useChatAIContext((ctx) => ctx.controller)
  const setController = useChatAIContext((ctx) => ctx.setController)
  const streamType = useChatAIContext((ctx) => ctx.streamType)
  const setStreamType = useChatAIContext((ctx) => ctx.setStreamType)
  const setCauseConfirmvisibilit = useChatAIContext((ctx) => ctx.setCauseConfirmvisibilit)
  const causeConfirmvisibility = useChatAIContext((ctx) => ctx.causeConfirmvisibility)
  const isAutoScroll = useChatAIContext((ctx) => ctx.isAutoScroll)
  const setIsAutoScroll = useChatAIContext((ctx) => ctx.setIsAutoScroll)
  // const [dialogue, setDialogue] = useState([])

  const [chatQuery, setChatQuery] = useState<any>(null)
  const [isConfigure, setIsConfigure] = useState<boolean>(true)
  const typeMessage = async (parsedData: any, index: number, setDialogue: any) => {
    const content = parsedData?.content || '' // 消息内容
    const step = 10 // 每次追加的字符数量
    let typeIndex = 0 // 当前字符索引
    let currentIndex = index // 当前对话框索引
    while (typeIndex < content.length && !controller.signal.aborted) {
      await new Promise<void>((resolve) => {
        setDialogue((prev: any) => {
          const updatedDialogue = [...prev]
          const nextChunk = content.slice(typeIndex, typeIndex + step) // 提取固定数量字符

          // 如果当前对话框不存在，则初始化
          if (!updatedDialogue[currentIndex]) {
            updatedDialogue[currentIndex] = {
              conversation: [
                {
                  role: parsedData.role,
                  message: nextChunk,
                  img: parsedData?.img,
                },
              ],
              explanation: Explanations[currentIndex],
            }
          } else if (
            updatedDialogue[currentIndex].conversation.length > 1 &&
            updatedDialogue[currentIndex].conversation.length < 4 &&
            currentIndex === 0
          ) {
            updatedDialogue[currentIndex].conversation[2] = {
              role: parsedData.role,
              message: (updatedDialogue[currentIndex].conversation[2]?.message || '') + nextChunk,
              img: parsedData?.img,
            }
          } else if (updatedDialogue[currentIndex].conversation.length > 1) {
            // 如果当前对话框有两条消息，则切换到下一个对话框
            currentIndex++
            // 初始化下一个对话框
            if (!updatedDialogue[currentIndex]) {
              updatedDialogue[currentIndex] = {
                conversation: [
                  {
                    role: parsedData.role,
                    message: nextChunk,
                    img: parsedData?.img,
                  },
                ],
                explanation: Explanations[currentIndex],
              }
            }
          }
          // 否则在当前对话框中追加字符
          else {
            updatedDialogue[currentIndex].conversation[0].message += nextChunk
          }

          // 在最后一个字符时，添加文件列表和 AI 回应初始化
          if (typeIndex + step >= content.length) {
            updatedDialogue[currentIndex].conversation[
              updatedDialogue[currentIndex].conversation.length - 1
            ].fileList = parsedData.filename ? [parsedData.filename] : null
            updatedDialogue[currentIndex].conversation[
              updatedDialogue[currentIndex].conversation.length
            ] = {
              role: 'ai',
              message: '',
            }
          }

          resolve() // 状态更新后调用 resolve
          return updatedDialogue // 返回更新后的状态
        })
      })

      // 等待 setTimeout 执行完成后递增 typeIndex
      await new Promise((resolve) => setTimeout(resolve, 10))
      typeIndex += step // 按固定步长递增索引
    }

    return currentIndex // 返回最后一个对话框索引
  }
  // 关闭流
  const closeStream = () => {
    // closeStoreStream()
    if (controller) {
      controller.abort()
      setController(null)
      console.log('Fetch stream closed.')
    }
  }
  const getChartData = async () => {
    setIsDone(false)
    const step = getStep(startTime, endTime)
    const response = await streamFetch('/ai/api/alertchat', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      signal: controller.signal,
      body: JSON.stringify({
        endTime,
        service,
        endpoint,
        entryService: service,
        entryEndpoint: endpoint,
        startTime,
        step,
        anormalTypes: anormalTypes.join(','),
        withTopology: true,
        removeClientCall: true,
      }),
    })
    const contentType = response.headers.get('Content-Type') // 获取响应头

    if (contentType && contentType.includes('application/json')) {
      try {
        const data = await response.json()
        if (data.code === 'B0302') {
          setIsConfigure(false)
        }
      } catch {
      } finally {
        setIsDone(true)
        return
      }
    }
    setIsConfigure(true)
    const reader = response.body.pipeThrough(new TextDecoderStream()).getReader()
    let index = dialogue?.length ? dialogue.length - 1 : 0
    let isPaused = false // 用于控制暂停和继续
    let hasNodeList = false
    let buffer = '' // 用于存储未处理的数据

    const processChunk = async () => {
      if (isPaused) {
        setIsDone(true)

        return // 如果暂停，则不继续读取
      }
      try {
        const { value, done } = await reader.read()

        if (done) {
          setIsDone(true)
          isPaused = true

          if (hasNodeList) {
            setIsAutoScroll(false)
            setCauseConfirmvisibilit(true)
          }
          return // 流结束时退出
        }
        // 确保 `value` 是字符串类型
        if (typeof value !== 'string') return processChunk()
        // 提取流中的数据

        buffer += value // 将新读取的数据追加到缓冲区
        // 按 `\n\ndata:` 分割数据
        const parts = buffer.split('\n\ndata:')

        for (const [pIndex, part] of parts.entries()) {
          let newPart = pIndex === 0 ? `${part}` : `data:${part}`
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
            // console.log('JSON 解析失败:', newPart)
            continue // 跳过解析失败的数据
          }
          buffer = ''
          // 处理内容
          if (parsedData?.nodeList?.length > 0) {
            hasNodeList = true
            changeSuspectedNodeList(parsedData.nodeList)
          }
          if (parsedData?.content || parsedData?.reportUrlList) {
            if (parsedData.role === 'ai') {
              setDialogue((prev) => {
                const updatedDialogue = [...prev]
                let aiIndex = updatedDialogue[index].conversation.length > 2 ? 3 : 1
                updatedDialogue[index].conversation[aiIndex] = {
                  message:
                    (updatedDialogue[index].conversation[aiIndex]?.message || '') +
                    (parsedData?.content || ''),
                  role: 'ai',
                  reportUrlList: [
                    ...(updatedDialogue[index].conversation[aiIndex]?.reportUrlList || []),
                    ...(parsedData.reportUrlList || []),
                  ],
                }
                return updatedDialogue
              })
            } else if (parsedData.role === 'human') {
              // 暂停流处理
              index = await typeMessage(parsedData, index, setDialogue)
            }
          }
        }
        processChunk()
      } catch (error) {
        console.error('Error:', error)
        setIsDone(true)
      }
    }

    // 开始处理流
    processChunk()
  }

  // 发起提问
  const addChart = async () => {
    const queryString = new URLSearchParams({ query: chatQuery }).toString()
    setDialogue((prev) => {
      // 使用深拷贝初始化副本，避免未定义元素
      const updatedDialogue = [...prev]
      updatedDialogue.push({
        conversation: [
          {
            role: 'human',
            message: chatQuery,
          },
          {
            role: 'ai',
            message: '',
          },
        ],
      })
      return updatedDialogue
    })
    setIsDone(false)
    const response = await streamFetch(`/ai/api/chat?${queryString}`, {
      method: 'GET',
      signal: controller.signal,
    })
    const reader = response.body.pipeThrough(new TextDecoderStream()).getReader()

    while (true) {
      try {
        // 读取流数据
        const { value, done } = await reader.read()
        if (done) {
          setIsDone(true)
          break
        } // 流结束时退出循环

        // 确保 `value` 是字符串类型
        if (typeof value !== 'string') continue
        const parts = value.split('\n\ndata:')
        parts?.forEach(async (part, pIndex) => {
          let newPart = pIndex === 0 ? `${part}` : `data:${part}`
          // 提取流中的数据
          const match = newPart.match(/data: (.+)/)
          if (!match) {
            return
          }

          // 解析 JSON 数据
          let parsedData
          try {
            parsedData = JSON.parse(match[1])
          } catch (jsonError) {
            console.error('JSON 解析失败:', jsonError, match[1])
            return // 跳过解析失败的数据
          }

          // 处理内容
          if (parsedData?.content) {
            setDialogue((prev) => {
              // 使用深拷贝初始化副本，避免未定义元素
              const updatedDialogue = [...prev]
              updatedDialogue[updatedDialogue?.length - 1].conversation[1] = {
                message:
                  (updatedDialogue[updatedDialogue?.length - 1].conversation[1]?.message || '') +
                  parsedData?.content,
                role: 'ai',
              }
              return updatedDialogue
            })
          }
        })
      } catch (error) {
        console.log('Error:', error)
        setIsDone(true)
        break // 根据业务需求决定是否退出循环
      }
    }

    setChatQuery(null)
  }
  // 分析其他节点

  const updateController = () => {
    setDialogue([])
    closeStream()
    setController(new AbortController())
  }
  useEffect(() => {
    if (controller) {
      if (chatQuery) {
        addChart()
        setCauseConfirmvisibilit(false)
      } else if (streamType === 'init') {
        getChartData()
      }
    }
  }, [controller])
  useEffect(() => {
    if (chatQuery) {
      setController(new AbortController())
    }
  }, [chatQuery])
  useDebounce(
    () => {
      if (startTime && endTime) {
        updateController()
        setStreamType('init')
        changeSuspectedNodeList([])
        setCauseConfirmvisibilit(false)
      }
    },
    300, // 延迟时间 300ms
    [startTime, endTime, service, endpoint, anormalTypes],
  )
  const chatContainerRef = useRef(null)

  useEffect(() => {
    const chatContainer = chatContainerRef.current

    const handleScroll = () => {
      if (!chatContainer) return

      // 检查是否在底部
      const isAtBottom =
        chatContainer.scrollHeight - chatContainer.scrollTop <= chatContainer.clientHeight + 5 // 容忍误差5px

      setIsAutoScroll(isAtBottom)
    }

    // 监听滚动事件
    chatContainer.addEventListener('scroll', handleScroll)
    return () => chatContainer.removeEventListener('scroll', handleScroll)
  }, [])
  const scrollToBottom = () => {
    const chatContainer = chatContainerRef.current
    if (chatContainer) chatContainer.scrollTop = chatContainer.scrollHeight
  }
  useEffect(() => {
    if (isAutoScroll) {
      scrollToBottom()
    }
  }, [dialogue, isAutoScroll, causeConfirmvisibility]) // 在消息更新时检查自动滚动状态
  return (
    <CustomCard style={{ overflow: 'hidden' }} classNames={{ body: 'h-full p-0' }}>
      <LoadingSpinner loading={!isDone && dialogue?.length === 0} />
      <div className="h-full w-full flex overflow-hidden flex-col">
        <div
          className="flex overflow-y-scroll px-4"
          style={{
            borderBottom: '1px solid #303030',
          }}
        >
          <div className="w-2/3 p-1 text-base font-bold flex items-center">
            {t('aiDialogue')} <HintCom />
          </div>
          <Divider type="vertical" className="h-full" />
          <div className="p-2 text-base font-bold flex-grow ml-1">
            <div className="w-full  flex justify-between">
              <div>{t('dataAndRules')}</div>
            </div>
          </div>
        </div>
        <Tooltip title={t('confirmRootCause')}>
          <Button
            className="absolute right-5 bottom-20 z-20"
            onClick={() => {
              if (!causeConfirmvisibility) {
                setIsAutoScroll(true)
              }

              setCauseConfirmvisibilit(!causeConfirmvisibility)
            }}
            type="primary"
            icon={<FaLocationCrosshairs />}
            disabled={!isDone}
            shape="circle"
            size="large"
          ></Button>
        </Tooltip>
        <div ref={chatContainerRef} className="w-full overflow-y-scroll flex-1 relative">
          {!isConfigure && (
            <GlassCard
              content={
                <p>
                  {t('notConfigured')}
                  <a
                    className="underline text-sky-500"
                    target="_blank"
                    href="https://kindlingx.com/docs/APO%20向导式可观测性中心/配置指南/APO接入大语言模型分析告警事件"
                  >
                    {t('configurationGuide')}
                  </a>
                  {t('configureNow')}
                </p>
              }
            />
          )}
          <div className=" px-4 pt-2 overflow-x-hidden" id="conversation">
            {dialogue.map((item, index) => (
              <ConversationCom
                dialogue={item}
                index={index}
                key={index}
                loading={!isDone && index === dialogue.length - 1}
              />
            ))}
          </div>
          <div className="relative mb-3">
            <CauseConfirm />
          </div>
        </div>
        <ChartInput addChart={setChatQuery} loading={!isDone} closeStream={closeStream} />
      </div>
    </CustomCard>
  )
}
export default React.memo(ChatAICom)
