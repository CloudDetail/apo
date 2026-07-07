import React, { useMemo } from 'react'
import PropTypes from 'prop-types'
import { Tooltip } from 'antd'
import ReactECharts from 'echarts-for-react'
import { MetricsLineChartColor } from 'src/constants'

import {
  buildMetricTrendOption,
  metricVisualFor,
} from '../metricVisuals.mjs'
import { Lamp, statusMeta } from './HealthStatus.jsx'

const metricTrendColors = {
  latency: MetricsLineChartColor.latency,
  errorRate: MetricsLineChartColor.errorRate,
  logs: MetricsLineChartColor.logs,
  resourceShift: '#ff8a45',
}

const safeText = (value) => (value === undefined || value === null ? '' : `${value}`.trim())

const SystemMetricTrendChart = ({ chart }) => {
  const option = useMemo(
    () =>
      buildMetricTrendOption(chart, {
        color: metricTrendColors[chart.colorRole] || '#4992ff',
        warningColor: statusMeta.yellow.color,
        criticalColor: statusMeta.red.color,
        textColor: '#667085',
        gridColor: 'rgba(102, 112, 133, 0.16)',
      }),
    [chart],
  )

  return (
    <div className="sho-metric-trend" aria-label={`${chart.key}趋势图`}>
      <ReactECharts notMerge option={option} style={{ width: '100%', height: 72 }} />
    </div>
  )
}

SystemMetricTrendChart.propTypes = {
  chart: PropTypes.shape({
    key: PropTypes.string.isRequired,
    colorRole: PropTypes.string.isRequired,
    unit: PropTypes.string,
    series: PropTypes.arrayOf(
      PropTypes.shape({
        timestamp: PropTypes.number.isRequired,
        value: PropTypes.number,
      }),
    ).isRequired,
    thresholds: PropTypes.arrayOf(
      PropTypes.shape({
        label: PropTypes.string,
        value: PropTypes.number,
        status: PropTypes.string,
      }),
    ),
  }).isRequired,
}

const splitMetricDisplayValue = (value) => {
  const text = safeText(value) || '--'
  const spacedUnit = text.match(/^([+-]?\d+(?:\.\d+)?)(?:\s+)(.+)$/)
  if (spacedUnit) {
    return { main: spacedUnit[1], unit: spacedUnit[2] }
  }

  const attachedUnit = text.match(/^([+-]?\d+(?:\.\d+)?)([a-zA-Z][\w/%]*)$/)
  if (attachedUnit) {
    return { main: attachedUnit[1], unit: attachedUnit[2] }
  }

  return { main: text, unit: '' }
}

const MetricValue = ({ value }) => {
  const displayValue = splitMetricDisplayValue(value)
  const isEmptyValue = displayValue.main === '暂无数据' && !displayValue.unit

  return (
    <b className={isEmptyValue ? 'sho-metric-value-empty' : undefined}>
      <span className="sho-metric-value-main">{displayValue.main}</span>
      {displayValue.unit ? (
        <span className="sho-metric-value-unit">{displayValue.unit}</span>
      ) : null}
    </b>
  )
}

MetricValue.propTypes = {
  value: PropTypes.node.isRequired,
}

const metricEvidenceLabel = (label) => {
  switch (label) {
    case 'P95水位':
      return 'P95 水位（覆盖大多数实例）'
    case '高水位实例':
      return '高水位影响面'
    case '异常实例':
      return '异常实例数'
    case '平均水位':
      return '平均水位'
    case '错误率最高服务':
    case '响应最慢服务':
    case '锁等待最高服务':
    case '错误日志最多服务':
    case '活跃告警最多服务':
    case '波动最大服务':
    case '水位最高服务':
      return '主要影响服务'
    case '当前水位':
      return '当前系统水位'
    case '24h变化':
      return '24h 同比变化'
    case '请求数':
      return '总请求数'
    case '样本不足':
      return '样本量说明'
    case '失败数':
      return '失败请求数'
    default:
      return label || '关键数据'
  }
}

const SERVICE_LOCATOR_BADGE_LABELS = new Set([
  '水位最高服务',
  '主要贡献服务',
  '错误率最高服务',
  '响应最慢服务',
  '错误日志最多服务',
  '活跃告警最多服务',
  '锁等待最高服务',
  '波动最大服务',
])

