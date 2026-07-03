import React from 'react'
import PropTypes from 'prop-types'
import { Button } from 'antd'
import { Activity, ArrowLeft, CircleAlert, TriangleAlert } from 'lucide-react'

import { DetailLoadingStatus } from './DetailLoading.jsx'
import { HealthRing } from './HealthStatus.jsx'

export default function DetailHeader({
  HealthScoreRangeHelpIcon,
  abnormalServiceRatio,
  detailHeaderLoading,
  onCloseDetail,
  redMetricTags,
  scoreThresholds,
  selectedSystem,
}) {
  return (
    <div className="sho-detail-compact-header">
      <div className="sho-detail-compact-inner">
        <div className="sho-detail-compact-system">
          <Button
            className="sho-detail-compact-back"
            icon={<ArrowLeft size={12} />}
            size="small"
            type="text"
            onClick={onCloseDetail}
          >
            返回
          </Button>
          <div className="sho-detail-compact-title">
            <strong>{selectedSystem.name}</strong>
          </div>
        </div>
        {detailHeaderLoading ? (
          <DetailLoadingStatus className="sho-detail-compact-loading" />
        ) : (
          <>
            <span className="sho-detail-compact-divider" aria-hidden="true" />
            <div className="sho-detail-compact-score">
              <div className="sho-detail-compact-score-label">
                <HealthScoreRangeHelpIcon scoreThresholds={scoreThresholds} />
                <span>健康度</span>
              </div>
              <HealthRing
                score={selectedSystem.score}
                scoreThresholds={scoreThresholds}
                size="mini"
                status={selectedSystem.status}
              />
            </div>
            <span className="sho-detail-compact-divider" aria-hidden="true" />
            <div className="sho-detail-compact-service-count">
              <div className="sho-detail-compact-service-label sho-detail-compact-service-label-critical">
                <TriangleAlert size={18} />
                <span>异常服务</span>
              </div>
              <div
                className="sho-detail-compact-service-value sho-summary-value"
                aria-label={`异常服务 ${abnormalServiceRatio.abnormal} / ${abnormalServiceRatio.total}`}
              >
                <span className="sho-summary-value-highlight-critical">
                  {abnormalServiceRatio.abnormal}
                </span>
                <span>/ {abnormalServiceRatio.total}</span>
              </div>
            </div>
            <span className="sho-detail-compact-divider" aria-hidden="true" />
            <div className="sho-detail-compact-service-count">
              <div className="sho-detail-compact-service-label sho-detail-compact-service-label-warning">
                <CircleAlert size={18} />
                <span>关注服务</span>
              </div>
              <div
                className="sho-detail-compact-service-value sho-summary-value"
                aria-label={`关注服务 ${abnormalServiceRatio.attention} / ${abnormalServiceRatio.total}`}
              >
                <span className="sho-summary-value-highlight-warning">
                  {abnormalServiceRatio.attention}
                </span>
                <span>/ {abnormalServiceRatio.total}</span>
              </div>
            </div>
            <span className="sho-detail-compact-divider" aria-hidden="true" />
            <div className="sho-detail-compact-exceptions">
              <div className="sho-detail-compact-exception-label">
                <Activity size={18} />
                <span>异常指标</span>
              </div>
              <div className="sho-tags sho-detail-compact-tags">{redMetricTags}</div>
            </div>
          </>
        )}
      </div>
    </div>
  )
}

DetailHeader.propTypes = {
  HealthScoreRangeHelpIcon: PropTypes.elementType.isRequired,
  abnormalServiceRatio: PropTypes.shape({
    abnormal: PropTypes.number,
    attention: PropTypes.number,
    total: PropTypes.number,
  }).isRequired,
  detailHeaderLoading: PropTypes.bool.isRequired,
  onCloseDetail: PropTypes.func.isRequired,
  redMetricTags: PropTypes.node.isRequired,
  scoreThresholds: PropTypes.shape({
    severeAbnormalScoreThreshold: PropTypes.number,
    generalAbnormalScoreThreshold: PropTypes.number,
  }).isRequired,
  selectedSystem: PropTypes.shape({
    name: PropTypes.string,
    score: PropTypes.number,
    status: PropTypes.string,
  }).isRequired,
}
