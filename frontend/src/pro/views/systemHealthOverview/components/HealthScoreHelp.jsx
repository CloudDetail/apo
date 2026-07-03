import React from 'react'
import PropTypes from 'prop-types'
import { Tooltip } from 'antd'
import { HelpCircle } from 'lucide-react'

import {
  DEFAULT_HEALTH_SCORE_THRESHOLDS,
  normalizeHealthScoreThresholds,
} from '../metricVisuals.mjs'
import { ThresholdStatusRing } from './HealthStatus.jsx'

export const HealthScoreCalculationTooltip = () => (
  <div className="sho-health-score-help">
    由资源水位、资源波动、错误率、响应延时、错误日志、活跃告警、锁等待综合计算
  </div>
)

const healthScoreRangesFor = (scoreThresholds = DEFAULT_HEALTH_SCORE_THRESHOLDS) => {
  const thresholds = normalizeHealthScoreThresholds(scoreThresholds)
  return [
    {
      status: 'red',
      range: `0-${thresholds.severeAbnormalScoreThreshold - 1}分`,
      label: '异常',
    },
    {
      status: 'yellow',
      range: `${thresholds.severeAbnormalScoreThreshold}-${thresholds.generalAbnormalScoreThreshold - 1}分`,
      label: '关注',
    },
    {
      status: 'green',
      range: `${thresholds.generalAbnormalScoreThreshold}-100分`,
      label: '健康',
    },
  ]
}

const HealthScoreRangeTooltip = ({ scoreThresholds = DEFAULT_HEALTH_SCORE_THRESHOLDS }) => (
  <div className="sho-health-score-range-help">
    {healthScoreRangesFor(scoreThresholds).map((item) => (
      <div className="sho-health-score-range-row" key={item.status}>
        <ThresholdStatusRing status={item.status} />
        <span className="sho-health-score-range-text">{item.range}</span>
        <span className="sho-health-score-range-label">{item.label}</span>
      </div>
    ))}
  </div>
)

HealthScoreRangeTooltip.propTypes = {
  scoreThresholds: PropTypes.shape({
    severeAbnormalScoreThreshold: PropTypes.number,
    generalAbnormalScoreThreshold: PropTypes.number,
  }),
}

export const HealthScoreRangeHelpIcon = ({ scoreThresholds }) => (
  <Tooltip
    title={<HealthScoreRangeTooltip scoreThresholds={scoreThresholds} />}
    overlayClassName="sho-health-score-range-popover"
  >
    <HelpCircle size={15} className="sho-help-icon" />
  </Tooltip>
)

HealthScoreRangeHelpIcon.propTypes = {
  scoreThresholds: PropTypes.shape({
    severeAbnormalScoreThreshold: PropTypes.number,
    generalAbnormalScoreThreshold: PropTypes.number,
  }),
}

export const HealthScoreColumnTitle = ({ scoreThresholds }) => (
  <span className="sho-health-score-title">
    <span>健康度</span>
    <HealthScoreRangeHelpIcon scoreThresholds={scoreThresholds} />
  </span>
)

HealthScoreColumnTitle.propTypes = {
  scoreThresholds: PropTypes.shape({
    severeAbnormalScoreThreshold: PropTypes.number,
    generalAbnormalScoreThreshold: PropTypes.number,
  }),
}
