import { Card } from 'antd'
import React, { useEffect, useState } from 'react'
import SingleLinkTopology from '../../../ReportTopology/SvgTopology'
import { useTranslation } from 'react-i18next'

const i18nKey = 'causeStep'
export default function ChainStep(props: any) {
  const { t } = useTranslation('pro/causeReport')

  const { data, component } = props
  const [topologyData, setTopologyData] = useState<any>(null)
  useEffect(() => {
    const nodes: any = []
    let edges: any[] = []
    data.metrics.chain_data.nodes.forEach((node: any, index: number) => {
      nodes.push({
        ...node,
        chartId: node.id + '-' + index,
      })
      if (index > 0) {
        edges.push({
          source: data.metrics.chain_data.nodes[index - 1].id + '-' + (index - 1),
          target: node.id + '-' + index,
          nextCallTime: data.metrics.chain_data.nodes[index - 1].nextCallTime,
        })
      }
    })
    setTopologyData({ nodes, edges })
  }, [data])
  return (
    <Card title={data.report_name}>
      {topologyData && (
        <div className="chart_warp">
          <SingleLinkTopology data={topologyData} category={'chain'} chartId={data.flow_name} />
        </div>
      )}
      {data?.conclusion?.length > 0 && (
        <div>
          <span>{t(`${i18nKey}.analysisConclusion`)}：</span>
          {/* <a className='conclusion_tag' href={`/d/c902c778-2a67-463c-8f08-ae495df7fd28/2?var-pql=${testPQL}&viewPanel=1`}>{data.conclusion}</a> */}
          <span className="conclusion_tag">{data.conclusion}</span>
        </div>
      )}
    </Card>
  )
}
