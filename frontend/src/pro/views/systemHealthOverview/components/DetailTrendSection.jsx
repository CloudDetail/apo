import React, { useMemo } from 'react'
import PropTypes from 'prop-types'
import { Select, Spin } from 'antd'
import ReactECharts from 'echarts-for-react'

import {
  DEFAULT_HEALTH_SCORE_THRESHOLDS,
} from '../metricVisuals.mjs'
import {
  buildHealthScoreBaselineItems,
  trendRangeOptions,
} from '../trendViewState.mjs'
import {
  detailTrendServiceColors,
  formatTrendNumber,
  formatTrendRatio,
  formatTrendTooltipLabel,
  healthScoreBaselineColor,
  historyTrendColors,
  noDataTrendGraphic,
  rowsHaveTrendData,
  trendAxisLabels,
  trendDataPointValue,
  TrendFailureState,
} from './TrendCommon.jsx'

const formatServiceTrendCount = (value) => {
  const number = Number(value)
  if (!Number.isFinite(number)) return '0'
  return `${Math.max(0, Math.round(number))}`
}

const detailServiceTrendTotalCount = (row) =>
  Object.values(row?.serviceCounts || {}).reduce((sum, value) => {
    const number = Number(value)
    return Number.isFinite(number) ? sum + number : sum
  }, 0)

const detailServiceTrendLegend = [
  { key: 'abnormal', label: '异常', color: detailTrendServiceColors.abnormal, kind: 'bar' },
  {
    key: 'attention',
    label: '关注',
    color: detailTrendServiceColors.attention,
    kind: 'bar',
  },
  { key: 'healthy', label: '健康', color: detailTrendServiceColors.healthy, kind: 'bar' },
  { key: 'noData', label: '无数据', color: detailTrendServiceColors.noData, kind: 'bar' },
]

const formatDetailHealthScoreTrendTooltip = (rows, granularity, params = []) => {
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
    `健康度分数：${formatTrendNumber(row.score, 'systemHealthScore')}`,
  ].join('<br/>')
}

const formatDetailServiceStatusTrendTooltip = (rows, granularity, params = []) => {
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
    formatDetailServiceTotalTooltipLine(row),
    ...detailServiceTrendLegend.map((item) => formatDetailServiceStatusTooltipLine(item, row)),
  ].join('<br/>')
}

const formatDetailServiceTotalTooltipLine = (row) => {
  if (row.serviceDistributionMode === 'average') {
    return '<span class="sho-detail-service-tooltip-line sho-detail-service-tooltip-total"><span>统计口径：时段平均占比</span></span>'
  }

  const total = formatServiceTrendCount(detailServiceTrendTotalCount(row))

  return `<span class="sho-detail-service-tooltip-line sho-detail-service-tooltip-total"><span>总服务数量：${total} 个</span></span>`
}

const formatDetailServiceStatusTooltipLine = (item, row) => {
  const ratio = formatTrendRatio(row.serviceRatios[item.key])
  if (row.serviceDistributionMode === 'average') {
    return `<span class="sho-detail-service-tooltip-line"><span class="sho-detail-service-tooltip-dot" style="background:${item.color}"></span><span>${item.label}：${ratio}</span></span>`
  }

  const count = formatServiceTrendCount(row.serviceCounts?.[item.key])

  return `<span class="sho-detail-service-tooltip-line"><span class="sho-detail-service-tooltip-dot" style="background:${item.color}"></span><span>${item.label}：${count} 个（${ratio}）</span></span>`
}

const buildDetailHealthScoreTrendOption = (
  rows,
  granularity,
  scoreThresholds = DEFAULT_HEALTH_SCORE_THRESHOLDS,
) => {
  const baselineItems = buildHealthScoreBaselineItems(scoreThresholds).map((item) => ({
    ...item,
    color: healthScoreBaselineColor(item),
    kind: 'baseline',
  }))

  return {
    animation: false,
    grid: { left: 38, right: 72, top: 16, bottom: 30 },
    tooltip: {
      trigger: 'axis',
      confine: true,
      formatter: (params = []) => formatDetailHealthScoreTrendTooltip(rows, granularity, params),
    },
    graphic: rowsHaveTrendData(rows, 'score') ? undefined : noDataTrendGraphic(),
    xAxis: {
      type: 'category',
      boundaryGap: false,
      data: trendAxisLabels(rows, granularity),
      axisLabel: { color: '#667085', fontSize: 11, hideOverlap: true },
      axisLine: { lineStyle: { color: 'rgba(102, 112, 133, 0.35)' } },
      axisTick: { show: false },
    },
    yAxis: {
      type: 'value',
      min: 0,
      max: 100,
      splitLine: { lineStyle: { color: 'rgba(102, 112, 133, 0.12)' } },
      axisLabel: { color: '#667085', fontSize: 11 },
      axisLine: { show: false },
      axisTick: { show: false },
    },
    series: [
      {
        name: '健康度分数趋势',
        type: 'line',
        data: rows.map((row) => trendDataPointValue(row, 'score')),
        smooth: true,
        showSymbol: rows.length <= 20,
        connectNulls: false,
        symbolSize: 5,
        lineStyle: { color: historyTrendColors.score, width: 2.4 },
        itemStyle: { color: historyTrendColors.score },
        areaStyle: { color: historyTrendColors.score, opacity: 0.12 },
        markLine: {
          symbol: 'none',
          silent: true,
          label: {
            position: 'end',
            fontSize: 11,
            formatter: ({ name }) => name,
          },
          data: baselineItems.map((item) => ({
            name: `${item.label} ${item.value}`,
            yAxis: item.value,
            lineStyle: { color: item.color, type: 'dashed', width: 1.6 },
            label: { color: item.color },
          })),
        },
      },
    ],
  }
}

