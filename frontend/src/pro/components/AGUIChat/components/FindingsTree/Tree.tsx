import React, { useMemo } from 'react'
import { useTranslation } from 'react-i18next'
import FindingsTopology from './FindingsTopology'

type NodeStatus = 'normal' | 'warning' | 'abnormal'

interface FindingsNode {
  id: string
  details: string
  title?: string
  subtitle?: string
  statusColor: NodeStatus
  summaryConclusion: string
  parentId: string | null
  metrics: {
    abnormalMetrics: any[]
    normalMetrics: any[]
  }
  traces: any[]
  logs: any[]
  table?: any
}

interface FindingsTreeProps {
  nodes: FindingsNode[]
  onNodeClick?: (nodeData: any) => void
  findingsId?: string
}

const FindingsTree = ({ nodes, onNodeClick, findingsId = 'findings-tree' }: FindingsTreeProps) => {
  const { t } = useTranslation('pro/agui')

  // 将 FindingsNode[] 转换为 FindingsTopology 所需的 {nodes, edges} 格式
  console.log(nodes)
  const data = useMemo(() => {
    if (!nodes || nodes.length === 0) {
      return { nodes: [], edges: [] }
    }

    // 转换节点格式
    const formattedNodes = nodes.map((node) => ({
      id: node.id,
      type: 'findingsNode',
      data: {
        title: node.title,
        subtitle: node.subtitle,
        content: node.details,
        status: node.statusColor,
        summaryConclusion: node.summaryConclusion,
        logs: node.logs,
        metrics: node.metrics,
        traces: node.traces,
        table: node.table || undefined,
      },
    }))

    // 根据 parentId 构建边
    const edges = nodes
      .filter((node) => node.parentId !== null && node.parentId !== undefined)
      .map((node) => ({
        id: `${node.parentId}-${node.id}`,
        source: node.parentId as string,
        target: node.id,
      }))

    return {
      nodes: formattedNodes,
      edges,
    }
  }, [nodes])

  if (!nodes || nodes.length === 0) {
    return <div style={{ padding: 20, textAlign: 'center' }}>{t('findingsTree.noData')}</div>
  }

  return (
    <div style={{ width: '100%', height: '100%' }}>
      <FindingsTopology data={data} onNodeClick={onNodeClick} findingsId={findingsId} />
    </div>
  )
}

export default FindingsTree
