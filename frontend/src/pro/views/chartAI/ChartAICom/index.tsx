import { marked } from 'marked'
import { useEffect, useState } from 'react'
import { useSelector } from 'react-redux'
import { useEventSource } from 'react-sse-hooks'
import { useDebounce } from 'react-use'
import { selectSecondsTimeRange } from 'src/core/store/reducers/timeRangeReducer'
import { FaCircleArrowUp } from 'react-icons/fa6'
import { useChartAIContext } from 'src/pro/contexts/ChartAIContext'
import { mockDialogue } from './mock'
import { Button, Card, ConfigProvider, Divider, Input } from 'antd'
import ChartInput from './ChartInput'
import AIMessage from './AIMessage'
import HumenMessage from './HumenMessage'
import ExampleCom from './ExampleCom'
const ChartAICom = () => {
  const { startTime, endTime } = useSelector(selectSecondsTimeRange)
  const { service, endpoint, anormalTypes } = useChartAIContext()
  const [dialogue, setDialogue] = useState([])
  // post('/api/alertchat', )
  const controller = new AbortController()
  const signal = controller.signal
  const getChartData = async () => {
    const response = await fetch('http://localhost:8083/api/alertchat', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        // 其他自定义请求头
      },
      body: JSON.stringify({
        endTime: 1731565024409000,
        endpoint: 'POST travel-plan',
        service: 'ts-gateway-service',
        entryService: 'POST travel-plan',
        entryEndpoint: 'ts-gateway-service',
        startTime: 1731563224409000,
        step: 60000000,
        anormalTypes: 'app,container,infra,network,error',
        withTopology: 'true',
      }),
    })
    const reader = response.body.pipeThrough(new TextDecoderStream()).getReader()
    let index = dialogue?.length || 0

    while (true) {
      const { value, done } = await reader.read()
      if (done) break
      const match = value.match(/data: (.+)/)
      if (match) {
        try {
          const parsedData = JSON.parse(match[1])
          if (parsedData && parsedData.content) {
            setDialogue((prev) => {
              // 创建新的对话数组，更新最后一个元素的内容
              const updatedDialogue = [...prev]
              if (updatedDialogue[index]?.role === parsedData.role) {
                updatedDialogue[index].content += parsedData.content
              } else {
                index++
                updatedDialogue[index] = parsedData
              }
              return updatedDialogue
            })
          }
        } catch (error) {
          console.error('JSON 解析错误:', error)
        }
      }
    }
  }

  useDebounce(
    () => {
      if (startTime && endTime) {
        // getChartData()
      }
    },
    300, // 延迟时间 300ms
    [startTime, endTime, service, endpoint, anormalTypes],
  )

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
        <div className="w-full overflow-y-auto flex-1 p-4">
          {mockDialogue.map((item, index) => (
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
          ))}
        </div>
        <ChartInput />
      </div>
    </Card>
  )
}
export default ChartAICom
