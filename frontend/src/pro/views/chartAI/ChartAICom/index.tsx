import { useEffect, useRef, useState } from 'react'
import { useSelector } from 'react-redux'
import { useDebounce } from 'react-use'
import { selectSecondsTimeRange } from 'src/core/store/reducers/timeRangeReducer'
import { useChartAIContext } from 'src/pro/contexts/ChartAIContext'
import { Card } from 'antd'
import ChartInput from './ChartInput'
import ConversationCom from './ConversationCom'
import { getStep } from 'src/core/utils/step'
import HintCom from './HintCom'
import './index.css'
import LoadingSpinner from 'src/core/components/Spinner'
const ChartAICom = () => {
  const { startTime, endTime } = useSelector(selectSecondsTimeRange)
  const { service, endpoint, anormalTypes } = useChartAIContext()
  const [dialogue, setDialogue] = useState([])
  const [isDone, setIsDone] = useState<boolean>(true)
  const [controller, setController] = useState<any>(null)
  const [chatQuery, setChatQuery] = useState<any>(null)
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
            }
          }
          // 如果当前对话框有两条消息，则切换到下一个对话框
          else if (updatedDialogue[currentIndex].conversation.length === 2) {
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
              }
            }
          }
          // 否则在当前对话框中追加字符
          else {
            updatedDialogue[currentIndex].conversation[0].message += nextChunk
          }

          // 在最后一个字符时，添加文件列表和 AI 回应初始化
          if (typeIndex + step >= content.length) {
            updatedDialogue[currentIndex].conversation[0].fileList = parsedData.filename
              ? [parsedData.filename]
              : null
            updatedDialogue[currentIndex].conversation[1] = {
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
    if (controller) {
      controller.abort()
      console.log('Fetch stream closed.')
    }
  }
  const getChartData = async () => {
    setIsDone(false)
    const step = getStep(startTime, endTime)
    const response = await fetch('/ai/api/alertchat', {
      // const response = await fetch('http://localhost:8083/stream', {
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
        anormalTypes: 'app,container,infra,network,error',
        withTopology: true,
      }),
    })

    const reader = response.body.pipeThrough(new TextDecoderStream()).getReader()
    let index = dialogue?.length ? dialogue.length - 1 : 0
    let isPaused = false // 用于控制暂停和继续

    let buffer = '' // 用于存储未处理的数据

    const processChunk = async () => {
      if (isPaused) return // 如果暂停，则不继续读取

      try {
        const { value, done } = await reader.read()
        if (done) {
          setIsDone(true)
          return // 流结束时退出
        }
        // 确保 `value` 是字符串类型
        if (typeof value !== 'string') return processChunk()
        // 提取流中的数据

        buffer += value // 将新读取的数据追加到缓冲区
        // 按 `\n\ndata:` 分割数据
        const parts = buffer.split('\n\ndata:')
        parts?.forEach(async (part, pIndex) => {
          let newPart = pIndex === 0 ? `${part}` : `data:${part}`
          const match = newPart.match(/data: (.+)/)
          if (!match) {
            return
          }

          // 解析 JSON 数据
          let parsedData
          try {
            parsedData = JSON.parse(match[1])
          } catch (jsonError) {
            // console.error('JSON 解析失败:', jsonError, match[1])
            console.log('JSON 解析失败:', newPart)
            return // 跳过解析失败的数据
          }
          buffer = ''
          // 处理内容
          if (parsedData?.content || parsedData?.reportUrlList) {
            if (parsedData.role === 'ai') {
              setDialogue((prev) => {
                const updatedDialogue = [...prev]
                updatedDialogue[index].conversation[1] = {
                  message:
                    (updatedDialogue[index].conversation[1]?.message || '') +
                    (parsedData?.content || ''),
                  role: 'ai',
                  reportUrlList: [
                    ...(updatedDialogue[index].conversation[1]?.reportUrlList || []),
                    ...(parsedData.reportUrlList || []),
                  ],
                }
                return updatedDialogue
              })
              processChunk() // 继续处理流
            } else if (parsedData.role === 'human') {
              // 暂停流处理
              isPaused = true
              index = await typeMessage(parsedData, index, setDialogue)
              isPaused = false // 恢复流处理
              processChunk() // 继续处理流
            }
          } else {
            processChunk() // 没有内容则继续处理
          }
        })
        processChunk()
      } catch (error) {
        console.error('流读取失败:', error)
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
    const response = await fetch(`/ai/api/chat?${queryString}`, {
      method: 'GET',
      signal: controller.signal,
    })
    const reader = response.body.pipeThrough(new TextDecoderStream()).getReader()

    while (true) {
      try {
        // 读取流数据
        const { value, done } = await reader.read()
        if (done) break // 流结束时退出循环

        // 确保 `value` 是字符串类型
        if (typeof value !== 'string') continue

        // 提取流中的数据
        const match = value.match(/data: (.+)/)
        if (!match) continue

        // 解析 JSON 数据
        let parsedData
        try {
          parsedData = JSON.parse(match[1])
        } catch (jsonError) {
          console.error('JSON 解析失败:', jsonError, match[1])
          continue // 跳过解析失败的数据
        }

        // 处理内容
        if (parsedData?.content) {
          setDialogue((prev) => {
            // 使用深拷贝初始化副本，避免未定义元素
            const updatedDialogue = [...prev]
            updatedDialogue[updatedDialogue?.length - 1].conversation[1] = {
              message:
                (updatedDialogue[updatedDialogue?.length - 1].conversation[1]?.message || '') +
                parsedData.content,
              role: 'ai',
            }
            return updatedDialogue
          })
        }
      } catch (error) {
        console.error('读取或处理消息时发生错误:', error)
        break // 根据业务需求决定是否退出循环
      }
    }
    setIsDone(true)
    setChatQuery(null)
  }
  const updateController = () => {
    setDialogue([])
    closeStream()
    setController(new AbortController())
  }
  useEffect(() => {
    if (controller) {
      if (chatQuery) {
        addChart()
      } else {
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
      }
    },
    300, // 延迟时间 300ms
    [startTime, endTime, service, endpoint, anormalTypes],
  )
  const chatContainerRef = useRef(null)
  const [isAutoScroll, setIsAutoScroll] = useState(true) // 是否自动滚动

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

  useEffect(() => {
    const chatContainer = chatContainerRef.current

    if (chatContainer && isAutoScroll) {
      // 如果是自动滚动模式，则滚动到最底部
      chatContainer.scrollTop = chatContainer.scrollHeight
    }
  }, [dialogue, isAutoScroll]) // 在消息更新时检查自动滚动状态
  return (
    <Card className="h-full overflow-hidden" classNames={{ body: 'p-0 h-full' }}>
      <LoadingSpinner loading={!isDone && dialogue?.length === 0} />
      <div className="h-full w-full flex overflow-hidden flex-col">
        <div
          className="flex "
          style={{
            borderBottom: '1px solid #303030',
            paddingLeft: '1.5rem',
            paddingRight: '1.6rem',
          }}
        >
          <div className="w-2/3 p-3  text-lg font-bold flex items-center">
            大模型对话 <HintCom />
          </div>
          <div className=" p-3  text-lg font-bold" style={{ borderLeft: '1.5px solid #303030' }}>
            数据与规则说明
          </div>
        </div>
        <div ref={chatContainerRef} className="w-full overflow-y-auto flex-1 p-4">
          {dialogue.map((item, index) => (
            <ConversationCom
              dialogue={item}
              index={index}
              key={index}
              loading={!isDone && index === dialogue.length - 1}
            />
          ))}
        </div>
        <ChartInput addChart={setChatQuery} loading={!isDone} closeStream={closeStream} />
      </div>
    </Card>
  )
}
export default ChartAICom
