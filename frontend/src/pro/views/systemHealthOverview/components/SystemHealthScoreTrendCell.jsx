import React, { useMemo, useState } from 'react'
import PropTypes from 'prop-types'
import { Tooltip } from 'antd'
import ReactECharts from 'echarts-for-react'
import { CircleAlert } from 'lucide-react'

import {
  DEFAULT_HEALTH_SCORE_THRESHOLDS,
} from '../metricVisuals.mjs'
import {
  buildHealthScoreBaselineItems,
} from '../trendViewState.mjs'
import {
  formatTrendNumber,
  formatTrendTooltipLabel,
  healthScoreBaselineColor,
  historyTrendColors,
  noDataTrendGraphic,
  rowsHaveTrendData,
  trendAxisLabels,
  trendDataPointValue,
} from './TrendCommon.jsx'

const safeText = (value) => (value === undefined || value === null ? '' : `${value}`.trim())

const formatSystemHealthScoreTooltip = (rows, granularity, params = []) => {
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
    `健康度：${formatTrendNumber(row.value, 'systemHealthScore')}`,
  ].join('<br/>')
}

const buildSystemHealthScoreSparklineSegments = (rows = [], width = 116, height = 34) => {
  const safeRows = Array.isArray(rows) ? rows : []
  if (!safeRows.length) return []
  const xStep = safeRows.length > 1 ? width / (safeRows.length - 1) : 0
  const segments = []
  let current = []

  safeRows.forEach((row, index) => {
    const value = trendDataPointValue(row, 'value')
    if (value === null) {
      if (current.length) {
        segments.push(current)
        current = []
      }
      return
    }
    const x = safeRows.length > 1 ? index * xStep : width / 2
    const y = height - (Math.min(100, Math.max(0, value)) / 100) * height
    current.push({ x: x.toFixed(1), y: y.toFixed(1) })
  })

  if (current.length) {
    segments.push(current)
  }
  return segments.map((segment) => {
    const linePoints = segment.map((point) => `${point.x},${point.y}`).join(' ')
    const pathPoints = segment.map((point) => `${point.x},${point.y}`).join(' L ')
    const firstPoint = segment[0]
    const lastPoint = segment[segment.length - 1]

    return {
      areaPath: `M ${firstPoint.x},${height} L ${pathPoints} L ${lastPoint.x},${height} Z`,
      linePoints,
    }
  })
}

const buildSystemHealthScoreTooltipOption = (
  rows,
  granularity,
  scoreThresholds = DEFAULT_HEALTH_SCORE_THRESHOLDS,
) => {
  const baselineItems = buildHealthScoreBaselineItems(scoreThresholds)

  return {
    animation: false,
    grid: { left: 58, right: 72, top: 18, bottom: 36 },
    tooltip: {
      trigger: 'axis',
      confine: true,
      formatter: (params = []) => formatSystemHealthScoreTooltip(rows, granularity, params),
    },
    graphic: rowsHaveTrendData(rows, 'value') ? undefined : noDataTrendGraphic(),
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
      name: '健康度',
      nameLocation: 'middle',
      nameRotate: 90,
      nameGap: 42,
      nameTextStyle: { color: '#667085', fontSize: 11 },
      splitLine: { lineStyle: { color: 'rgba(102, 112, 133, 0.14)' } },
      axisLabel: { color: '#667085', fontSize: 11 },
      axisLine: { show: true, lineStyle: { color: 'rgba(102, 112, 133, 0.35)' } },
      axisTick: { show: false },
    },
    series: [
      {
        type: 'line',
        data: rows.map((row) => trendDataPointValue(row, 'value')),
        smooth: true,
        showSymbol: rows.length <= 20,
        connectNulls: false,
        symbolSize: 5,
        lineStyle: { color: historyTrendColors.score, width: 2 },
        itemStyle: { color: historyTrendColors.score },
        areaStyle: { color: historyTrendColors.score, opacity: 0.08 },
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
            lineStyle: { color: healthScoreBaselineColor(item), type: 'dashed', width: 1.4 },
            label: { color: healthScoreBaselineColor(item) },
          })),
        },
      },
    ],
  }
}

export default function SystemHealthScoreTrendCell({
  rows = [],
  loading,
  error,
  granularity,
  systemName,
  trendRangeLabel = '历史',
  scoreThresholds = DEFAULT_HEALTH_SCORE_THRESHOLDS,
}) {
  const [tooltipOpen, setTooltipOpen] = useState(false)
  const sparklineSegments = useMemo(() => buildSystemHealthScoreSparklineSegments(rows), [rows])
  const sparklineHasData = rowsHaveTrendData(rows, 'value')
  const tooltipOption = useMemo(
    () => (tooltipOpen ? buildSystemHealthScoreTooltipOption(rows, granularity, scoreThresholds) : null),
    [granularity, rows, scoreThresholds, tooltipOpen],
  )

  if (loading) {
    return <span className="sho-system-trend-state">加载中</span>
  }

  if (error) {
    return (
      <Tooltip title={error}>
        <span className="sho-system-trend-state sho-system-trend-error">
          <CircleAlert size={13} />
          <span>趋势失败</span>
        </span>
      </Tooltip>
    )
  }

  if (!rows.length) {
    return <span className="sho-system-trend-state">暂无数据</span>
  }

  const safeSystemName = safeText(systemName) || '系统'
  const tooltipTitle = `${safeSystemName}${trendRangeLabel}健康度趋势图`
  const tooltipChart = (
    <div className="sho-system-trend-tooltip-chart">
      <div className="sho-system-trend-tooltip-title">{tooltipTitle}</div>
      {tooltipOpen && tooltipOption ? (
        <ReactECharts notMerge option={tooltipOption} style={{ width: 420, height: 220 }} />
      ) : null}
    </div>
  )

  return (
    <Tooltip
      onOpenChange={setTooltipOpen}
      open={tooltipOpen}
      overlayClassName="sho-system-trend-tooltip"
      title={tooltipChart}
    >
      <div className="sho-system-trend-cell" aria-label="历史值趋势图">
        <svg
          aria-hidden="true"
          className="sho-system-trend-svg"
          focusable="false"
          preserveAspectRatio="none"
          viewBox="0 0 116 34"
        >
          {sparklineHasData ? (
            sparklineSegments.map((segment, index) => (
              <g key={`${segment.linePoints}-${index}`}>
                <path className="sho-system-trend-svg-area" d={segment.areaPath} />
                <polyline
                  className="sho-system-trend-svg-line"
                  fill="none"
                  points={segment.linePoints}
                />
              </g>
            ))
          ) : (
            <g>
              <line
                className="sho-system-trend-svg-empty-line"
                x1="8"
                x2="108"
                y1="18"
                y2="18"
              />
              <text className="sho-system-trend-svg-empty-text" x="58" y="14">
                无数据
              </text>
            </g>
          )}
        </svg>
      </div>
    </Tooltip>
  )
}

SystemHealthScoreTrendCell.propTypes = {
  rows: PropTypes.arrayOf(
    PropTypes.shape({
      timestamp: PropTypes.number.isRequired,
      value: PropTypes.number,
      stats: PropTypes.object,
    }),
  ),
  loading: PropTypes.bool,
  error: PropTypes.string,
  granularity: PropTypes.string,
  systemName: PropTypes.string,
  trendRangeLabel: PropTypes.string,
  scoreThresholds: PropTypes.shape({
    severeAbnormalScoreThreshold: PropTypes.number,
    generalAbnormalScoreThreshold: PropTypes.number,
  }),
}
