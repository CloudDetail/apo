import React from 'react'
import PropTypes from 'prop-types'
import { Tooltip } from 'antd'

import {
  DEFAULT_HEALTH_SCORE_THRESHOLDS,
  healthScoreStatus,
} from '../metricVisuals.mjs'

export const statusMeta = {
  green: { label: '健康', color: '#12b76a' },
  yellow: { label: '关注', color: '#f79009' },
  red: { label: '异常', color: '#f04438' },
  gray: { label: '无数据', color: '#98a2b3' },
}

export const statusOrder = ['red', 'yellow', 'green', 'gray']

export const ringColor = (score, scoreThresholds, status = 'gray') => {
  if (status === 'gray') return statusMeta.gray.color
  return statusMeta[healthScoreStatus(score, scoreThresholds)]?.color || statusMeta.gray.color
}

export const Lamp = ({ status }) => (
  <Tooltip title={statusMeta[status]?.label || '无数据'}>
    <span className={`sho-lamp sho-lamp-${status || 'gray'}`} />
  </Tooltip>
)

Lamp.propTypes = {
  status: PropTypes.string,
}

export const HealthRing = ({
  score,
  scoreThresholds = DEFAULT_HEALTH_SCORE_THRESHOLDS,
  size = 'small',
  status = 'gray',
}) => {
  const scoreNumber = Number(score)
  const safeScore = Number.isFinite(scoreNumber) ? scoreNumber : 0
  const hasHealthData = status !== 'gray'
  const displayScore = hasHealthData ? safeScore : '无数据'

  return (
    <Tooltip title={hasHealthData ? `${safeScore} 分` : '无数据'}>
      <span
        className={`sho-health-ring sho-health-ring-${size}${
          hasHealthData ? '' : ' sho-health-ring-no-data'
        }`}
        style={{
          '--sho-score': hasHealthData ? safeScore : 0,
          '--sho-ring': ringColor(safeScore, scoreThresholds, status),
        }}
      >
        {displayScore}
      </span>
    </Tooltip>
  )
}

HealthRing.propTypes = {
  score: PropTypes.number,
  scoreThresholds: PropTypes.shape({
    severeAbnormalScoreThreshold: PropTypes.number,
    generalAbnormalScoreThreshold: PropTypes.number,
  }),
  size: PropTypes.oneOf(['mini', 'small']),
  status: PropTypes.oneOf(['red', 'yellow', 'green', 'gray']),
}

export const ThresholdStatusRing = ({ status }) => (
  <span
    className={`sho-threshold-status-ring sho-threshold-status-ring-${status}`}
    style={{ '--sho-ring': statusMeta[status]?.color || statusMeta.gray.color }}
    aria-hidden="true"
  />
)

ThresholdStatusRing.propTypes = {
  status: PropTypes.oneOf(['red', 'yellow', 'green']).isRequired,
}

export const HealthTable = ({ children, className = '' }) => (
  <div className={`sho-table-wrap ${className}`}>
    <table>{children}</table>
  </div>
)

HealthTable.propTypes = {
  children: PropTypes.node,
  className: PropTypes.string,
}

export const Legend = () => (
  <div className="sho-legend">
    {Object.entries(statusMeta).map(([key, meta]) => (
      <span className="sho-legend-item" key={key}>
        <Lamp status={key} />
        <b>{meta.label}</b>
      </span>
    ))}
  </div>
)
