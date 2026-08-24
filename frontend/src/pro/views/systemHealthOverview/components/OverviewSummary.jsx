import React, { useMemo } from 'react'
import PropTypes from 'prop-types'
import { Spin, Statistic, Tooltip } from 'antd'
import ReactECharts from 'echarts-for-react'

import { safePercent } from '../metricVisuals.mjs'
import { buildTrendChartRows } from '../trendViewState.mjs'
import { statusMeta, statusOrder } from './HealthStatus.jsx'
import {
  formatTrendNumber,
  formatTrendRatio,
  formatTrendTooltipLabel,
  historyTrendColors,
  historyTrendRatioColors,
  isTrendPercentMetric,
  noDataTrendGraphic,
  summaryRowsHaveTrendData,
  summaryTrendDataPointValue,
  trendAxisLabels,
  TrendFailureState,
} from './TrendCommon.jsx'

const StatusDistributionBar = ({ distribution }) => {
  const total = Number.isFinite(Number(distribution.total)) && Number(distribution.total) > 0
    ? Number(distribution.total)
    : 0

  return (
    <div className="sho-status-distribution" aria-label="系统状态分布">
      {statusOrder.map((status) => {
        const count = Number.isFinite(Number(distribution[status])) && Number(distribution[status]) > 0
          ? Number(distribution[status])
          : 0
        if (!count) return null
        const meta = statusMeta[status]
        return (
          <Tooltip title={`${meta.label}：${count} 个系统`} key={status}>
            <span
              className={`sho-status-distribution-segment sho-status-distribution-${status}`}
              style={{ flexGrow: count, '--sho-segment-percent': `${safePercent(count, total)}%` }}
            />
          </Tooltip>
        )
      })}
      {!total && <span className="sho-status-distribution-empty" />}
    </div>
  )
}

StatusDistributionBar.propTypes = {
  distribution: PropTypes.shape({
    total: PropTypes.number,
    red: PropTypes.number,
    yellow: PropTypes.number,
    green: PropTypes.number,
    gray: PropTypes.number,
  }).isRequired,
}

const SummaryMetricCard = ({
  label,
  help,
  value,
  icon,
  numerator,
  denominator,
  highlightNumerator = false,
  percent,
  tone = 'neutral',
  visual = 'meter',
  children,
}) => {
  const hasBody = Boolean(children)
  const shouldShowMeter = visual === 'meter' && !hasBody
  const numeratorClassName = highlightNumerator
    ? `sho-summary-value-highlight sho-summary-value-highlight-${tone}`
    : undefined
  const renderedValue =
    visual === 'ratio' ? (
      <span className="sho-summary-value" aria-label={`${label} ${value}`}>
        <span className={numeratorClassName}>{numerator}</span>
        <span> / {denominator}</span>
      </span>
    ) : (
      value
    )
  const statisticFormatter = visual === 'ratio' ? () => renderedValue : undefined
  const titleContent = (
    <span
      className={help ? 'sho-summary-title-trigger' : undefined}
      tabIndex={help ? 0 : undefined}
    >
      {icon}
      <span>{label}</span>
    </span>
  )

  return (
    <div
      className={`sho-summary-card sho-summary-card-${tone} sho-summary-card-${visual}${
        hasBody ? ' sho-summary-card-with-body' : ''
      }`}
    >
      <div className="sho-summary-card-head">
        {help ? <Tooltip title={help}>{titleContent}</Tooltip> : titleContent}
        <Statistic className="sho-summary-statistic" formatter={statisticFormatter} value={value} />
      </div>
      {hasBody ? <div className="sho-summary-card-body">{children}</div> : null}
      {shouldShowMeter ? (
        <div className="sho-summary-meter" aria-hidden="true">
          <span style={{ '--sho-fill': `${percent}%` }} />
        </div>
      ) : null}
    </div>
  )
}

SummaryMetricCard.propTypes = {
  label: PropTypes.string.isRequired,
  help: PropTypes.node,
  value: PropTypes.node.isRequired,
  icon: PropTypes.node,
  numerator: PropTypes.number,
  denominator: PropTypes.number,
  highlightNumerator: PropTypes.bool,
  percent: PropTypes.number,
  tone: PropTypes.oneOf(['critical', 'warning', 'stable', 'neutral']),
  visual: PropTypes.oneOf(['meter', 'number', 'ratio']),
  children: PropTypes.node,
}

