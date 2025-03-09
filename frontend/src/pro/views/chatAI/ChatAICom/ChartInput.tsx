import { Button, ConfigProvider, Input } from 'antd'
import { t } from 'i18next'
import React from 'react'
import { useRef, useState } from 'react'
import { FaRegStopCircle } from 'react-icons/fa'
import { FaCircleArrowUp } from 'react-icons/fa6'

interface ChartInputProps {
  addChart: any
  loading: boolean
  closeStream: any
}
const ChartInput = ({ addChart, loading, closeStream }: ChartInputProps) => {
  const [question, setQuestion] = useState<string>('')
  const textareaRef = useRef(null)
  const askQuestion = (e: any) => {
    if (question && !loading) {
      addChart(question)
      setQuestion('')
    }
  }
  const changeQueationValue = (e: any) => {
    if (!loading && e.target.value !== '\n') {
      setQuestion(e.target.value)
    }

    // if (question && !loading) {
    //   addChart(question)
    //   setQuestion('')
    //   requestAnimationFrame(() => {
    //     e.preventDefault()
    //   })
    // }
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
            placeholder={t('pro/chatAI:chatWithAI')}
            value={question}
            onChange={(e) => changeQueationValue(e)}
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
            onClick={() => closeStream()}
          ></Button>
        ) : (
          <Button
            type="text"
            shape="circle"
            className="mr-2  "
            icon={<FaCircleArrowUp size={24} />}
            onClick={(e) => askQuestion(e)}
          ></Button>
        )}
      </div>
    </div>
  )
}
export default React.memo(ChartInput)
