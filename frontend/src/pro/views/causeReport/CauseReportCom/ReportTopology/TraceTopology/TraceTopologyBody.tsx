import { Card } from 'antd'
import { useCauseReportContext } from 'src/pro/contexts/CauseReportContext'
import TraceTopologyCom from './TraceTopologyCom'
import ErrorNodeInfo from '../../ReportContent/CauseStep/components/ErrorNodeInfo'
import { useTranslation } from 'react-i18next'

const TraceTopologyBody = () => {
  const { t } = useTranslation('pro/causeReport')

  const { topologyData, topologyTreeData, category, traceId, node } = useCauseReportContext()

  return (
    topologyData?.nodes &&
    topologyData?.nodes.length > 0 && (
      <div>
        <Card title={category === 'error' ? t('topology.fullLinkInvocation') : ''}>
          {topologyTreeData && <TraceTopologyCom data={topologyTreeData} category={category} />}
        </Card>
        {category === 'error' && <ErrorNodeInfo traceId={traceId} node={node} />}
      </div>
    )
  )
}
export default TraceTopologyBody
