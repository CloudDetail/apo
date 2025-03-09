import { Card, Collapse } from 'antd'
import React, { Children, useEffect, useState } from 'react'
import TimeProgress from '../../../ReportHeader/TimeProgress'
import { formatUnit } from 'src/core/utils/time'
import HighlightCode from 'src/core/components/HighlightCode/HighlightCode'
import { useTranslation } from 'react-i18next'

// const testPQL = 'sum by (device) (rate(node_disk_io_time_seconds_total{instance="192.168.1.18:30003"}[5m]))*100';
// const testPQL = 'sum(rate(node_disk_io_time_seconds_total{}[5m]))';
const metric = 'kindling_span_trace_duration_nanoseconds_total'

interface Props {
  data: any
}
const i18nKey = 'causeStep'
function CauseStep1({ data }: Props) {
  const { t } = useTranslation('pro/causeReport')
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
    <Card title={t(`${i18nKey}.faultNodeAnalysis`)}>
      <div className="w-full-with-children">
        <div className="grid">
          <div>
            {t(`${i18nKey}.serviceName`)}: {data.text.service_name}
          </div>
          <div>
            {t(`${i18nKey}.hostName`)}: {data.text.pod}
          </div>
          <div>
            {t(`${i18nKey}.occurrenceTime`)}: {formatUnit(data.timestamp / 1000000, 'date')}
          </div>
          <div style={{ gridColumn: 'span 3' }}>
            {t(`${i18nKey}.requestURL`)}: {data.text.content_key}
          </div>
        </div>
      </div>
      <div>
        <TimeProgress p90={p90} error={error} />

        {logCount > 0 ? (
          <React.Fragment>
            <label style={{ marginBottom: '5px' }}>{t(`${i18nKey}.relatedLogs`)}</label>
            <Collapse
              items={getCollapseItems()}
              defaultActiveKey={[data.logs?.length > 0 ? data.logs[0].id : '']}
            ></Collapse>
          </React.Fragment>
        ) : null}

        {data?.conclusion?.length > 0 && (
          <div>
            <span>{t(`${i18nKey}.analysisConclusion`)}：</span>
            <span className="conclusion_tag">{data.conclusion}</span>
          </div>
        )}
      </div>
    </Card>
  )
}

export default CauseStep1
