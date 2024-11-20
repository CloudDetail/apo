import React from 'react'
import AIMessage from './AIMessage'
import HumanMessage from './HumanMessage'
import { Divider } from 'antd'
import ExampleCom from './ExampleCom'
import { mockDialogue } from './mock'

interface CnversationData {
  message: string
  role: string
  fileList: string[]
  reportUrlList?: string[]
  showTyping?: boolean
  img?: any
}

interface DialogueData {
  conversation: CnversationData[]
}
interface ConversationComProps {
  dialogue: DialogueData
  index: number
  loading: boolean
}

const ConversationCom = ({ dialogue, index, loading }: ConversationComProps) => {
  return (
    <>
      <div className="flex ">
        <div className="w-2/3 flex-grow-0 flex-shrink-0">
          {dialogue?.conversation?.map((item) => {
            return item.role === 'ai' ? (
              <AIMessage
                message={item.message || ''}
                loading={loading}
                reportUrlList={item?.reportUrlList || []}
              />
            ) : (
              <HumanMessage
                message={item.message || ''}
                fileList={item.fileList || []}
                img={item?.img || null}
              />
            )
          })}
        </div>
        <div>
          <Divider type="vertical" className="h-full" />
        </div>

        <ExampleCom children={mockDialogue[index]?.explanation || <></>} />
      </div>
      {!loading && <Divider />}
    </>
  )
}

export default ConversationCom
