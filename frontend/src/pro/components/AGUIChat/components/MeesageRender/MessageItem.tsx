import React, { useEffect, useMemo, useState, useCallback } from 'react'
import { Alert, Spin } from 'antd'
import { RightOutlined, DownOutlined, LoadingOutlined } from '@ant-design/icons'
import { ExtendedAGUIMessage } from 'src/pro/contexts/AGUIContext'
import MarkdownRender from 'src/core/components/Markdown/MarkdownRender'
import ToolRender from '../AGUITool'
import FinalMessageSuffix from './MessageSuffix'
import styles from '../../AGUIChat.module.scss'
import ShinyText from 'src/pro/components/Animation/ShinyText'
import AgentToolDisplay from '../AgentToolDisplay'
import ToolTag from '../ToolRender/ToolTag'
import { AnimatePresence, motion, useReducedMotion } from 'framer-motion'
import FindingsMessageItem from './FindingsMessageItem'
import { shouldRenderFinalMessageSuffix } from 'src/pro/api/conversationHistoryView'

export type ToolKind = 'toolCall' | 'agentTool'

export function stableId(obj: any, type: ToolKind, idx: number): string {
  const name = obj?.toolName || obj?.name || 'unknown'
  return `${type}-${name}-${idx}`
}
export function ScaleCollapse({
  expanded = true,
  children,
  className = '',
}: {
  expanded: boolean
  children: React.ReactNode
  className?: string
}) {
  const prefersReduce = useReducedMotion()
  const duration = prefersReduce ? 0.32 : 0.32

  return (
    <AnimatePresence initial={false}>
      {expanded && (
        <motion.div
          key="scale"
          initial={{ opacity: 0, scaleY: 0.9 }}
          animate={{ opacity: 1, scaleY: 1 }}
          exit={{ opacity: 0, scaleY: 0.9 }}
          transition={{ duration, ease: [0.22, 1, 0.36, 1] }}
          className={`origin-top overflow-hidden ${className}`}
          style={{ willChange: 'transform, opacity', contain: 'paint' }}
        >
          {children}
        </motion.div>
      )}
    </AnimatePresence>
  )
}
type ToolPointer = { id: string; type: ToolKind; index: number; toolName?: string }
const AnalysisMessageItem = ({ message }: { message: ExtendedAGUIMessage }) => {
  const isStreaming = message.status === 'pending'
  // 分析文本：pending 展开，complete 折叠
  const [expanded, setExpanded] = useState<boolean>(isStreaming)

  // 工具数组
  const toolCalls = Array.isArray(message.toolCalls) ? message.toolCalls : []
  const agentTools = Array.isArray(message.agentToolDisplay) ? message.agentToolDisplay : []

  // 先计算 items（用于初始化 expandedIds）
  const items: ToolPointer[] = [
    ...toolCalls.map((t, i) => ({
      id: stableId(t, 'toolCall', i),
      type: 'toolCall' as const,
      toolName: t.name,
      index: i,
    })),
    ...agentTools.map((t, i) => ({
      id: stableId(t, 'agentTool', i),
      type: 'agentTool' as const,
      toolName: t.toolName,
      index: i,
    })),
  ]

  // 默认展开：pending 时默认全量展开；非 pending 默认空
  const [expandedIds, setExpandedIds] = useState<Set<string>>(() =>
    isStreaming ? new Set(items.map((it) => it.id)) : new Set(),
  )
  const [manualOpenIds, setManualOpenIds] = useState<Set<string>>(() => new Set())
  // 仅在 pending 状态下，items 变化时把“新增的工具 id”补充进展开集合
  useEffect(() => {
    if (message.status !== 'pending') return
    setExpandedIds((prev) => {
      const next = new Set(prev)
      let changed = false
      for (const it of items) {
        if (!next.has(it.id)) {
          next.add(it.id)
          changed = true
        }
      }
      return changed ? next : prev
    })
  }, [items, message.status])

  // 根据流式状态控制分析文本展开/折叠；complete 时同时收起所有工具
  useEffect(() => {
    if (message.status === 'pending') {
      setExpanded(true)
      // 如果想在重新进入 pending 时再次“全量展开”，取消下行注释：
      // setExpandedIds(new Set(items.map((it) => it.id)))
    } else {
      setExpanded(false)
      setExpandedIds(
        () =>
          new Set(
            items
              .filter((it) => it.type === 'agentTool' && it.toolName === 'result')
              .map((it) => it.id),
          ),
      )
    }
  }, [message.status])

  // 仅用于显示标签文案
  const toolMap = (() => {
    const map = new Map<string, any>()
    for (let i = 0; i < (message.toolCalls?.length || 0); i++) {
      map.set(stableId(message.toolCalls[i], 'toolCall', i), message.toolCalls[i])
    }
    for (let i = 0; i < (message.agentToolDisplay?.length || 0); i++) {
      map.set(stableId(message.agentToolDisplay[i], 'agentTool', i), message.agentToolDisplay[i])
    }
    return map
  })()

  /** 标签点击：三态切换
   * 1) 未展开 -> 展开 + 标记为手动展开
   * 2) 自动展开(未手动) -> 保持展开 + 升级为手动展开（open=false -> open=true）
   * 3) 手动展开 -> 收起（两个集合都移除）
   */
  const onToggle = useCallback((id: string) => {
    setExpandedIds((prevExp) => {
      const nextExp = new Set(prevExp)
      setManualOpenIds((prevMan) => {
        const nextMan = new Set(prevMan)
        const isExp = nextExp.has(id)
        const isMan = nextMan.has(id)

        if (!isExp) {
          // 未展开 -> 展开 + 手动
          nextExp.add(id)
          nextMan.add(id)
        } else if (!isMan) {
          // 自动展开 -> 升级为手动（保持展开）
          nextMan.add(id)
        } else {
          // 手动展开 -> 收起
          nextExp.delete(id)
          nextMan.delete(id)
        }
        return nextMan
      })
      return nextExp
    })
  }, [])

  /** 只渲染“已展开”的详情，渲染面积最小化 */
  const expandedItems = useMemo(
    () => items.filter((it) => expandedIds.has(it.id)),
    [items, expandedIds],
  )

  const hasTools = items.length > 0

  return (
    <div className="ml-10 text-[var(--ant-color-text-secondary)] italic">
      {/* 头部行：仅在非流式时显示 ToolTag（你原本的逻辑保持不变） */}
      {!isStreaming && (
        <div className="inline-flex flex-wrap items-center gap-2">
          <span
            className="p-0 text-[var(--ant-color-text-secondary)] text-xs mr-2 inline-flex items-center cursor-pointer select-none"
            onClick={() => setExpanded((v) => !v)}
          >
            {expanded ? <DownOutlined /> : <RightOutlined />}
            <span className="ml-1">{expanded ? 'Hide Analysis' : 'View Analysis'}</span>
          </span>

          {hasTools && (
            <ToolTag
              items={items}
              toolMap={toolMap}
              expandedIds={expandedIds}
              onToggle={onToggle}
              maxVisible={3}
            />
          )}
        </div>
      )}

      {/* 内容区：流式与非流式分开处理 */}
      {isStreaming && (
        <>
          <div className=" text-gray-400 pr-2 py-[2px] mr-1 rounded-lg text-xs flex items-center ">
            <Spin
              indicator={<LoadingOutlined spin />}
              size="small"
              className="mr-1 text-[var(--ant-color-text-secondary)]"
            />
            <ShinyText text="Analyzing" />
          </div>
        </>
      )}
      <ScaleCollapse expanded={expanded}>
        <div className="mt-1 whitespace-pre-wrap text-xs leading-5">{message.content}</div>
      </ScaleCollapse>
      {/* 工具详情：complete 后默认都收起；用户可手动再展开 */}
      {expandedItems.length > 0 && (
        <div className="mt-2 w-full space-y-2">
          {expandedItems.map((desc) => {
            const obj = toolMap.get(desc.id)
            if (!obj) return null
            const isManualOpen = manualOpenIds.has(desc.id)
            return desc.type === 'toolCall' ? (
              <div key={`detail-${desc.id}`} className="rounded-md">
                <ToolRender toolCall={obj} open={isManualOpen} />
              </div>
            ) : (
              <div key={`detail-${desc.id}`} className="rounded-md">
                <AgentToolDisplay
                  toolName={obj.toolName}
                  title={obj.title}
                  content={obj.content}
                  open={isManualOpen}
                />
              </div>
            )
          })}
        </div>
      )}
    </div>
  )
}

