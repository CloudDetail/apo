import { Alert, Button, Collapse, Empty, Modal, Tag } from 'antd'
import { BarChartOutlined, CodeOutlined, DeploymentUnitOutlined, ExpandOutlined, FileTextOutlined } from '@ant-design/icons'
import { useState } from 'react'
import Metrics from '../../../components/AGUIChat/components/ToolRender/Metrics'
import NodeLogs from '../../../components/AGUIChat/components/EvidencePanel/NodeLogs'
import Topology from '../../../../core/components/ReactFlow/Topology'
import type {
  ApoLogRecord,
  ApoMetricItem,
  ApoTopologyEdge,
  ApoTopologyNode,
  HermesEvidence,
} from '../client/types.ts'
import styles from '../HermesAgentWorkspace.module.scss'

export default function EvidenceRenderer({ evidence }: { evidence: HermesEvidence[] }) {
  if (!evidence.length) return null
  return (
    <div className={styles.evidenceList}>
      {evidence.map((item, index) => (
        <EvidenceCard key={item.id || `${item.kind}-${index}`} evidence={item} />
      ))}
    </div>
  )
}

function EvidenceCard({ evidence }: { evidence: HermesEvidence }) {
  const content = renderEvidence(evidence)
  return (
    <section className={styles.evidenceCard}>
      <div className={styles.evidenceHeading}>
        <span>{kindIcon(evidence.kind)}</span>
        <div>
          <strong>{evidence.title || '查询依据'}</strong>
          {evidence.summary && <p>{evidence.summary}</p>}
        </div>
        <Tag bordered={false}>{kindLabel(evidence.kind)}</Tag>
      </div>
      {content}
    </section>
  )
}

function renderEvidence(evidence: HermesEvidence) {
  if (evidence.kind === 'apo_metrics') {
    if (!isMetrics(evidence.data)) return invalidEvidence(evidence, '指标数据结构不符合展示契约')
    return evidence.data.length ? <Metrics args={evidence.data} /> : <Empty description="暂无指标数据" />
  }
  if (evidence.kind === 'apo_logs') {
    if (!isLogs(evidence.data)) return invalidEvidence(evidence, '日志数据结构不符合展示契约')
    return <NodeLogs logs={evidence.data} />
  }
  if (evidence.kind === 'apo_topology') {
    if (!isTopology(evidence.data)) return invalidEvidence(evidence, '拓扑数据结构不符合展示契约')
    if (!evidence.data.nodes.length) return <Empty description="暂无拓扑数据" />
    const data = {
      nodes: evidence.data.nodes.map((node) => ({ ...node, position: { ...node.position } })),
      edges: evidence.data.edges.map((edge) => ({ ...edge, data: { ...edge.data } })),
    }
    return <TopologyEvidence data={data} title={evidence.title} />
  }
  return jsonEvidence(evidence)
}

function TopologyEvidence({
  data,
  title,
}: {
  data: { nodes: ApoTopologyNode[]; edges: ApoTopologyEdge[] }
  title: string
}) {
  const [open, setOpen] = useState(false)
  return (
    <>
      <div className={styles.evidenceTopology}>
        <Button
          className={styles.evidenceTopologyExpand}
          size="small"
          icon={<ExpandOutlined />}
          onClick={() => setOpen(true)}
        >
          放大查看
        </Button>
        <Topology data={data} active />
      </div>
      <Modal
        title={title || '拓扑详情'}
        open={open}
        onCancel={() => setOpen(false)}
        footer={null}
        width="92vw"
        styles={{ body: { padding: 0 } }}
        destroyOnClose
      >
        <div className={styles.evidenceTopologyModal}>
          <Topology data={data} active />
        </div>
      </Modal>
    </>
  )
}

function invalidEvidence(evidence: HermesEvidence, message: string) {
  return (
    <>
      <Alert type="warning" showIcon message={message} description="已切换为 JSON 兜底展示。" />
      {jsonEvidence(evidence)}
    </>
  )
}

function jsonEvidence(evidence: HermesEvidence) {
  return (
    <Collapse
      ghost
      size="small"
      items={[{
        key: 'json',
        label: evidence.conversionError ? '查看原始数据与转换错误' : '查看原始数据',
        children: (
          <div className={styles.evidenceJson}>
            {evidence.conversionError && <Alert type="warning" message={evidence.conversionError} />}
            <pre>{safeJson(evidence.data)}</pre>
          </div>
        ),
      }]}
    />
  )
}

function isMetrics(data: unknown): data is ApoMetricItem[] {
  return Array.isArray(data) && data.every((item) =>
    isObject(item) && typeof item.title === 'string' && typeof item.unit === 'string' &&
    Array.isArray(item.timeseries) && item.timeseries.every((series) =>
      isObject(series) && typeof series.legend === 'string' && isObject(series.chart) && isObject(series.chart.chartData),
    ),
  )
}

function isLogs(data: unknown): data is ApoLogRecord[] {
  return Array.isArray(data) && data.every((item) => isObject(item) &&
    (typeof item.timestamp === 'string' || typeof item.timestamp === 'number') && typeof item.message === 'string')
}

function isTopology(data: unknown): data is { nodes: ApoTopologyNode[]; edges: ApoTopologyEdge[] } {
  return isObject(data) && Array.isArray(data.nodes) && Array.isArray(data.edges) &&
    data.nodes.every((node) => isObject(node) && typeof node.id === 'string' && isObject(node.data)) &&
    data.edges.every((edge) => isObject(edge) && typeof edge.id === 'string' &&
      typeof edge.source === 'string' && typeof edge.target === 'string')
}

function isObject(value: unknown): value is Record<string, any> {
  return typeof value === 'object' && value !== null && !Array.isArray(value)
}

function kindIcon(kind: string) {
  if (kind === 'apo_metrics') return <BarChartOutlined />
  if (kind === 'apo_logs') return <FileTextOutlined />
  if (kind === 'apo_topology') return <DeploymentUnitOutlined />
  return <CodeOutlined />
}

function kindLabel(kind: string) {
  if (kind === 'apo_metrics') return '指标'
  if (kind === 'apo_logs') return '日志'
  if (kind === 'apo_topology') return '拓扑'
  return 'JSON'
}

function safeJson(value: unknown) {
  if (value === undefined) return '该 Evidence 未携带可展示的数据。'
  try { return JSON.stringify(value, null, 2) }
  catch { return String(value) }
}