const isServiceLocatorBadge = (badge) => SERVICE_LOCATOR_BADGE_LABELS.has(badge?.label)

const splitMetricTooltipBadges = (badges = []) => {
  const normalizedBadges = Array.isArray(badges) ? badges : []
  const locatorIndex = normalizedBadges.findIndex(isServiceLocatorBadge)
  if (locatorIndex < 0) {
    return {
      overallBadges: normalizedBadges,
      locatorBadges: [],
    }
  }
  return {
    overallBadges: normalizedBadges.slice(0, locatorIndex),
    locatorBadges: normalizedBadges.slice(locatorIndex),
  }
}

const metricPrimaryValueLabel = (metricKey, visual = {}) => {
  switch (metricKey) {
    case 'resource':
      return safeText(visual.helper) || '当前水位'
    case 'resourceShift':
      return '水位变化'
    case 'errorRate':
      return '错误率'
    case 'latency':
      return '当前平均延时'
    case 'errorLog':
      return '错误日志'
    case 'badRequests':
      return '活跃告警'
    case 'lockCost':
      return '平均锁等待'
    default:
      return '当前值'
  }
}

const metricLocatorLabel = (metricKey) => {
  switch (metricKey) {
    case 'resource':
      return '当前最高服务'
    case 'resourceShift':
      return '主要来源'
    case 'errorRate':
      return '失败贡献最高服务'
    case 'latency':
      return '延时来源'
    case 'errorLog':
      return '日志主要来源'
    case 'badRequests':
      return '告警来源'
    case 'lockCost':
      return '锁等待来源'
    default:
      return '定位线索'
  }
}

const metricBadgeFor = (badges, label) => badges.find((badge) => badge.label === label)

const metricBadgeHasRisk = (badges, label) => {
  const badge = metricBadgeFor(badges, label)
  return badge?.status === 'red' || badge?.status === 'yellow'
}

const metricTooltipConclusion = (metricKey, status, badges = []) => {
  if (status === 'gray') return '缺少数据组整体聚合数据，暂不做系统判定。'
  switch (metricKey) {
    case 'resource':
      if (status === 'red') return '整体资源水位过高，存在资源打满风险。'
      if (status === 'yellow') return '整体资源水位偏高，需要关注。'
      return '整体资源水位正常。'
    case 'resourceShift':
      if (status === 'red') return '资源水位较昨天同时间明显变化。'
      if (status === 'yellow') return '资源水位较昨天同时间上涨，需要关注。'
      return '资源水位相对昨天同时间稳定。'
    case 'errorRate':
      if (status === 'red') return '整体请求失败率过高。'
      if (status === 'yellow') return '当前窗口出现请求失败，需要关注。'
      return '整体请求失败率正常。'
    case 'latency':
      if (status !== 'green' && metricBadgeHasRisk(badges, '24h变化')) {
        return '平均响应延时比昨天同时间明显升高。'
      }
      if (status === 'red') return '平均响应延时过高。'
      if (status === 'yellow') return '平均响应延时需要关注。'
      return '平均响应延时稳定。'
    case 'errorLog':
      if (status === 'red') return '错误日志比昨天同时间明显增加。'
      if (status === 'yellow') return '当前窗口出现较多错误日志。'
      return '当前窗口错误日志处于稳定区间。'
    case 'badRequests':
      if (status === 'red') return '当前存在红色活跃告警。'
      if (status === 'yellow') return '当前有活跃告警需要关注。'
      return '当前没有活跃告警。'
    case 'lockCost':
      if (status === 'red') return '平均每个请求锁等待过高。'
      if (status === 'yellow') return '平均每个请求锁等待升高。'
      return '锁等待处于稳定区间。'
    default:
      return '系统指标处于当前状态。'
  }
}

