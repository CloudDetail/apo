import { Card, Collapse } from 'antd'
import React, { Children, useEffect, useState } from 'react'
import TimeProgress from '../../../ReportHeader/TimeProgress'
import { formatUnit } from 'src/core/utils/time'
import HighlightCode from 'src/core/components/HighlightCode/HighlightCode'

// const testPQL = 'sum by (device) (rate(node_disk_io_time_seconds_total{instance="192.168.1.18:30003"}[5m]))*100';
// const testPQL = 'sum(rate(node_disk_io_time_seconds_total{}[5m]))';
const metric = 'kindling_span_trace_duration_nanoseconds_total'

interface Props {
  data: any
}
function CauseStep1({ data }: Props) {
  const p90 = data.metrics[metric][0].baseline.value || 0
  const error = data.metrics[metric][0].current.value || 0
  const [openIdList, setOpenIdList] = useState<number[]>([])
  const [logCount, setLogCount] = useState<number>(0)
  useEffect(() => {
    let count = 0
    let openId: any = null
    data?.logs.forEach((log: any) => {
      if (log?.id !== undefined && log?.name && log?.values.length > 0) {
        count++
        if (openId === null) {
          openId = log.id
        }
      }
    })
    setLogCount(count)
    setOpenIdList([openId])
    // setOpenId(openId);
  }, [data])

  const getCollapseItems = () => {
    return data.logs.map((item: any) => ({
      key: item.id,
      label: item.name,
      children: <HighlightCode rows={item.values} title={item.name} />,
    }))
  }
  return (
    <Card title="故障节点分析">
      <div className="w-full-with-children">
        <div className="grid">
          <div>服务名: {data.text.service_name}</div>
          <div>主机名: {data.text.pod}</div>
          <div>发生时间: {formatUnit(data.timestamp / 1000000, 'date')}</div>
          <div style={{ gridColumn: 'span 3' }}>请求URL: {data.text.content_key}</div>
        </div>
        {/* {[
          `服务名: ${data.text.service_name}`,
          `主机名: ${data.text.pod}`,
          `请求URL: ${data.text.content_key}`,
          `发生时间: ${formatUnit(data.timestamp / 1000000, 'date')}`,
        ]} */}
      </div>
      <div>
        <TimeProgress p90={p90} error={error} />

        {logCount > 0 ? (
          <React.Fragment>
            <label style={{ marginBottom: '5px' }}>相关日志</label>
            <Collapse
              items={getCollapseItems()}
              defaultActiveKey={[data.logs?.length > 0 ? data.logs[0].id : '']}
            ></Collapse>
          </React.Fragment>
        ) : null}
        {data?.conclusion?.length > 0 && (
          <div>
            <span>分析结论：</span>
            {/* <a className='conclusion_tag' href={`/d/9CWBz0bik/node-exporter?var-varPql=${testPQL}&viewPanel=181`}>{data.conclusion}</a> */}
            <span className="conclusion_tag">{data.conclusion}</span>
          </div>
        )}
      </div>
    </Card>
  )
}

export default CauseStep1
