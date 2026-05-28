import PropTypes from 'prop-types'
import Logs from './Logs'
import REDMetrics from './RED'
import MultiLineChart from 'src/core/components/Chart/MultiLineChart'
import { formatMGT } from 'src/core/utils/time'
import styles from './index.module.scss'

interface EvidenceItem {
  name: string
  type: 'log' | 'red' | 'metrics' | 'endpoints' | 'trace' | 'traces'
  data: any
  unit?: string
  description?: string
}

interface SupportingEvidenceProps {
  evidence: EvidenceItem[]
  startTime: number
  endTime: number
  getEvidenceItemId?: (index: number) => string
  activeEvidenceIndex?: number | null
  activeEvidenceClassName?: string
}

const YFormatter = (value: number, unit) => {
  switch (unit) {
    case 'percent':
      if (value > 0 && value < 0.01) return '< 0.01%'
      return `${Number.parseFloat(value.toFixed(2))}%`
    case 'bytes': {
      // 先将 bytes 转换为 MB，再调用 formatMGT（formatMGT 期望输入是 MB 单位）
      const valueInMB = value / (1024 * 1024)
      if (value > 0 && value < 1024 * 1024 * 0.01) return '< 0.01M'
      return `${formatMGT(Number.parseFloat(valueInMB.toFixed(2)))}`
    }
    default:
      if (value > 0 && value < 0.01) return `< 0.01 ${unit}`
      return `${Number.parseFloat(value.toFixed(2))} ${unit}`
  }
}

const convertMetricsData = (data) => {
  return (Array.isArray(data) ? data : [])
    .map((item) => ({
      data: Object.entries(item?.chart?.chartData || item?.chartData || {})
        .map(([key, value]) => [Number(key) / 1000, value])
        .sort((a, b) => a[0] - b[0]),
      legend: item?.legend || item?.name || item?.type || 'series',
    }))
    .filter((item) => item.data.length > 0)
}

const evidenceTypeLabel: Record<EvidenceItem['type'], string> = {
  log: '日志',
  red: 'RED',
  metrics: '指标',
  endpoints: '拓扑',
  trace: 'Trace',
  traces: 'Trace',
}

const formatTraceData = (data) => {
  if (typeof data === 'string') return data
  if (data === null || data === undefined) return '暂无 Trace 样本'

  try {
    return JSON.stringify(data, null, 2)
  } catch {
    return String(data)
  }
}

const TraceEvidence = ({ data }) => {
  return <pre className={styles.traceEvidence}>{formatTraceData(data)}</pre>
}

TraceEvidence.propTypes = {
  data: PropTypes.any,
}

const SupportingEvidence = ({
  evidence,
  startTime,
  endTime,
  getEvidenceItemId,
  activeEvidenceIndex,
  activeEvidenceClassName,
}: SupportingEvidenceProps) => {
  const visibleEvidence = evidence
    .map((item, evidenceIndex) => ({ item, evidenceIndex }))
    .filter(({ item }) => item.type !== 'endpoints')
  const getChildren = (type, data, unit, description) => {
    return (
      <>
        {type === 'log' ? (
          <Logs logs={data} />
        ) : type === 'red' ? (
          <REDMetrics data={data} startTime={startTime} endTime={endTime} />
        ) : type === 'metrics' ? (
          <div className={styles.chartFrame}>
            {/* <LineChart data={data} unit={unit} /> */}
            <MultiLineChart
              chartData={convertMetricsData(data)}
              unit={unit}
              allowTimeBrush={false}
              YFormatter={(value) => YFormatter(value, unit)}
            />
          </div>
        ) : type === 'trace' || type === 'traces' ? (
          <TraceEvidence data={data} />
        ) : type === 'endpoints' ? null : (
          <TraceEvidence data={data} />
        )}
        {description && <p className={styles.evidenceDescription}>{description}</p>}
      </>
    )
  }

  return (
    <div className={styles.evidenceStack}>
      {visibleEvidence.map(({ item, evidenceIndex }) => (
        <article
          className={[
            styles.evidenceItem,
            activeEvidenceIndex === evidenceIndex ? activeEvidenceClassName : '',
          ]
            .filter(Boolean)
            .join(' ')}
          id={getEvidenceItemId?.(evidenceIndex)}
          key={`${item.type}-${item.name}-${evidenceIndex}`}
        >
          <header>
            <strong>{item.name}</strong>
            <span>{evidenceTypeLabel[item.type] || '证据'}</span>
          </header>
          <div className={styles.evidenceBody}>
            {getChildren(item.type, item.data, item.unit, item.description)}
          </div>
        </article>
      ))}
    </div>
  )
}

SupportingEvidence.propTypes = {
  evidence: PropTypes.arrayOf(
    PropTypes.shape({
      name: PropTypes.string.isRequired,
      type: PropTypes.oneOf(['log', 'red', 'metrics', 'endpoints', 'trace', 'traces']).isRequired,
      data: PropTypes.any.isRequired,
      unit: PropTypes.string,
      description: PropTypes.string,
    }),
  ).isRequired,
  startTime: PropTypes.number.isRequired,
  endTime: PropTypes.number.isRequired,
  getEvidenceItemId: PropTypes.func,
  activeEvidenceIndex: PropTypes.number,
  activeEvidenceClassName: PropTypes.string,
}

export default SupportingEvidence
