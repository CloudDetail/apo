import { Card, Col, Row, Space } from 'antd'
import React from 'react'
import CopyButton from 'src/core/components/CopyButton'
import { useCauseReportContext } from 'src/pro/contexts/CauseReportContext'
import { convertTime, formatUnit } from 'src/core/utils/time'
import TimeProgress from './TimeProgress'
import NodeInfo from './NodeInfo'
import Conclusion from './Conclusion'

const ReportHeader = () => {
  const { entryTrace, category, uploadTraceId, traceId, reportData, causeNode } =
    useCauseReportContext()
  return (
    <Card title="故障概览">
      <Card classNames={{ body: 'px-3 py-2' }} className="mb-2">
        <Space>
          <span className="font-bold text-base mb-1">TraceID：</span>
          {uploadTraceId ? uploadTraceId : traceId}
          <CopyButton value={uploadTraceId ? uploadTraceId : traceId} />
        </Space>
      </Card>
      <NodeInfo title="当前节点" {...causeNode} />
      <NodeInfo title="受影响的业务入口" {...entryTrace} />
      <Card classNames={{ body: 'px-3 py-2' }}>
        <div className="flex">
          <span className="font-bold text-base mb-1 flex-shrink-0 flex-grow-0">故障结论：</span>

          <div className="flex-1 overflow-hidden">
            <Conclusion />
          </div>
        </div>
      </Card>
    </Card>
  )
}

export default ReportHeader
