import React, { useState, useEffect } from 'react'
import _ from 'lodash'
import { useCauseReportContext } from 'src/pro/contexts/CauseReportContext'
import ReportHeader from './ReportHeader'
import ReportTopology from './ReportTopology'
import ReportContent from './ReportContent'
import TraceTopologyBody from './ReportTopology/TraceTopology/TraceTopologyBody'
import LoadingSpinner from 'src/core/components/Spinner'
import { Collapse, Empty } from 'antd'
export interface CustomVizOptions {
  traceId?: string
  entryTrace?: any
  reportData?: any
  topologyData?: any
  topologyTreeData?: any
  loading?: boolean
  isTimeOut?: boolean
  category: string
}
function CauseReportCom() {
  const { traceId, category, loading, dataMissing } = useCauseReportContext()

  const [showEmptyData, setShowEmptyData] = useState<boolean>(false)

  return (
    <>
      <LoadingSpinner loading={loading} />
      {
        dataMissing ? (<div className='mt-60'>
          <Empty description="服务入口数据缺失" />
        </div>) : !traceId ? (<></>) : (
          // <Empty>当前无导入数据，请点击按钮导入</Empty>
          <>
            <div className="mb-2 ">
              <ReportHeader />
            </div>
            {category === 'slow' && <ReportTopology />}
            <Collapse
              className="mb-2 bg-[#141414]"
              items={[
                {
                  key: 'report',
                  label: '故障分析过程',
                  classNames: { header: 'text-base h-[56px]' },
                  styles: { header: { alignItems: 'center' } },
                  children: (
                    <>
                      <div className="report_warp">
                        <ReportContent />
                      </div>
                      {category === 'error' && <TraceTopologyBody />}
                    </>
                  ),
                },
              ]}
            ></Collapse>
          </>
        )}
    </>
  )
}

export default CauseReportCom
