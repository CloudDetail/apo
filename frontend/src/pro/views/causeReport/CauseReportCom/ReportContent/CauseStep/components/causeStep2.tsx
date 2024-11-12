import React, { useEffect, useState } from 'react'
import BasicTable from 'src/core/components/Table/basicTable'
import { formatUnit } from 'src/core/utils/time'
import { Button, Card } from 'antd'
import { useCauseReportContext } from 'src/pro/contexts/CauseReportContext'

// const testPQL = 'sum by (device) (rate(node_disk_io_time_seconds_total{instance=\"10.10.103.219:9100\"}[5m]))*100';
const metric = 'kindling_profiling_duration_nanoseconds_total'
interface EventProps {
  name: string
  baseline: number
  basePercent?: number
  current: number
  currentPercent?: number
  warn: boolean
  hide?: boolean
}
interface Props {
  data: any
}
function CauseStep2({ data }: Props) {
  const [eventList, setEventList] = useState<EventProps[]>([])
  const { traceId, spanId } = useCauseReportContext()
  useEffect(() => {
    const list: EventProps[] = []
    const events = data.metrics[metric] || []
    events.forEach((item: any) => {
      list.push({
        name: item.labels.cpuType,
        baseline: item.baseline.value,
        current: item.current.value,
        warn: item.is_warn,
      })
    })
    setEventList(list)
  }, [data])

  const columns = [
    {
      title: '',
      accessor: 'name',
      Cell: ({ value }: any) => {
        return value
        // const href = TO_DASHBOARD[value];
        // return (
        //     <>
        //         {
        //             href ?
        //                 <Tooltip content="点击跳转至相关Dashboard">
        //                     <a href={href}>{value}</a>

        //                 </Tooltip>
        //                 : value

        //         }
        //     </>
        // );
      },
    },
    {
      title: '历史基线',
      accessor: 'baseline',
      Cell: ({ value }: any) => formatUnit(value, 'ns'),
    },
    {
      title: '本次耗时',
      accessor: 'current',
      Cell: (cell: any) => {
        const { value } = cell
        const { original } = cell.cell.row
        // console.log('props', cell, original);
        return (
          <span style={{ color: original.warn ? '#f3485c' : '' }}>{formatUnit(value, 'ns')}</span>
        )
      },
    },
    {
      title: '操作',
      accessor: 'action',
      Cell: (cell: any) => {
        const { value } = cell
        const { original } = cell.cell.row
        return original.warn ? (
          <Button type="text">当前根因</Button>
        ) : (
          <Button
            type="link"
            target="_blank"
            href={`/#/cause/report/${encodeURIComponent(traceId)}/${encodeURIComponent(spanId)}?mutatedType=${encodeURIComponent(original.name)}`}
          >
            确认为根因
          </Button>
        )
      },
    },
  ]
  return (
    <Card title="接口执行耗时分析">
      <div className="table_step_description">
        {/* <Table data={eventList} width={800} height={300} /> */}
        <BasicTable columns={columns} data={eventList} />
        {data?.conclusion?.length > 0 && (
          <div>
            <span>分析结论：</span>
            {/* <a className='conclusion_tag' href={`/d/c902c778-2a67-463c-8f08-ae495df7fd28/2?var-pql=${testPQL}&viewPanel=1`}>{data.conclusion}</a> */}
            <span className="conclusion_tag">{data.conclusion}</span>
          </div>
        )}
      </div>
    </Card>
  )
}

export default CauseStep2