const MessageItem = ({
  message,
  activeThreadId,
}: {
  message: ExtendedAGUIMessage
  activeThreadId?: string
}) => {
  const messageClass = `${styles.message} ${message.role === 'user' ? styles.user : styles.assistant}`
  return (
    <div key={message.id} className={messageClass}>
      {message.hidden ? (
        <AnalysisMessageItem message={message} />
      ) : message.type === 'findings' ? (
        // findings 类型直接显示 FindingsMessageItem，不显示 message.content
        <FindingsMessageItem message={message} />
      ) : (
        <div className={styles.messageContent}>
          <div className={`${styles.messageText} ${message?.type === 'analysisResult' ? '' : ''}`}>
            {message.role === 'assistant' && message.content && !message.hidden ? (
              <MarkdownRender content={message.content} theme="light" />
            ) : (
              message.content
            )}
          </div>
        </div>
      )}
      {/* 非隐藏消息：保持原逻辑（如需也改成 ToolTag + 父渲染，告诉我） */}
      {!message.hidden && message.toolCalls?.length > 0 && (
        <div className="message-tool-calls ml-6 ">
          {message.toolCalls.map((toolCall: any) => (
            <ToolRender key={toolCall.id} toolCall={toolCall} />
          ))}
        </div>
      )}

      {!message.hidden && message.agentToolDisplay?.length > 0 && (
        <div className="message-tool-calls ml-6">
          {message.agentToolDisplay.map((toolCall: any, index: number) => (
            <AgentToolDisplay
              key={toolCall.toolName + index}
              toolName={toolCall.toolName}
              title={toolCall.title}
              content={toolCall.content}
            />
          ))}
        </div>
      )}

      {message.errorMessage && (
        <Alert
          message={message.errorMessage}
          type="error"
          className="mx-4 mt-2 text-[var(--ant-color-text-secondary)]"
        />
      )}
      {shouldRenderFinalMessageSuffix(message) && (
        <FinalMessageSuffix message={message} activeThreadId={activeThreadId} />
      )}
    </div>
  )
}

export default MessageItem
