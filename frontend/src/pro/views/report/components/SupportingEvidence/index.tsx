import { Collapse } from 'antd'
import { SlArrowDown } from 'react-icons/sl'
import PropTypes from 'prop-types'
import Logs from './Logs'
import REDMetrics from './RED'
import AffectedEndpoints from '../AffectedEndpoints'
import MultiLineChart from 'src/core/components/Chart/MultiLineChart'
import { formatMGT } from 'src/core/utils/time'

interface EvidenceItem {
  name: string
  type: 'log' | 'red' | 'metrics' | 'endpoints'
  data: any
  unit?: string
}

interface SupportingEvidenceProps {
  evidence: EvidenceItem[]
  startTime: number
  endTime: number
}

const YFormatter = (value: number, unit) => {
  switch (unit) {
    case 'percent':
      if (value > 0 && value < 0.01) return '< 0.01%'
      return `${Number.parseFloat(value.toFixed(2))}%`
    case 'bytes':
      if (value > 0 && value < 1024 * 0.01) return '< 0.01M'
      return `${formatMGT(Number.parseFloat(value.toFixed(2)))}`
    default:
      if (value > 0 && value < 0.01) return `< 0.01 ${unit}`
      return `${Number.parseFloat(value.toFixed(2))} ${unit}`
  }
}

const convertMetricsData = (data) => {
  return data.map((item) => ({
    data: Object.entries(item.chart.chartData)
      ?.map(([key, value]) => [Number(key) / 1000, value])
      .sort((a, b) => a[0] - b[0]),
    legend: item.legend,
  }))
}

const SupportingEvidence = ({ evidence, startTime, endTime }: SupportingEvidenceProps) => {
  const getChildren = (type, data, unit, description) => {
    return (
      <>
        {type === 'log' ? (
          <Logs logs={data} />
        ) : type === 'red' ? (
          <REDMetrics data={data} startTime={startTime} endTime={endTime} />
        ) : type === 'metrics' ? (
          <div className="h-[300px]">
            {/* <LineChart data={data} unit={unit} /> */}
            <MultiLineChart
              chartData={convertMetricsData(data)}
              unit={unit}
              startTime={startTime}
              endTime={endTime}
              allowTimeBrush={false}
              YFormatter={(value) => YFormatter(value, unit)}
            />
          </div>
        ) : (
          <AffectedEndpoints topology={data} />
        )}
        <div className="mt-3 border-l-4 border-blue-400 pl-4 py-2 bg-[var(--ant-color-info-bg)] rounded-r-md">
          <div className="text-[var(--ant-color-text)] text-sm">{description}</div>
        </div>
      </>
    )
  }
  return (
    <>
      {evidence.map((item, index) => (
        <div className="mt-3 border border-gray-200 dark:border-gray-700 rounded-lg overflow-hidden">
          <Collapse
            items={[
              {
                key: index,
                label: item.name,
                children: getChildren(item.type, item.data, item.unit, item.description),
              },
            ]}
            defaultActiveKey={index}
            //   className={styles.collapse}
            // bordered={false}
            expandIconPosition={'end'}
            expandIcon={({ isActive }) => {
              return isActive ? (
                <SlArrowDown className={'rotate-180'} size={14} />
              ) : (
                <SlArrowDown className={'rotate-0'} size={14} />
              )
            }}
          />
        </div>
      ))}
    </>
  )
}

SupportingEvidence.propTypes = {
  evidence: PropTypes.arrayOf(
    PropTypes.shape({
      name: PropTypes.string.isRequired,
      type: PropTypes.oneOf(['log', 'red', 'metrics', 'endpoints']).isRequired,
      data: PropTypes.any.isRequired,
      unit: PropTypes.string,
    }),
  ).isRequired,
  startTime: PropTypes.number.isRequired,
  endTime: PropTypes.number.isRequired,
}

export default SupportingEvidence
