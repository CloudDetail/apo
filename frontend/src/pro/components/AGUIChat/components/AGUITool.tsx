import { Collapse } from 'antd'
import { LoadingOutlined, CloseCircleOutlined } from '@ant-design/icons'
import MarkdownRender from 'src/core/components/Markdown/MarkdownRender'
import styles from '../AGUIChat.module.scss'
import ToolCollapseIcon from './ToolCollapseIcon'
import { useEffect, useState } from 'react'
import ShinyText from 'src/pro/components/Animation/ShinyText'
import { FaTools } from 'react-icons/fa'
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
  open?: boolean
}

const ToolRender: React.FC<ToolRenderProps> = ({ toolCall, open = false }) => {
  const displayToolCall = toolCall
  const [hovered, setHovered] = useState(false)
  const [activeKey, setActiveKey] = useState([])
  const getStatusIcon = () => {
    if (displayToolCall.status === 'error') {
      return <CloseCircleOutlined className="text-red-500 ml-2" />
    }
    if (displayToolCall.status === 'completed' || displayToolCall.result) {
      return <ToolCollapseIcon hovered={hovered} />
    }
    return <LoadingOutlined className="text-blue-500 ml-2" />
  }
  useEffect(() => {
    if (open) {
      setActiveKey(displayToolCall.id)
    }
  }, [open])
  return (
    <div>
      <Collapse
        ghost
        className={styles.AGUItoolCallCollapse}
        expandIcon={() => getStatusIcon()}
        onChange={setActiveKey}
        activeKey={activeKey}
        items={[
          {
            key: displayToolCall.id,
            label: (
              <div
                className="flex items-start "
                onMouseEnter={() => setHovered(true)}
                onMouseLeave={() => setHovered(false)}
              >
                <div className="items-center flex mr-2 mt-1">
                  <FaTools />
                </div>
                <span className="text-[var(--ant-color-text-tertiary)] text-">
                  <ShinyText
                    text={displayToolCall.name}
                    disabled={displayToolCall.status !== 'pending'}
                    speed={3}
                  />
                </span>
              </div>
            ),
            children: (
              <div className="m-[6px] px-3 relative border-l-2 border-[var(--ant-color-border)] ">
                <div className="bg-[var(--ant-color-bg-elevated)] rounded-2xl p-3 max-h-[500px] overflow-y-auto">
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
