import { Card, Col, Row, Space } from 'antd'
import React from 'react'
import CopyButton from 'src/core/components/CopyButton'
import { useCauseReportContext } from 'src/pro/contexts/CauseReportContext'
import { convertTime, formatUnit } from 'src/core/utils/time'
import TimeProgress from './TimeProgress'
import NodeInfo from './NodeInfo'
import Conclusion from './Conclusion'
import { useTranslation } from 'react-i18next'

const ReportHeader = () => {
  const { t } = useTranslation('pro/causeReport')
  const { entryTrace, category, uploadTraceId, traceId, reportData, causeNode } =
    useCauseReportContext()
  return (
    <Card title={t('faultOverview')}>
      <Card classNames={{ body: 'px-3 py-2' }} className="mb-2">
        <Space>
          <span className="font-bold text-base mb-1">{t('traceID')}：</span>
          {uploadTraceId ? uploadTraceId : traceId}
          <CopyButton value={uploadTraceId ? uploadTraceId : traceId} />
        </Space>
      </Card>
      <NodeInfo title={t('currentNode')} {...causeNode} />
      <NodeInfo title={t('affectedBusinessEntry')} {...entryTrace} />
      <Card classNames={{ body: 'px-3 py-2' }}>
        <div className="flex">
          <span className="font-bold text-base mb-1 flex-shrink-0 flex-grow-0">
            {t('faultConclusion')}：
          </span>
          <div className="flex-1 overflow-hidden">
            <Conclusion />
          </div>
        </div>
      </Card>
    </Card>
  )
}

export default ReportHeader
