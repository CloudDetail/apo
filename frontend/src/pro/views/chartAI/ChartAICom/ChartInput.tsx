import { Button, ConfigProvider, Input } from 'antd'
import { useRef, useState } from 'react'
import { FaRegStopCircle } from 'react-icons/fa'
import { FaCircleArrowUp } from 'react-icons/fa6'

interface ChartInputProps {
  addChart: any
  loading: boolean
}
const ChartInput = ({ addChart, loading }: ChartInputProps) => {
  const [question, setQuestion] = useState<string>('')
  const textareaRef = useRef(null)
  const askQuestion = (e: any) => {
    e.preventDefault()
    if (question && !loading) {
      addChart(question)
      setQuestion('')
    }
  }
  return (
    <div className="w-full  min-h-[80px] rounded-lg  text-center left-0 flex items-center justify-center ">
      <div className="w-2/3 min-h-[40px] bg-[#2f2f2f] rounded-lg  flex items-center justify-center mb-3 overflow-hidden">
        <ConfigProvider
          theme={{
            components: {
              Input: {
                activeBorderColor: 'none',
                activeBg: '#2f2f2f',
                hoverBorderColor: '#2f2f2f',
                hoverBg: '#2f2f2f',
                activeShadow: '#2f2f2f',
              },
            },
          }}
        >
          <Input.TextArea
            ref={textareaRef}
            className=" bg-[#2f2f2f] pr-10 "
            variant="filled"
            autoSize={{ minRows: 1, maxRows: 5 }}
            style={{ background: '#2f2f2f ' }}
            placeholder="对话大模型"
            value={question}
            onChange={(e) => setQuestion(e.target.value)}
            onPressEnter={(e) => askQuestion(e)}
            rows={1}
          ></Input.TextArea>
        </ConfigProvider>
        {loading ? (
          <Button
            type="text"
            shape="circle"
            className="mr-2  "
            icon={<FaRegStopCircle size={24} />}
            disabled
          ></Button>
        ) : (
          <Button
            type="text"
            shape="circle"
            className="mr-2  "
            icon={<FaCircleArrowUp size={24} />}
            onClick={askQuestion}
          ></Button>
        )}
      </div>
    </div>
  )
}
export default ChartInput
