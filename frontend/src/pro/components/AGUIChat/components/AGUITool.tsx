import { useAGUIChat } from 'src/pro/contexts/AGUIContext'
import { Collapse } from 'antd'
import { LoadingOutlined, CloseCircleOutlined } from '@ant-design/icons'
import MarkdownRender from 'src/core/components/Markdown/MarkdownRender'
import styles from '../AGUIChat.module.scss'
import ToolCollapseIcon from './ToolCollapseIcon'
import { useState } from 'react'

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
  const [hovered, setHovered] = useState(false)
  const getStatusIcon = () => {
    if (displayToolCall.status === 'error') {
      return <CloseCircleOutlined className="text-red-500 ml-2" />
    }
    if (displayToolCall.status === 'completed' || displayToolCall.result) {
      return <ToolCollapseIcon hovered={hovered} />
    }
    return <LoadingOutlined className="text-blue-500 ml-2" />
  }

  return (
    <div>
      <Collapse
        ghost
        className={styles.AGUItoolCallCollapse}
        expandIcon={() => getStatusIcon()}
        items={[
          {
            key: displayToolCall.id,
            label: (
              <div
                className="flex items-center "
                onMouseEnter={() => setHovered(true)}
                onMouseLeave={() => setHovered(false)}
              >
                <span className="mr-2">🛠️</span>
                <span className="text-[var(--ant-color-text-tertiary)] text-">
                  {displayToolCall.name}
                </span>
              </div>
            ),
            children: (
              <div className="m-[6px] px-3 relative border-l-2 border-[var(--ant-color-border)] ">
                <div className="bg-[var(--ant-color-bg-elevated)] rounded-2xl p-3">
                  <MarkdownRender content={displayToolCall.args} />
                </div>
              </div>
            ),
          },
        ]}
      />
    </div>
  )
}

export default ToolRender
