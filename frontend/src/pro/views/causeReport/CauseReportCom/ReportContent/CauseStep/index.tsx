import React from 'react'
import CauseStep1 from './components/causeStep1'
import TableStep from './components/tableStep'
import ChartStep from './components/chartStep'
import ErrorNodeInfo from './components/ErrorNodeInfo'
import ChainStep from './components/chainStep'
import _ from 'lodash'
import CauseStep2 from './components/causeStep2'
import { useTranslation } from 'react-i18next'
import SpanDetailStep from './components/SpanDetailStep'
interface CauseProps {
  data: any
  occurrenceAt?: number
}

function CauseStep({ data, occurrenceAt }: CauseProps) {
  const { t } = useTranslation('pro/causeReport')

  if (data.type === 'trace_entry') {
    return <CauseStep1 data={data} />
  } else if (data.type === 'profiling_duration') {
    return <CauseStep2 data={data} />
  } else if (data.type === 'spanDetail') {
    return <SpanDetailStep data={data} />
  } else if (
    data.type === 'metrics_baseline' ||
    data.type === 'tableMetrics' ||
    data.type === 'table'
  ) {
    return <TableStep data={data} />
  } else if (data.type === 'metrics_timeseries') {
    return <ChartStep data={data} occurrenceAt={occurrenceAt} />
  } else if (data.type === 'display') {
    return <ChartStep data={data} />
  } else if (data.type === 'exception') {
    // return <ErrorNodeInfo node={data} />
  } else if (data.type === 'callChain') {
    return <ChainStep data={data} />
  } else {
    return <div>{t('noMatchingType')}</div>
  }
}

export default CauseStep
