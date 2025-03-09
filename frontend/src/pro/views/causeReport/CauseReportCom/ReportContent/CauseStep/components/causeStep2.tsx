import React, { useEffect, useState } from 'react'
import BasicTable from 'src/core/components/Table/basicTable'
import { formatUnit } from 'src/core/utils/time'
import { Button, Card } from 'antd'
import { useCauseReportContext } from 'src/pro/contexts/CauseReportContext'
import { useTranslation } from 'react-i18next'

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
const i18nKey = 'causeStep'

function CauseStep2({ data }: Props) {
  const { t } = useTranslation('pro/causeReport')

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
      Cell: ({ value }: any) => value,
    },
    {
      title: t(`${i18nKey}.historicalBaseline`),
      accessor: 'baseline',
      Cell: ({ value }: any) => formatUnit(value, 'ns'),
    },
    {
      title: t(`${i18nKey}.currentDuration`),
      accessor: 'current',
      Cell: (cell: any) => {
        const { value } = cell
        const { original } = cell.cell.row
        return (
          <span style={{ color: original.warn ? '#f3485c' : '' }}>{formatUnit(value, 'ns')}</span>
        )
      },
    },
    {
      title: t(`${i18nKey}.action`),
      accessor: 'action',
      Cell: (cell: any) => {
        const { original } = cell.cell.row
        return original.warn ? (
          <Button type="text">{t(`${i18nKey}.currentRootCause`)}</Button>
        ) : (
          <Button
            type="link"
            target="_blank"
            href={`/#/cause/report/${encodeURIComponent(traceId)}/${encodeURIComponent(spanId)}?mutatedType=${encodeURIComponent(original.name)}`}
          >
            {t(`${i18nKey}.confirmRootCause`)}
          </Button>
        )
      },
    },
  ]

  return (
    <Card title={t(`${i18nKey}.interfaceExecutionTimeAnalysis`)}>
      <div className="table_step_description">
        {/* <Table data={eventList} width={800} height={300} /> */}
        <BasicTable columns={columns} data={eventList} />
        {data?.conclusion?.length > 0 && (
          <div>
            <span>{t(`${i18nKey}.analysisConclusion`)}：</span>
            {/* <a className='conclusion_tag' href={`/d/c902c778-2a67-463c-8f08-ae495df7fd28/2?var-pql=${testPQL}&viewPanel=1`}>{data.conclusion}</a> */}
            <span className="conclusion_tag">{data.conclusion}</span>
          </div>
        )}
      </div>
    </Card>
  )
}

export default CauseStep2
