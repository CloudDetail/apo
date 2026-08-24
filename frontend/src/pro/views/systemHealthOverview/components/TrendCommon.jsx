import React from 'react'
import PropTypes from 'prop-types'
import { Button, Tooltip } from 'antd'
import { CircleAlert, RefreshCw } from 'lucide-react'

import { trendAxisUsesDateOnly, trendRangeOptions } from '../trendViewState.mjs'
import { statusMeta } from './HealthStatus.jsx'

export const historyTrendColors = {
  critical: '#f04438',
  warning: '#f79009',
  stable: '#12b76a',
  score: '#4992ff',
}

export const historyTrendRatioColors = {
  critical: '#b42318',
  warning: '#b54708',
}

export const detailTrendServiceColors = {
  abnormal: '#E24B4A',
  attention: '#EF9F27',
  healthy: '#97C459',
  noData: '#98A2B3',
}

const healthScoreBaselineToneColors = {
  critical: historyTrendColors.critical,
  warning: historyTrendColors.warning,
}

const trendRangeTitleLabels = {
  today: '今天',
  last24h: '近24小时',
  thisWeek: '本周',
  last7d: '过去七天',
  last15d: '过去15天',
  last30d: '过去一个月',
  last6mo: '过去半年',
  last1y: '过去一年',
}

export const getTrendRangeTitleLabel = (rangeKey) =>
  trendRangeTitleLabels[rangeKey] ||
  trendRangeOptions.find((option) => option.key === rangeKey)?.label ||
  '历史'

export const formatTrendAxisLabel = (timestamp, dateOnly = false) => {
  const date = new Date(timestamp)
  if (dateOnly) {
    return `${date.getMonth() + 1}/${date.getDate()}`
  }
  return `${date.getMonth() + 1}/${date.getDate()} ${String(date.getHours()).padStart(2, '0')}:00`
}

export const formatTrendTooltipLabel = (timestamp, granularity) =>
  formatTrendAxisLabel(timestamp, granularity === 'day' || granularity === 'week')

export const trendAxisLabels = (rows, granularity) => {
  const dateOnly = trendAxisUsesDateOnly(rows, granularity)
  return rows.map((row) => formatTrendAxisLabel(row.timestamp, dateOnly))
}

export const isTrendPercentMetric = (metric) => metric?.valueType === 'percent'

export const formatTrendRatio = (value) => {
  const number = Number(value)
  if (!Number.isFinite(number) || number <= 0) return '0%'
  return `${Math.round(number * 100)}%`
}

export const formatTrendNumber = (value, metric) => {
  if (value === undefined || value === null) return '--'
  const number = Number(value)
  if (!Number.isFinite(number)) return '--'
  if (isTrendPercentMetric(metric)) return formatTrendRatio(number)
  return `${Math.round(number)}`
}

export const normalizeTrendDataValue = (value) => {
  if (value === undefined || value === null || value === '') return null
  const number = Number(value)
  return Number.isFinite(number) ? number : null
}

export const trendDataPointValue = (row, key) => {
  if (row?.missingSnapshot) return null
  return normalizeTrendDataValue(row?.[key])
}

export const summaryTrendDataPointValue = (row, metricKey) => {
  if (row?.missingSnapshot) return null
  return normalizeTrendDataValue(row?.values?.[metricKey])
}

export const rowsHaveTrendData = (rows = [], key) =>
  (Array.isArray(rows) ? rows : []).some((row) => trendDataPointValue(row, key) !== null)

export const summaryRowsHaveTrendData = (rows = [], metrics = []) =>
  (Array.isArray(metrics) ? metrics : []).some((metric) =>
    (Array.isArray(rows) ? rows : []).some(
      (row) => summaryTrendDataPointValue(row, metric.key) !== null,
    ),
  )

export const noDataTrendGraphic = () => ({
  type: 'text',
  left: 'center',
  top: 'middle',
  silent: true,
  style: {
    text: '无数据',
    fill: '#98a2b3',
    fontSize: 12,
    fontWeight: 600,
  },
})

export const healthScoreBaselineColor = (item) =>
  healthScoreBaselineToneColors[item?.tone] || statusMeta.gray.color

export const TrendFailureState = ({
  error,
  title = '趋势暂不可用',
  description = '无法读取趋势数据',
  onRetry,
  retrying = false,
}) => (
  <Tooltip title={error}>
    <div className="sho-trend-failure" role="status">
      <div className="sho-trend-failure-inner">
        <span className="sho-trend-failure-icon" aria-hidden="true">
          <CircleAlert size={15} />
        </span>
        <div className="sho-trend-failure-copy">
          <span className="sho-trend-failure-title">{title}</span>
          <span className="sho-trend-failure-description">{description}</span>
        </div>
        {onRetry ? (
          <Button
            aria-label="重试加载趋势"
            className="sho-trend-failure-action"
            icon={retrying ? null : <RefreshCw size={13} />}
            loading={retrying}
            size="small"
            type="text"
            onClick={onRetry}
          >
            重试
          </Button>
        ) : null}
      </div>
    </div>
  </Tooltip>
)

TrendFailureState.propTypes = {
  error: PropTypes.string,
  title: PropTypes.string,
  description: PropTypes.string,
  onRetry: PropTypes.func,
  retrying: PropTypes.bool,
}
