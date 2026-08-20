import type { ReactNode } from 'react'
import { Alert, Card, Descriptions, Empty, List, Space, Tag, Typography } from 'antd'
import type { CapabilityProjection, WorkflowRun, WorkflowStep, WorkCapability } from '../client/types.ts'
import styles from '../pages/index.module.scss'

const { Text, Paragraph } = Typography

const LABELS: Record<string, string> = {
  active: '自动匹配开启',
  enabled: '自动匹配开启',
  disabled: '自动匹配关闭',
  archived: '已归档',
  unknown: '状态未知',
  succeeded: '成功',
  completed: '已完成',
  failed: '失败',
  cancelled: '已取消',
  running: '运行中',
  pending: '处理中',
  in_progress: '进行中',
  partial: '部分完成',
  blocked: '已阻断',
  skipped: '已跳过',
  complete: '完整',
  none: '无结果',
  reused: '复用',
  improved: '改进',
  created: '新建',
  disabled_reuse: '复用已停用',
  workflow_reuse: '任务复用',
  agent_emitted: '动态规划',
  pending_match: '等待匹配',
  verified_read: '已验证读取',
  verified_write: '已验证写入',
  declared: '声明信息',
  high: '高',
  medium: '中',
  low: '低',
}

export function statusLabel(value?: unknown): string {
  const normalized = String(value || 'unknown')
  return LABELS[normalized] || normalized
}

export function statusColor(value?: unknown): string {
  const normalized = String(value || 'unknown')
  if (['active', 'enabled', 'succeeded', 'completed', 'complete', 'reused', 'improved'].includes(normalized)) return 'success'
  if (['running', 'pending', 'in_progress', 'partial'].includes(normalized)) return 'processing'
  if (['disabled', 'archived', 'failed', 'blocked', 'cancelled', 'none'].includes(normalized)) return 'error'
  return 'default'
}

export function StatusTag({ value, label }: { value?: unknown; label?: string }) {
  return <Tag color={statusColor(value)}>{label || statusLabel(value)}</Tag>
}

export function formatDate(value?: unknown): string {
  if (!value) return '未提供'
  const date = new Date(String(value))
  return Number.isNaN(date.getTime()) ? String(value) : date.toLocaleString('zh-CN')
}

export function formatRate(value?: number | null): string {
  return typeof value === 'number' && Number.isFinite(value) ? `${Math.round(value * 100)}%` : '未提供'
}

export function formatCount(value?: number | null): string {
  return typeof value === 'number' && Number.isFinite(value) ? String(value) : '未提供'
}

export function blockerText(item: WorkCapability['actionBlockers'][number]): string {
  if (typeof item === 'string') return item
  return item.reason || item.message || item.code || '服务端未说明阻断原因。'
}

export function SectionCard({
  title,
  extra,
  children,
  className,
}: {
  title: string
  extra?: ReactNode
  children: ReactNode
  className?: string
}) {
  return (
    <Card className={`${styles.sectionCard} ${className || ''}`} title={title} extra={extra} size="small">
      {children}
    </Card>
  )
}

export function StepList({ steps }: { steps: WorkflowStep[] }) {
  if (!steps.length) return <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="服务端未返回稳定步骤" />
  return (
    <ol className={styles.stepList}>
      {steps.map((step, index) => (
        <li className={styles.stepItem} key={step.id}>
          <span className={styles.stepNumber}>{index + 1}</span>
          <div className={styles.stepContent}>
            <strong>{step.purpose}</strong>
            <Paragraph className={styles.stepDescription}>预期输出：{step.expectedOutput}</Paragraph>
          </div>
        </li>
      ))}
    </ol>
  )
}

export function RunSummary({ run, onOpen }: { run: WorkflowRun; onOpen?: (run: WorkflowRun) => void }) {
  return (
    <div className={styles.runItem}>
      <div className={styles.runMain}>
        <strong>{run.title || run.runId}</strong>
        <Text type="secondary">
          {run.relationship ? `${statusLabel(run.relationship)} · ` : ''}
          {run.version ? `v${run.version} · ` : ''}
          {formatDate(run.createdAt || run.updatedAt)}
        </Text>
        {run.sessionId ? <Text type="secondary">会话：{run.sessionId}</Text> : null}
      </div>
      <Space size={8}>
        <StatusTag value={run.status} />
        {onOpen ? <button type="button" className={styles.textButton} onClick={() => onOpen(run)}>查看运行</button> : null}
      </Space>
    </div>
  )
}