const MetricDecisionTooltip = ({ metricKey, label, visual }) => {
  const statusLabel = statusMeta[visual.status]?.label || '无数据'
  const badges = Array.isArray(visual.badges) ? visual.badges : []
  const { overallBadges, locatorBadges } = splitMetricTooltipBadges(badges)
  const conclusion = metricTooltipConclusion(metricKey, visual.status, badges)
  const evidenceLimit = metricKey === 'resourceShift' ? 3 : 2
  const overallRows = [
    {
      label: metricPrimaryValueLabel(metricKey, visual),
      value: visual.value,
      status: visual.status,
    },
    ...overallBadges.slice(0, evidenceLimit).map((badge) => ({
      label: metricEvidenceLabel(badge.label),
      value: badge.value,
      status: badge.status,
    })),
  ]
  const locatorBadge = locatorBadges[0]

  return (
    <div className="sho-metric-decision-tooltip-content">
      <div className="sho-metric-decision-title">
        <span>{label}</span>
        <strong className={`sho-metric-decision-status sho-metric-decision-status-${visual.status}`}>
          {statusLabel}
        </strong>
      </div>
      <p className="sho-metric-decision-summary">{conclusion}</p>
      <div className="sho-metric-decision-list">
        {overallRows.map((row) => (
          <div className="sho-metric-decision-row" key={`${row.label}-${row.value}`}>
            <span>{row.label}</span>
            <strong className={`sho-metric-decision-status sho-metric-decision-status-${row.status || 'gray'}`}>
              {row.value}
            </strong>
          </div>
        ))}
        {locatorBadge ? (
          <div className="sho-metric-decision-row" key={`${locatorBadge.label}-${locatorBadge.value}`}>
            <span>{metricLocatorLabel(metricKey)}</span>
            <strong className="sho-metric-decision-locator-value">{locatorBadge.value}</strong>
          </div>
        ) : null}
      </div>
    </div>
  )
}

MetricDecisionTooltip.propTypes = {
  metricKey: PropTypes.string.isRequired,
  label: PropTypes.string.isRequired,
  visual: PropTypes.shape({
    status: PropTypes.string,
    value: PropTypes.node,
    note: PropTypes.string,
    badges: PropTypes.array,
  }).isRequired,
}

export default function MetricVisualTile({ metricKey, label, detail, status }) {
  const visual = metricVisualFor(metricKey, detail, status)
  const isEmptyValue = visual.value === '暂无数据'
  const decisionTooltip = (
    <MetricDecisionTooltip
      metricKey={metricKey}
      label={label}
      visual={visual}
    />
  )

  return (
    <Tooltip
      mouseEnterDelay={0.2}
      overlayClassName="sho-metric-decision-tooltip"
      title={decisionTooltip}
    >
      <div className={`sho-metric-tile sho-metric-tile-${visual.tone}`}>
        <div className="sho-metric-tile-head">
          <strong>{label}</strong>
          <Lamp status={visual.status} />
        </div>
        <div
          className={`sho-metric-visual sho-metric-visual-${visual.kind}${
            isEmptyValue ? ' sho-metric-visual-empty' : ''
          }`}
        >
          <div className="sho-metric-value-block">
            <MetricValue value={visual.value} />
          </div>
          {isEmptyValue ? null : visual.chart ? (
            <SystemMetricTrendChart chart={visual.chart} />
          ) : visual.kind === 'number' ? null : visual.kind === 'waterline' ? (
            <div className="sho-waterline" aria-label={`${label}${visual.helper}${visual.percent}%`}>
              {visual.showThresholdMarker ? <span className="sho-waterline-marker" /> : null}
              <span className="sho-waterline-fill" style={{ '--sho-fill': `${visual.percent}%` }} />
            </div>
          ) : (
            <div className="sho-risk-meter" aria-label={`${label}${visual.helper}${visual.percent}%`}>
              <span style={{ '--sho-fill': `${visual.percent}%` }} />
            </div>
          )}
        </div>
      </div>
    </Tooltip>
  )
}

MetricVisualTile.propTypes = {
  metricKey: PropTypes.string.isRequired,
  label: PropTypes.string.isRequired,
  detail: PropTypes.shape({
    status: PropTypes.string,
    value: PropTypes.string,
    note: PropTypes.string,
    unit: PropTypes.string,
    source: PropTypes.string,
    series: PropTypes.array,
    seriesUnit: PropTypes.string,
    seriesLabel: PropTypes.string,
    thresholds: PropTypes.array,
    badges: PropTypes.array,
  }).isRequired,
  status: PropTypes.string,
}
