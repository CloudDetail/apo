import { Button, Spin, Tooltip } from 'antd'
import {
  CheckCircleOutlined,
  ClockCircleOutlined,
  CloseCircleOutlined,
  CopyOutlined,
  PauseCircleOutlined,
  RobotOutlined,
  UserOutlined,
} from '@ant-design/icons'
import { useEffect, useRef, useState } from 'react'
import ReactMarkdown from 'react-markdown'
import remarkBreaks from 'remark-breaks'
import remarkGfm from 'remark-gfm'
import type {
  HermesCheckpoint,
  HermesRunStatusTimelineItem,
  HermesTimelineItem,
} from '../client/types.ts'
import styles from '../HermesAgentWorkspace.module.scss'
import CheckpointCard from './CheckpointCard'
import PlanCard from './PlanCard'
import ToolActivity from './ToolActivity'

interface ChatMessagesProps {
  timeline: HermesTimelineItem[]
  reasoning: string
  busy: boolean
  activeCheckpoint: HermesCheckpoint | null
  hydrationVersion: number
  onCheckpointRespond: (answer: string) => Promise<void>
  onExample: (text: string) => void
}

const examples = [
  {
    title: '分析最近出现延时的服务',
    detail: '最近 30 分钟 · 服务与实例',
    prompt: '分析最近 30 分钟内的出现延时的服务，并定位可能原因。',
  },
  {
    title: '查询服务 RED 指标',
    detail: '最近 1 小时 · 请求量/错误/耗时',
    prompt: '查询最近 1 小时的服务 RED 指标，找出错误率或耗时异常的服务。',
  },
  {
    title: '查看错误日志',
    detail: '最近 30 分钟 · 错误与异常堆栈',
    prompt: '检查最近 30 分钟的错误日志，按服务汇总主要异常并给出排查建议。',
  },
  {
    title: '查询主机资源',
    detail: '最近 1 小时 · 主机资源',
    prompt: '查询最近一小时的主机资源使用情况，并总结趋势。',
  },
]

export default function ChatMessages({
  timeline,
  reasoning,
  busy,
  activeCheckpoint,
  hydrationVersion,
  onCheckpointRespond,
  onExample,
}: ChatMessagesProps) {
  const viewportRef = useRef<HTMLDivElement | null>(null)
  const previousHydrationVersion = useRef(-1)
  const stickToBottom = useRef(true)

  useEffect(() => {
    const viewport = viewportRef.current
    if (!viewport) return
    const restored = hydrationVersion > 0 && previousHydrationVersion.current !== hydrationVersion
    previousHydrationVersion.current = hydrationVersion

    requestAnimationFrame(() => {
      if (restored) {
        const answers = viewport.querySelectorAll<HTMLElement>('[data-final-answer="true"]')
        const runningStatus = viewport.querySelector<HTMLElement>('[data-run-active="true"]')
        const target = answers.item(answers.length - 1) || runningStatus
        if (target) target.scrollIntoView({ block: 'start' })
        else viewport.scrollTop = 0
        stickToBottom.current = false
        return
      }
      if (stickToBottom.current) {
        viewport.scrollTo({ top: viewport.scrollHeight, behavior: 'smooth' })
      }
    })
  }, [hydrationVersion, reasoning, timeline])

  const onScroll = () => {
    const viewport = viewportRef.current
    if (!viewport) return
    stickToBottom.current = viewport.scrollHeight - viewport.scrollTop - viewport.clientHeight < 120
  }

  return (
    <div ref={viewportRef} className={styles.messageViewport} onScroll={onScroll}>
      {timeline.length === 0 ? (
        <div className={styles.emptyConversation}>
          <div className={styles.emptyHero}>
            <div className={styles.emptyHeroIcon}>
              <RobotOutlined />
            </div>
            <div>
              <h2>从一个真实运维任务开始</h2>
              <p>请尽量说明对象、时间范围和异常现象，我会按步骤分析并给出可继续追问的结论。</p>
            </div>
          </div>
          <div className={styles.exampleQuestions}>
            {examples.map((example) => (
              <button
                type="button"
                key={example.title}
                className={styles.exampleQuestion}
                onClick={() => onExample(example.prompt)}
              >
                <strong>{example.title}</strong>
                <span>{example.detail}</span>
              </button>
            ))}
          </div>
          <p className={styles.emptyHint}>
            可查询的数据范围以当前环境已接入的服务、主机和可观测数据为准。
          </p>
        </div>
      ) : (
        <div className={styles.messageThread}>
          {renderTimeline(timeline, activeCheckpoint, onCheckpointRespond)}
          {reasoning.trim() && <ToolActivity activities={[]} reasoning={reasoning} live />}
          {busy && !reasoning.trim() && !hasActiveRunStatus(timeline) && (
            <div className={styles.inlineLoading}>
              <Spin size="small" />
              正在等待任务进度…
            </div>
          )}
        </div>
      )}
    </div>
  )
}

