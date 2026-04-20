import React, { useMemo } from 'react'
import { useTranslation } from 'react-i18next'
import type { FindingsNodePayload } from '../../../../api/conversation'
import FindingsTopology from './FindingsTopology'
import { buildFindingsTopologyData } from './findingsContract'

interface FindingsTreeProps {
  nodes: FindingsNodePayload[]
  onNodeClick?: (nodeData: any) => void
  findingsId?: string
}

const FindingsTree = ({ nodes, onNodeClick, findingsId = 'findings-tree' }: FindingsTreeProps) => {
  const { t } = useTranslation('pro/agui')

  const data = useMemo(() => {
    const topologyData = buildFindingsTopologyData(nodes)
    return {
      nodes: topologyData.nodes.map((node) => ({
        ...node,
        data: {
          ...node.data,
          findingsId,
          onClick: onNodeClick,
        },
      })),
      edges: topologyData.edges,
    }
  }, [findingsId, nodes, onNodeClick])

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
