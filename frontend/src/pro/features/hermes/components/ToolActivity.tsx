import { Collapse, Spin, Tag } from 'antd'
import {
  BulbOutlined,
  CheckCircleOutlined,
  CloseCircleOutlined,
  ToolOutlined,
} from '@ant-design/icons'
import type { HermesActivity } from '../client/types.ts'
import styles from '../HermesAgentWorkspace.module.scss'

export default function ToolActivity({
  activities,
  reasoning,
  live = false,
}: {
  activities: HermesActivity[]
  reasoning: string
  live?: boolean
}) {
  if (!activities.length && !reasoning.trim()) return null

  const visibleActivities = [
    ...activities,
    ...(reasoning.trim()
      ? [{ id: 'live-reasoning', kind: 'commentary' as const, text: reasoning }]
      : []),
  ]
  const tools = visibleActivities.filter((item) => item.kind === 'tool')
  const completed = tools.filter((item) => item.status === 'completed').length
  const failed = tools.filter((item) => item.status === 'failed').length
  const running = tools.filter((item) => item.status === 'running').length
  const durationMs = tools.reduce((sum, item) => sum + (item.durationMs || 0), 0)

  const items = [
    {
      key: 'execution-summary',
      label: (
        <span className={styles.activitySummaryLabel}>
          {running > 0 || live ? <Spin size="small" /> : <ToolOutlined />}
          <strong>{running > 0 || live ? '执行过程正在更新' : '查看执行过程'}</strong>
          <span>{summaryText(visibleActivities.length, completed, failed, running, durationMs)}</span>
        </span>
      ),
      children: (
        <div className={styles.activityList}>
          {visibleActivities.map((activity) => (
            <div key={activity.id} className={styles.activityItem}>
              <div className={styles.activityTitle}>
                {activity.kind === 'commentary' ? (
                  <BulbOutlined />
                ) : activity.status === 'running' ? (
                  <Spin size="small" />
                ) : activity.status === 'failed' ? (
                  <CloseCircleOutlined className={styles.failedIcon} />
                ) : (
                  <CheckCircleOutlined className={styles.successIcon} />
                )}
                <strong>{activity.kind === 'commentary' ? '分析过程' : activity.name}</strong>
                {activity.kind === 'tool' && (
                  <Tag bordered={false}>{toolStatusText(activity.status)}</Tag>
                )}
              </div>
              {activity.kind === 'commentary' ? (
                <div className={styles.activityText}>{activity.text}</div>
              ) : (
                <div className={styles.toolDetail}>
                  {activity.context && <p>{activity.context}</p>}
                  {activity.args && (
                    <details>
                      <summary>调用参数</summary>
                      <pre>{safeJson(activity.args)}</pre>
                    </details>
                  )}
                  {activity.result !== undefined && (
                    <details>
                      <summary>执行结果</summary>
                      <pre>{safeJson(activity.result)}</pre>
                    </details>
                  )}
                </div>
              )}
            </div>
          ))}
        </div>
      ),
    },
  ]

  return (
    <section className={styles.activityPanel}>
      <Collapse ghost size="small" items={items} />
    </section>
  )
}

function summaryText(
  total: number,
  completed: number,
  failed: number,
  running: number,
  durationMs: number,
) {
  const parts = [`${total} 项`]
  if (completed) parts.push(`${completed} 项完成`)
  if (failed) parts.push(`${failed} 项失败`)
  if (running) parts.push(`${running} 项进行中`)
  if (durationMs) parts.push(`累计 ${formatDuration(durationMs)}`)
  return parts.join(' · ')
}

function formatDuration(durationMs: number) {
  if (durationMs < 1_000) return `${Math.round(durationMs)} 毫秒`
  const seconds = Math.round(durationMs / 1_000)
  return seconds < 60 ? `${seconds} 秒` : `${Math.floor(seconds / 60)} 分 ${seconds % 60} 秒`
}

function toolStatusText(status: 'running' | 'completed' | 'failed') {
  if (status === 'running') return '执行中'
  if (status === 'failed') return '失败'
  return '已完成'
}

function safeJson(value: unknown): string {
  try {
    const content = JSON.stringify(value, null, 2)
    return content.length > 6_000 ? `${content.slice(0, 6_000)}\n…内容已截断` : content
  } catch {
    return String(value)
  }
}
