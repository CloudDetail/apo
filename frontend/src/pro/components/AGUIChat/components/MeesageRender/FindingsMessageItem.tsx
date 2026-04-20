import React, { useEffect, useMemo, useState } from 'react'
import { RightOutlined, DownOutlined } from '@ant-design/icons'
import { ExtendedAGUIMessage, useAGUIFindings } from 'src/pro/contexts/AGUIContext'
import FindingsTopology from '../FindingsTree/FindingsTopology'
import type { FindingsNodeData } from '../FindingsTree/findingsContract'
import NodeDetailPanel from './NodeDetailPanel'
import MarkdownRender from 'src/core/components/Markdown/MarkdownRender'

const FindingsMessageItem = ({
  message,
  expanded,
}: {
  message: ExtendedAGUIMessage
  expanded: boolean
}) => {
  const { getFindingsToolCalls, getFindingsAgentToolDisplay } = useAGUIFindings()

  const [panelVisible, setPanelVisible] = useState(false)
  const [selectedNode, setSelectedNode] = useState<FindingsNodeData | null>(null)
  const [isExpanded, setIsExpanded] = useState(expanded) // 控制拓扑图展开/收起状态
  const { tools: historyTools, displays: historyDisplays } = message.findingsData

  const isPending = message.status === 'pending'
  const isCompleted = message.status === 'completed'

  // 从历史数据或上下文中获取工具调用数据
  const getToolCalls = (nodeId: string, findingsId: string) => {
    const historyData = historyTools?.[nodeId]
    return historyData || getFindingsToolCalls(findingsId, nodeId)
  }

  // 从历史数据或上下文中获取代理工具显示数据
  const getAgentToolDisplay = (nodeId: string, findingsId: string) => {
    const historyData = historyDisplays?.[nodeId]
    return historyData || getFindingsAgentToolDisplay(findingsId, nodeId)
  }

  const handleNodeClick = (data: FindingsNodeData) => {
    setSelectedNode(data)
    setPanelVisible(true)
  }

  const handlePanelClose = () => {
    setPanelVisible(false)
    setSelectedNode(null)
  }

  // 使用 useMemo 缓存选中节点的数据，避免不必要的重新计算
  const selectedToolCalls = useMemo(() => {
    if (!selectedNode) return []
    return getToolCalls(selectedNode.nodeId, selectedNode.findingsId)
  }, [selectedNode, historyTools])

  const selectedAgentToolDisplay = useMemo(() => {
    if (!selectedNode) return []
    return getAgentToolDisplay(selectedNode.nodeId, selectedNode.findingsId)
  }, [selectedNode, historyDisplays])
  useEffect(() => {
    setIsExpanded(expanded)
  }, [expanded])
  return (
    <div className="w-full flex flex-col">
      {/* 收起/展开切换按钮 */}
      {isCompleted && (
        <span
          className="ml-10 p-0 text-[var(--ant-color-text-secondary)] text-xs mr-2 inline-flex items-center cursor-pointer select-none mb-2"
          onClick={() => setIsExpanded((v) => !v)}
        >
          {isExpanded ? <DownOutlined /> : <RightOutlined />}
          <span className="ml-1">{isExpanded ? 'Hide Findings' : 'View Findings'}</span>
        </span>
      )}

      {/* 展开状态下显示完整拓扑图 */}
      {(isPending || isExpanded) && (
        <>
          {/* 左侧拓扑图区域 */}
          <div className="w-full flex gap-2 items-start justify-center">
            <div className="w-full h-full transition-all duration-300">
              {/* 外部容器：仅保留顶部细进度条的加载态 */}
              <div
                className="relative h-[500px] rounded-lg z-10 ring-1 ring-inset ring-emerald-200 transition-all duration-200"
                aria-busy={isPending}
              >
                {/* 顶部细进度条（无限流动） */}
                {isPending && (
                  <div className="pointer-events-none absolute left-0 top-0 h-[2px] w-full overflow-hidden rounded-t-lg">
                    <div className="h-full w-1/3 bg-emerald-500 animate-loading-bar" />
                  </div>
                )}

                {/* 拓扑内容 */}
                <div className="relative h-full">
                  <FindingsTopology
                    data={message.findingsData}
                    direction="TB"
                    onNodeClick={handleNodeClick}
                    findingsId={message.id}
                  />
                </div>
              </div>
            </div>

            {/* 右侧详情面板 */}
            {panelVisible && (
              <div className="w-1/2 h-[500px]">
                <NodeDetailPanel
                  title={selectedNode?.title || ''}
                  content={selectedNode?.content || ''}
                  visible={panelVisible}
                  onClose={handlePanelClose}
                  toolCalls={selectedToolCalls}
                  agentToolDisplay={selectedAgentToolDisplay}
                />
              </div>
            )}
          </div>
        </>
      )}

      {/* Result 节点 - 仅在展开状态下显示 */}
      {isCompleted && isExpanded && message.findingsData.result && (
        <div className="w-full flex justify-center">
          <div className="w-2/3 relative flex pt-8 overflow-hidden">
            {/* 有向箭头线 */}
            <svg
              className="absolute"
              style={{
                top: '-20px',
                left: '50%',
                transform: 'translateX(-50%)',
                width: '2px',
                overflow: 'visible',
              }}
            >
              <defs>
                <marker
                  id="arrowhead-result"
                  markerWidth="10"
                  markerHeight="10"
                  refX="5"
                  refY="5"
                  orient="auto"
                >
                  <polygon points="0 0, 10 5, 0 10" fill="rgb(16, 185, 129)" />
                </marker>
              </defs>
              <line
                x1="1"
                y1="0"
                x2="1"
                y2="40"
                stroke="rgb(16, 185, 129)"
                strokeWidth="1.5"
                markerEnd="url(#arrowhead-result)"
              />
            </svg>

            <div
              className={[
                'relative select-none max-w-full',
                'rounded-2xl px-3.5 py-3',
                'ring-1 ring-inset transition-all duration-200',
                'ring-emerald-200',
                'shadow-[0_6px_24px_-8px_rgba(16,185,129,0.35)]',
              ].join(' ')}
            >
              <div className="gap-3 text-center">
                <div className="text-base font-bold leading-snug flex flex-col">
                  {message.findingsData.result?.title ||
                    message.findingsData.result?.content ||
                    'Result'}
                  {message.findingsData.result?.content && (
                    <div className="text-left">
                      <MarkdownRender
                        content={message.findingsData.result?.content}
                        theme="light"
                      />
                    </div>
                  )}
                </div>
              </div>
            </div>
          </div>
        </div>
      )}

      {/* 局部样式：仅用于顶部进度条动画 */}
      <style>{`
        @keyframes loading-bar {
          0% { transform: translateX(-120%); }
          100% { transform: translateX(400%); }
        }
        .animate-loading-bar {
          animation: loading-bar 2.2s cubic-bezier(.22,.61,.36,1) infinite;
        }
      `}</style>
    </div>
  )
}

export default FindingsMessageItem
