import { useAGUIChat } from 'src/pro/contexts/AGUIContext'
import { Collapse } from 'antd'
import { LoadingOutlined, CheckCircleOutlined, CloseCircleOutlined } from '@ant-design/icons'
import MarkdownRender from 'src/core/components/Markdown/MarkdownRender'

interface ToolRenderProps {
  toolCall?: {
    id: string
    name: string
    args: string
    result?: any
    content?: string
    status?: 'pending' | 'completed' | 'error'
  }
  isCurrent?: boolean
}

const ToolRender: React.FC<ToolRenderProps> = ({ toolCall, isCurrent = false }) => {
  const { currentToolCall } = useAGUIChat()
  const displayToolCall = toolCall || currentToolCall

  const getStatusIcon = () => {
    if (displayToolCall.status === 'error') {
      return <CloseCircleOutlined className="text-red-500 ml-2" />
    }
    if (displayToolCall.status === 'completed' || displayToolCall.result) {
      return <CheckCircleOutlined className="text-green-500 ml-2" />
    }
    return <LoadingOutlined className="text-blue-500 ml-2" />
  }

  return (
    <div>
      <Collapse
        ghost
        items={[
          {
            key: displayToolCall.id,
            label: (
              <div className="flex items-center ">
                <span className="mr-2">🛠️</span>
                <span className="text-[var(--ant-color-text-secondary)] text-">
                  {displayToolCall.name}
                </span>
                {getStatusIcon()}
              </div>
            ),
            children: (
              <div className="pl-3 border-l border-[var(--ant-color-border)]">
                {(() => {
                  switch (displayToolCall.name) {
                    default:
                      return <MarkdownRender content={displayToolCall.args} />
                  }
                })()}
              </div>
            ),
          },
        ]}
      />
    </div>
  )
}

export default ToolRender