const formatSummaryTrendTooltipMetricLabel = (metric) =>
  `${metric?.seriesLabel || metric?.label || ''}峰值`

const summaryTrendSeriesColor = (metric) => {
  if (isTrendPercentMetric(metric)) {
    return historyTrendRatioColors[metric.tone] || historyTrendColors[metric.tone] || '#155eef'
  }
  return historyTrendColors[metric.tone] || '#4992ff'
}

const buildSummaryCombinedTrendRows = (points, metrics = []) => {
  const rowsByMetric = {}
  const timestamps = new Set()
  metrics.forEach((metric) => {
    const metricRows = buildTrendChartRows(points, metric.key)
    rowsByMetric[metric.key] = new Map(metricRows.map((row) => [row.timestamp, row]))
    metricRows.forEach((row) => timestamps.add(row.timestamp))
  })

  return Array.from(timestamps)
    .sort((left, right) => left - right)
    .map((timestamp) => {
      const values = {}
      const stats = {}
      let missingSnapshot = false
      metrics.forEach((metric) => {
        const row = rowsByMetric[metric.key]?.get(timestamp)
        values[metric.key] = row?.value ?? null
        stats[metric.key] = row?.stats || {}
        if (row?.missingSnapshot) {
          missingSnapshot = true
        }
      })
      return { timestamp, values, stats, ...(missingSnapshot ? { missingSnapshot: true } : {}) }
    })
}

const buildSummaryTrendOption = (rows, group, granularity) => {
  const metrics = Array.isArray(group?.metrics) ? group.metrics : []
  return {
    animation: false,
    grid: { left: 30, right: 34, top: 14, bottom: 24 },
    tooltip: {
      trigger: 'axis',
      confine: true,
      formatter: (params = []) => {
        const index = params[0]?.dataIndex ?? 0
        const row = rows[index]
        if (!row) return ''
        if (row.missingSnapshot) {
          return [
            `<strong>${formatTrendTooltipLabel(row.timestamp, granularity)}</strong>`,
            '无数据',
          ].join('<br/>')
        }
        return [
          `<strong>${formatTrendTooltipLabel(row.timestamp, granularity)}</strong>`,
          ...metrics.map((metric) => {
            const value = row.values?.[metric.key]
            return `${formatSummaryTrendTooltipMetricLabel(metric)}：${formatTrendNumber(value, metric)}`
          }),
        ].join('<br/>')
      },
    },
    graphic: summaryRowsHaveTrendData(rows, metrics) ? undefined : noDataTrendGraphic(),
    xAxis: {
      type: 'category',
      boundaryGap: false,
      data: trendAxisLabels(rows, granularity),
      axisLabel: { color: '#667085', fontSize: 10, hideOverlap: true },
      axisLine: { show: false },
      axisTick: { show: false },
    },
    yAxis: [
      {
        type: 'value',
        min: 0,
        minInterval: 1,
        splitLine: { show: false },
        axisLine: { show: false },
        axisTick: { show: false },
        axisLabel: {
          color: '#667085',
          fontSize: 10,
          formatter: (value) => Math.round(value),
        },
      },
      {
        type: 'value',
        min: 0,
        max: 1,
        splitLine: { show: false },
        axisLine: { show: false },
        axisTick: { show: false },
        axisLabel: {
          color: '#667085',
          fontSize: 10,
          formatter: (value) => formatTrendRatio(value),
        },
      },
    ],
    series: metrics.map((metric) => {
      const color = summaryTrendSeriesColor(metric)
      return {
        name: metric.seriesLabel || metric.label,
        type: 'line',
        yAxisIndex: isTrendPercentMetric(metric) ? 1 : 0,
        data: rows.map((row) => summaryTrendDataPointValue(row, metric.key)),
        smooth: true,
        showSymbol: false,
        connectNulls: false,
        symbolSize: 5,
        lineStyle: { color, width: 2 },
        itemStyle: { color },
        ...(isTrendPercentMetric(metric)
          ? { lineStyle: { color, width: 2, type: 'dashed' } }
          : { areaStyle: { color, opacity: 0.08 } }),
      }
    }),
  }
}