const buildDetailServiceStatusTrendOption = (rows, granularity) => ({
  animation: false,
  grid: { left: 38, right: 18, top: 16, bottom: 30 },
  tooltip: {
    trigger: 'axis',
    confine: true,
    formatter: (params = []) => formatDetailServiceStatusTrendTooltip(rows, granularity, params),
  },
  xAxis: {
    type: 'category',
    boundaryGap: true,
    data: trendAxisLabels(rows, granularity),
    axisLabel: { color: '#667085', fontSize: 11, hideOverlap: true },
    axisLine: { lineStyle: { color: 'rgba(102, 112, 133, 0.35)' } },
    axisTick: { show: false },
  },
  yAxis: {
    type: 'value',
    min: 0,
    max: 1,
    splitLine: { lineStyle: { color: 'rgba(102, 112, 133, 0.12)' } },
    axisLabel: {
      color: '#667085',
      fontSize: 11,
      formatter: (value) => formatTrendRatio(value),
    },
    axisLine: { show: false },
    axisTick: { show: false },
  },
  series: [
    {
      name: '异常',
      type: 'bar',
      stack: 'serviceRatio',
      barMaxWidth: 24,
      data: rows.map((row) => (row.missingSnapshot ? null : row.serviceRatios.abnormal)),
      itemStyle: { color: detailTrendServiceColors.abnormal },
    },
    {
      name: '关注',
      type: 'bar',
      stack: 'serviceRatio',
      barMaxWidth: 24,
      data: rows.map((row) => (row.missingSnapshot ? null : row.serviceRatios.attention)),
      itemStyle: { color: detailTrendServiceColors.attention },
    },
    {
      name: '健康',
      type: 'bar',
      stack: 'serviceRatio',
      barMaxWidth: 24,
      data: rows.map((row) => (row.missingSnapshot ? null : row.serviceRatios.healthy)),
      itemStyle: { color: detailTrendServiceColors.healthy },
    },
    {
      name: '无数据',
      type: 'bar',
      stack: 'serviceRatio',
      barMaxWidth: 24,
      data: rows.map((row) => (row.missingSnapshot ? null : row.serviceRatios.noData)),
      itemStyle: { color: detailTrendServiceColors.noData },
    },
  ],
})

const DetailTrendState = ({ loading, error, rows, onRetry }) => {
  if (loading) {
    return (
      <div className="sho-detail-trend-state">
        <Spin size="small" />
        <span>趋势加载中</span>
      </div>
    )
  }

  if (error) {
    return <TrendFailureState description="无法读取历史快照" error={error} onRetry={onRetry} />
  }

  if (!rows.length) {
    return <div className="sho-detail-trend-state">暂无趋势快照</div>
  }

  return null
}

DetailTrendState.propTypes = {
  loading: PropTypes.bool.isRequired,
  error: PropTypes.string,
  rows: PropTypes.arrayOf(PropTypes.object).isRequired,
  onRetry: PropTypes.func,
}

const DetailTrendLegend = ({ items }) => (
  <div className="sho-detail-trend-legend">
    {items.map((item) => (
      <span className="sho-detail-trend-legend-item" key={item.key} style={{ color: item.color }}>
        <i className={`sho-detail-trend-legend-mark sho-detail-trend-legend-mark-${item.kind}`} />
        {item.label}
      </span>
    ))}
  </div>
)

DetailTrendLegend.propTypes = {
  items: PropTypes.arrayOf(
    PropTypes.shape({
      key: PropTypes.string.isRequired,
      label: PropTypes.string.isRequired,
      color: PropTypes.string.isRequired,
      kind: PropTypes.string.isRequired,
    }),
  ).isRequired,
}

