import { useEffect, useRef, useState } from 'react'
import { useSelector } from 'react-redux'
import { useDebounce } from 'react-use'
import { selectSecondsTimeRange } from 'src/core/store/reducers/timeRangeReducer'
import { useChartAIContext } from 'src/pro/contexts/ChartAIContext'
import { mockDialogue } from './mock'
import { Button, Card, ConfigProvider, Divider, Input, message } from 'antd'
import ChartInput from './ChartInput'
import AIMessage from './AIMessage'
import HumenMessage from './HumenMessage'
import ExampleCom from './ExampleCom'
import ConversationCom from './ConversationCom'
import { getStep } from 'src/core/utils/step'
const ChartAICom = () => {
  const { startTime, endTime } = useSelector(selectSecondsTimeRange)
  const { service, endpoint, anormalTypes } = useChartAIContext()
  const [dialogue, setDialogue] = useState([])
  const [isDone, setIsDone] = useState<boolean>(true)
  // post('/api/alertchat', )
  const controller = new AbortController()
  const signal = controller.signal
  const getChartData = async () => {
    setIsDone(false)
    const step = getStep(startTime, endTime)
    // const response = await fetch('http://localhost:8083/ai/api/alertchat', {
    const response = await fetch('/ai/api/alertchat', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        // 其他自定义请求头
      },
      body: JSON.stringify({
        endTime: startTime,
        service: service,
        endpoint: endpoint,
        entryService: service,
        entryEndpoint: endpoint,
        startTime: 1731563224409000,
        step: step,
        anormalTypes: 'app,container,infra,network,error',
        withTopology: true,
      }),
    })
    const reader = response.body.pipeThrough(new TextDecoderStream()).getReader()
    let index = dialogue?.length ? dialogue.length - 1 : 0

    while (true) {
      try {
        // 读取流数据
        const { value, done } = await reader.read()
        // console.log(done)
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
            if (!updatedDialogue[index]) {
              updatedDialogue[index] = {
                conversation: [],
              }
            }
            if (parsedData.role === 'ai') {
              updatedDialogue[index].conversation[1] = {
                message:
                  (updatedDialogue[index].conversation[1]?.message || '') + parsedData.content,
                role: 'ai',
              }
            } else if (
              parsedData.role === 'human' &&
              updatedDialogue[index].conversation.length === 0
            ) {
              updatedDialogue[index] = {
                conversation: [
                  {
                    role: 'human',
                    message: parsedData.content,
                    fileList: parsedData?.filename ? [parsedData?.filename] : null,
                  },
                ],
              }
            } else {
              updatedDialogue[++index] = {
                conversation: [
                  {
                    role: 'human',
                    message: parsedData.content,
                    fileList: parsedData?.filename ? [parsedData?.filename] : null,
                  },
                ],
              }
            }
            // console.log(updatedDialogue)
            return updatedDialogue
          })
        }
      } catch (error) {
        console.error('读取或处理消息时发生错误:', error)
        break // 根据业务需求决定是否退出循环
      }
    }
    setIsDone(true)
  }

  useDebounce(
    () => {
      if (startTime && endTime) {
        getChartData()
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
      <div className="h-full w-full flex overflow-hidden flex-col">
        <div
          className="flex "
          style={{
            borderBottom: '1px solid #303030',
            paddingLeft: '1.5rem',
            paddingRight: '1.6rem',
          }}
        >
          <div className="w-2/3 p-3  text-lg font-bold">大模型对话</div>
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
          {/* {mockDialogue.map((item, index) => (
            <>
              <div className="flex " key={index}>
                <div className="w-2/3 flex-grow-0 flex-shrink-0">
                  {item.conversation?.map((conv, cIndex) => (
                    <div key={cIndex}>
                      {conv?.role === 'ai' ? (
                        <AIMessage message={conv?.message || ''} />
                      ) : (
                        <HumenMessage message={conv?.message || ''} />
                      )}
                    </div>
                  ))}
                </div>
                <div>
                  <Divider type="vertical" className="h-full" />
                </div>
                <ExampleCom children={item?.explanation} />
              </div>
              <Divider />
            </>
          ))} */}
        </div>
        <ChartInput />
      </div>
    </Card>
  )
}
export default ChartAICom
