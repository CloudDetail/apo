import { Button, theme } from 'antd'
import MarkdownRender from 'src/core/components/Markdown/MarkdownRender'
import { ToolCallItem, AgentToolDisplayItem } from 'src/pro/contexts/AGUIContext'
import AgentToolDisplay from '../AgentToolDisplay'
import ToolRender from '../AGUITool'
import { CloseOutlined } from '@ant-design/icons'
interface NodeDetailPanelProps {
  visible: boolean
  onClose: () => void
  toolCalls: ToolCallItem[]
  agentToolDisplay: AgentToolDisplayItem[]
  title: string
  content: string
}
const { useToken } = theme

const NodeDetailPanel: React.FC<NodeDetailPanelProps> = ({
  visible,
  onClose,
  toolCalls,
  agentToolDisplay,
  title,
  content,
}) => {
  const { token } = useToken()

  if (!visible) return null
  console.log('content', content)
  return (
    <div
      className="w-full h-full border p-2 rounded-lg flex flex-col"
      style={{
        // backgroundColor: token.colorBgLayout,
        borderColor: token.colorBorder,
      }}
    >
      {/* 头部 */}
      <div
        className="flex items-center justify-between p-3 border-b"
        style={{
          backgroundColor: token.colorBgContainer,
          borderColor: token.colorBorder,
          boxShadow: token.boxShadowTertiary,
        }}
      >
        <div className="text-base font-semibold" style={{ color: token.colorText }}>
          {title}
        </div>
        <Button type="text" icon={<CloseOutlined />} onClick={onClose} size="small" />
      </div>

      {/* 内容区域 */}
      <div className="flex-1 overflow-y-auto p-3 space-y-2">
        {content && <MarkdownRender content={content} theme="light" />}
        {/* 工具调用 */}
        {toolCalls.map((toolCall) => (
          <div key={toolCall.id} className="rounded-md">
            <ToolRender toolCall={toolCall} />
          </div>
        ))}
        {/* 展示工具 */}
        {agentToolDisplay.map((display, index) => (
          <div key={`${display.toolName}-${index}`} className="rounded-md">
            <AgentToolDisplay
              toolName={display.toolName}
              title={display.title}
              content={display.content}
            />
          </div>
        ))}
        {/* 无数据提示 */}
        {/* {toolCalls.length === 0 && agentToolDisplay.length === 0 && (
          <div className="text-center py-8 text-sm" style={{ color: token.colorTextSecondary }}>
            暂无数据
          </div>
        )} */}
      </div>
    </div>
  )
}
export default NodeDetailPanel