function renderTimeline(
  timeline: HermesTimelineItem[],
  activeCheckpoint: HermesCheckpoint | null,
  onCheckpointRespond: (answer: string) => Promise<void>,
) {
  const nodes = []
  for (let index = 0; index < timeline.length; index += 1) {
    const item = timeline[index]
    if (item.kind === 'plan') {
      nodes.push(<PlanCard key={item.id} item={item} />)
      continue
    }
    if (item.kind === 'activity') {
      const activities = [item.activity]
      const runId = item.runId
      let cursor = index + 1
      while (
        cursor < timeline.length &&
        timeline[cursor].kind === 'activity' &&
        timeline[cursor].runId === runId
      ) {
        const next = timeline[cursor]
        if (next.kind === 'activity') activities.push(next.activity)
        cursor += 1
      }
      nodes.push(
        <ToolActivity key={`activity-group-${item.id}`} activities={activities} reasoning="" />,
      )
      index = cursor - 1
      continue
    }
    if (item.kind === 'message') {
      const message = item.message
      const assistant = message.role === 'assistant'
      nodes.push(
        <article
          key={item.id}
          className={`${styles.messageRow} ${
            message.role === 'user' ? styles.userMessageRow : styles.assistantMessageRow
          }`}
          data-final-answer={assistant && message.complete ? 'true' : undefined}
        >
          <div className={styles.messageAvatar}>
            {message.role === 'user' ? <UserOutlined /> : <RobotOutlined />}
          </div>
          <div className={assistant ? styles.assistantMessageBody : styles.userMessageBubble}>
            {assistant && (
              <div className={styles.assistantMessageHeader}>
                <span>Hermes 分析</span>
                <time>{formatTime(item.occurredAt)}</time>
              </div>
            )}
            {assistant && !message.content && message.complete === false ? (
              <span className={styles.thinkingText}>
                <Spin size="small" />
                正在组织回答…
              </span>
            ) : assistant ? (
              <div className={styles.markdownBody}>
                <ReactMarkdown remarkPlugins={[remarkGfm, remarkBreaks]}>
                  {message.content}
                </ReactMarkdown>
              </div>
            ) : (
              <>
                <div className={styles.userMessageContent}>{message.content}</div>
                <time className={styles.userMessageTime}>{formatTime(item.occurredAt)}</time>
              </>
            )}
            {assistant && message.content && (
              <div className={styles.assistantMessageActions}>
                <Tooltip title="复制回答">
                  <Button
                    type="text"
                    size="small"
                    icon={<CopyOutlined />}
                    onClick={() => void navigator.clipboard.writeText(message.content)}
                  >
                    复制
                  </Button>
                </Tooltip>
              </div>
            )}
          </div>
        </article>,
      )
      continue
    }
    if (item.kind === 'checkpoint') {
      nodes.push(
        <CheckpointCard
          key={item.id}
          checkpoint={item.checkpoint}
          resolved={item.resolved}
          responseText={item.responseText}
          active={!item.resolved && activeCheckpoint?.requestId === item.checkpoint.requestId}
          onRespond={onCheckpointRespond}
        />,
      )
      continue
    }
    if (item.kind === 'run_status') nodes.push(<RunStatus key={item.id} item={item} />)
  }
  return nodes
}

function RunStatus({ item }: { item: HermesRunStatusTimelineItem }) {
  const [now, setNow] = useState(Date.now())
  const active = ['running', 'stopping'].includes(item.status)

  useEffect(() => {
    if (!active) return undefined
    const timer = window.setInterval(() => setNow(Date.now()), 1_000)
    return () => window.clearInterval(timer)
  }, [active])

  const elapsed =
    item.durationMs ?? Math.max(0, now - new Date(item.startedAt || item.occurredAt).getTime())

  return (
    <section
      className={`${styles.runStatusCard} ${styles[`runStatus_${item.status}`] || ''}`}
      data-run-active={active ? 'true' : undefined}
    >
      <span className={styles.runStatusIcon}>{runStatusIcon(item.status)}</span>
      <div className={styles.runStatusContent}>
        <strong>{item.label}</strong>
        {item.detail && <span>{item.detail}</span>}
      </div>
      <span className={styles.runDuration}>
        <ClockCircleOutlined /> {formatDuration(elapsed)}
      </span>
    </section>
  )
}

function runStatusIcon(status: HermesRunStatusTimelineItem['status']) {
  if (status === 'running' || status === 'stopping') return <Spin size="small" />
  if (status === 'completed') return <CheckCircleOutlined />
  if (status === 'failed') return <CloseCircleOutlined />
  return <PauseCircleOutlined />
}

function formatDuration(durationMs: number) {
  const seconds = Math.max(0, Math.round(durationMs / 1_000))
  if (seconds < 60) return `${seconds} 秒`
  const minutes = Math.floor(seconds / 60)
  return `${minutes} 分 ${seconds % 60} 秒`
}

function formatTime(value: string) {
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return ''
  return date.toLocaleTimeString('zh-CN', { hour: '2-digit', minute: '2-digit' })
}

function hasActiveRunStatus(timeline: HermesTimelineItem[]) {
  return timeline.some(
    (item) => item.kind === 'run_status' && ['running', 'stopping'].includes(item.status),
  )
}