export function RunProjectionPanel({ projection }: { projection: CapabilityProjection | null }) {
  if (!projection) return <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="暂无运行投影" />
  const nodes = Array.isArray(projection.nodes) ? projection.nodes : []
  const evidence = Array.isArray(projection.evidence) ? projection.evidence : []
  const gaps = Array.isArray(projection.gaps) ? projection.gaps : []
  const failures = Array.isArray(projection.failures) ? projection.failures : []
  const result = projection.result
  const executionMode = projection.workflowRun?.executionMode || projection.workflow?.executionMode
  return (
    <div className={styles.projectionPanel}>
      <Descriptions size="small" column={1} bordered>
        <Descriptions.Item label="运行状态"><StatusTag value={projection.status} /></Descriptions.Item>
        <Descriptions.Item label="执行计划">
          {projection.hasPlan === undefined ? '未提供' : projection.hasPlan ? `已生成${projection.revision ? ` · 修订 ${projection.revision}` : ''}` : '尚未生成'}
        </Descriptions.Item>
        {executionMode ? <Descriptions.Item label="执行方式">{statusLabel(executionMode)}</Descriptions.Item> : null}
        {projection.workflow ? (
          <Descriptions.Item label="关联任务">
            {projection.workflow.name}{projection.workflow.version ? ` · v${projection.workflow.version}` : ''}
          </Descriptions.Item>
        ) : null}
      </Descriptions>
      {result ? (
        <Card size="small" title="运行结果" className={styles.projectionSubcard}>
          <Space direction="vertical" size={4}>
            {result.summary ? <Text>{result.summary}</Text> : null}
            {result.confidence ? <Text type="secondary">置信度：{statusLabel(result.confidence)}</Text> : null}
            {result.nextAction ? <Text type="secondary">后续动作：{result.nextAction}</Text> : null}
          </Space>
        </Card>
      ) : null}
      {nodes.length ? (
        <Card size="small" title="步骤状态" className={styles.projectionSubcard}>
          <List
            size="small"
            dataSource={nodes}
            renderItem={(node) => (
              <List.Item key={node.id}>
                <Space direction="vertical" size={0}>
                  <Space size={6} wrap>
                    <strong>{node.title}</strong>
                    <StatusTag value={node.status} />
                    {node.requiresUserAction ? <Tag color="warning">等待用户操作</Tag> : null}
                  </Space>
                  {node.summary ? <Text type="secondary">{node.summary}</Text> : null}
                  {node.facts.length ? <Text type="secondary">事实：{node.facts.join('；')}</Text> : null}
                </Space>
              </List.Item>
            )}
          />
        </Card>
      ) : null}
      {evidence.length ? (
        <Card size="small" title="Evidence 摘要" className={styles.projectionSubcard}>
          <List
            size="small"
            dataSource={evidence}
            renderItem={(item) => (
              <List.Item key={item.id}>
                <Space direction="vertical" size={0}>
                  <Space size={6} wrap>
                    <strong>{item.title}</strong>
                    {item.trust ? <Tag>{statusLabel(item.trust)}</Tag> : null}
                  </Space>
                  <Text type="secondary">{item.summary || '服务端未提供摘要'}</Text>
                  {item.receiptId ? <Text type="secondary">Receipt：{item.receiptId}</Text> : null}
                </Space>
              </List.Item>
            )}
          />
        </Card>
      ) : null}
      {gaps.length ? <Alert className={styles.projectionAlert} type="warning" showIcon message="运行存在数据缺口" description={<ul>{gaps.map((gap) => <li key={gap}>{gap}</li>)}</ul>} /> : null}
      {failures.length ? <Alert className={styles.projectionAlert} type="error" showIcon message="步骤失败信息" description={<ul>{failures.map((failure) => <li key={failure}>{failure}</li>)}</ul>} /> : null}
    </div>
  )
}