const SummaryTrendGroup = ({
  group,
  points,
  granularity,
  loading,
  error,
  onRetry,
  trendRangeLabel,
}) => {
  const rows = useMemo(
    () => buildSummaryCombinedTrendRows(points, group.metrics),
    [group.metrics, points],
  )
  const option = useMemo(
    () => buildSummaryTrendOption(rows, group, granularity),
    [granularity, group, rows],
  )

  if (loading) {
    return (
      <div className="sho-summary-trend-state">
        <Spin size="small" />
        <span>趋势加载中</span>
      </div>
    )
  }

  if (error) {
    return <TrendFailureState description="无法读取历史快照" error={error} onRetry={onRetry} />
  }

  if (!points.length) {
    return <div className="sho-summary-trend-state">暂无趋势快照</div>
  }

  return (
    <div className="sho-summary-trends">
      <div className="sho-summary-trend-item">
        <div className="sho-summary-trend-head">
          <span>{`${trendRangeLabel}趋势图`}</span>
          <div className="sho-summary-trend-legend">
            {group.metrics.map((legendMetric) => (
              <span
                className="sho-summary-trend-legend-item"
                key={legendMetric.key}
                style={{ color: summaryTrendSeriesColor(legendMetric) }}
              >
                <i
                  className={
                    isTrendPercentMetric(legendMetric)
                      ? 'sho-summary-trend-legend-mark sho-summary-trend-legend-mark-percent'
                      : 'sho-summary-trend-legend-mark'
                  }
                />
                {legendMetric.seriesLabel || legendMetric.label}
              </span>
            ))}
          </div>
        </div>
        <ReactECharts notMerge option={option} style={{ width: '100%', height: 118 }} />
      </div>
    </div>
  )
}

SummaryTrendGroup.propTypes = {
  group: PropTypes.shape({
    chartLabel: PropTypes.string,
    label: PropTypes.string,
    metrics: PropTypes.arrayOf(PropTypes.object).isRequired,
  }).isRequired,
  points: PropTypes.arrayOf(PropTypes.object).isRequired,
  granularity: PropTypes.string.isRequired,
  loading: PropTypes.bool.isRequired,
  error: PropTypes.string,
  onRetry: PropTypes.func,
  trendRangeLabel: PropTypes.string.isRequired,
}

export default function OverviewSummary({
  summaryCards,
  summaryCardIcons,
  summaryTrendGroupByKey,
  statusDistribution,
  trendError,
  trendGranularity,
  trendLoading,
  trendPoints,
  trendRangeLabel,
  onTrendRetry,
}) {
  return (
    <div className="sho-summary">
      {summaryCards.map((card) => {
        const trendGroup = summaryTrendGroupByKey[card.key]

        return (
          <SummaryMetricCard
            denominator={card.denominator}
            highlightNumerator={card.highlightNumerator}
            icon={summaryCardIcons[card.key] || null}
            key={card.key}
            label={card.label}
            help={card.help}
            numerator={card.numerator}
            percent={card.percent}
            tone={card.tone}
            value={card.value}
            visual={card.visual}
          >
            {trendGroup ? (
              <SummaryTrendGroup
                error={trendError}
                granularity={trendGranularity}
                group={trendGroup}
                loading={trendLoading}
                onRetry={onTrendRetry}
                points={trendPoints}
                trendRangeLabel={trendRangeLabel}
              />
            ) : card.distribution ? (
              <StatusDistributionBar distribution={statusDistribution} />
            ) : null}
          </SummaryMetricCard>
        )
      })}
    </div>
  )
}

OverviewSummary.propTypes = {
  summaryCards: PropTypes.arrayOf(PropTypes.object).isRequired,
  summaryCardIcons: PropTypes.objectOf(PropTypes.node).isRequired,
  summaryTrendGroupByKey: PropTypes.objectOf(PropTypes.object).isRequired,
  statusDistribution: PropTypes.shape({
    total: PropTypes.number,
    red: PropTypes.number,
    yellow: PropTypes.number,
    green: PropTypes.number,
    gray: PropTypes.number,
  }).isRequired,
  trendError: PropTypes.string.isRequired,
  trendGranularity: PropTypes.string.isRequired,
  trendLoading: PropTypes.bool.isRequired,
  trendPoints: PropTypes.arrayOf(PropTypes.object).isRequired,
  trendRangeLabel: PropTypes.string.isRequired,
  onTrendRetry: PropTypes.func.isRequired,
}