const detailTrendRowsPropType = PropTypes.arrayOf(
  PropTypes.shape({
    timestamp: PropTypes.number.isRequired,
    missingSnapshot: PropTypes.bool,
    score: PropTypes.number,
    serviceDistributionMode: PropTypes.string,
    serviceCounts: PropTypes.shape({
      abnormal: PropTypes.number,
      attention: PropTypes.number,
      healthy: PropTypes.number,
      noData: PropTypes.number,
    }).isRequired,
    serviceRatios: PropTypes.shape({
      abnormal: PropTypes.number,
      attention: PropTypes.number,
      healthy: PropTypes.number,
      noData: PropTypes.number,
    }).isRequired,
  }),
)

const DetailHealthScoreTrendChart = ({
  rows = [],
  loading,
  error,
  granularity,
  onRetry,
  scoreThresholds = DEFAULT_HEALTH_SCORE_THRESHOLDS,
}) => {
  const option = useMemo(
    () => buildDetailHealthScoreTrendOption(rows, granularity, scoreThresholds),
    [granularity, rows, scoreThresholds],
  )
  const state = <DetailTrendState error={error} loading={loading} onRetry={onRetry} rows={rows} />

  return (
    <div className="sho-detail-trend-block">
      <div className="sho-detail-trend-block-head">
        <span>健康度分数趋势</span>
      </div>
      {loading || error || !rows.length ? (
        state
      ) : (
        <div className="sho-detail-trend-chart" aria-label="健康度分数面积趋势图">
          <ReactECharts notMerge option={option} style={{ width: '100%', height: 180 }} />
        </div>
      )}
    </div>
  )
}

DetailHealthScoreTrendChart.propTypes = {
  rows: detailTrendRowsPropType,
  loading: PropTypes.bool.isRequired,
  error: PropTypes.string,
  granularity: PropTypes.string.isRequired,
  onRetry: PropTypes.func,
  scoreThresholds: PropTypes.shape({
    severeAbnormalScoreThreshold: PropTypes.number,
    generalAbnormalScoreThreshold: PropTypes.number,
  }),
}

const DetailServiceStatusTrendChart = ({ rows = [], loading, error, granularity, onRetry }) => {
  const option = useMemo(
    () => buildDetailServiceStatusTrendOption(rows, granularity),
    [granularity, rows],
  )
  const state = <DetailTrendState error={error} loading={loading} onRetry={onRetry} rows={rows} />

  return (
    <div className="sho-detail-trend-block">
      <div className="sho-detail-trend-block-head">
        <span>服务状态分布趋势</span>
        <DetailTrendLegend items={detailServiceTrendLegend} />
      </div>
      {loading || error || !rows.length ? (
        state
      ) : (
        <div className="sho-detail-trend-chart" aria-label="服务状态分布趋势图">
          <ReactECharts notMerge option={option} style={{ width: '100%', height: 180 }} />
        </div>
      )}
    </div>
  )
}

DetailServiceStatusTrendChart.propTypes = {
  rows: detailTrendRowsPropType,
  loading: PropTypes.bool.isRequired,
  error: PropTypes.string,
  granularity: PropTypes.string.isRequired,
  onRetry: PropTypes.func,
}

export default function DetailTrendSection({
  detailTrendError,
  detailTrendGranularity,
  detailTrendLoading,
  detailTrendRows,
  onRangeChange,
  onRetry,
  scoreThresholds,
  trendRange,
}) {
  return (
    <section className="sho-panel sho-detail-trend-panel">
      <div className="sho-panel-head">
        <div className="sho-panel-title">趋势分析</div>
        <div className="sho-detail-trend-head-meta">
          <Select
            aria-label="趋势分析时间范围"
            className="sho-detail-trend-range-select"
            options={trendRangeOptions.map((option) => ({
              label: option.label,
              value: option.key,
            }))}
            size="small"
            value={trendRange}
            onChange={onRangeChange}
          />
        </div>
      </div>
      <div className="sho-detail-trend-grid">
        <DetailHealthScoreTrendChart
          error={detailTrendError}
          granularity={detailTrendGranularity}
          loading={detailTrendLoading}
          onRetry={onRetry}
          rows={detailTrendRows}
          scoreThresholds={scoreThresholds}
        />
        <DetailServiceStatusTrendChart
          error={detailTrendError}
          granularity={detailTrendGranularity}
          loading={detailTrendLoading}
          onRetry={onRetry}
          rows={detailTrendRows}
        />
      </div>
    </section>
  )
}

DetailTrendSection.propTypes = {
  detailTrendError: PropTypes.string.isRequired,
  detailTrendGranularity: PropTypes.string.isRequired,
  detailTrendLoading: PropTypes.bool.isRequired,
  detailTrendRows: detailTrendRowsPropType.isRequired,
  onRangeChange: PropTypes.func.isRequired,
  onRetry: PropTypes.func.isRequired,
  scoreThresholds: PropTypes.shape({
    severeAbnormalScoreThreshold: PropTypes.number,
    generalAbnormalScoreThreshold: PropTypes.number,
  }).isRequired,
  trendRange: PropTypes.string.isRequired,
}
