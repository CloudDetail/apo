import React from 'react'
import PropTypes from 'prop-types'
import { Button, InputNumber } from 'antd'

import { ThresholdStatusRing } from './HealthStatus.jsx'

export default function ThresholdSettingsContent({
  hasThresholdChange,
  normalizedDraft,
  onDraftChange,
  onSave,
  thresholdDraft,
  thresholdSaving,
}) {
  const severeThreshold = normalizedDraft.severeAbnormalScoreThreshold
  const generalThreshold = normalizedDraft.generalAbnormalScoreThreshold
  const severeRange = `0-${severeThreshold - 1}分`
  const generalRange = `${severeThreshold}-${generalThreshold - 1}分`
  const healthyRange = `${generalThreshold}-100分`

  return (
    <div className="sho-threshold-popover">
      <div className="sho-threshold-popover-title">健康异常判定</div>
      <div className="sho-threshold-editor">
        <div className="sho-threshold-inputs">
          <label className="sho-threshold-input-row">
            <span>健康度低于</span>
            <InputNumber
              aria-label="异常健康度阈值"
              max={99}
              min={1}
              precision={0}
              size="small"
              value={thresholdDraft.severeAbnormalScoreThreshold}
              onChange={(value) =>
                onDraftChange((current) => ({
                  ...current,
                  severeAbnormalScoreThreshold: value,
                }))
              }
              onPressEnter={onSave}
            />
            <span>分，异常</span>
          </label>
          <label className="sho-threshold-input-row">
            <span>健康度低于</span>
            <InputNumber
              aria-label="关注系统健康度阈值"
              max={100}
              min={2}
              precision={0}
              size="small"
              value={thresholdDraft.generalAbnormalScoreThreshold}
              onChange={(value) =>
                onDraftChange((current) => ({
                  ...current,
                  generalAbnormalScoreThreshold: value,
                }))
              }
              onPressEnter={onSave}
            />
            <span>分，关注</span>
          </label>
        </div>
        <div className="sho-threshold-abnormal-note">
          <div>系统健康度低于 {severeThreshold} 分，计为异常系统</div>
          <div>
            系统健康度 {severeThreshold}-{generalThreshold - 1} 分之间，计为关注系统
          </div>
        </div>
        <div className="sho-threshold-legend" aria-label="健康度分值说明">
          <div className="sho-threshold-legend-row">
            <ThresholdStatusRing status="red" />
            <span className="sho-threshold-legend-range">{severeRange}</span>
            <span className="sho-threshold-legend-label">异常</span>
          </div>
          <div className="sho-threshold-legend-row">
            <ThresholdStatusRing status="yellow" />
            <span className="sho-threshold-legend-range">{generalRange}</span>
            <span className="sho-threshold-legend-label">关注</span>
          </div>
          <div className="sho-threshold-legend-row">
            <ThresholdStatusRing status="green" />
            <span className="sho-threshold-legend-range">{healthyRange}</span>
            <span className="sho-threshold-legend-label">健康</span>
          </div>
        </div>
        <div className="sho-threshold-actions">
          <span className="sho-threshold-save-wrap">
            <Button
              aria-label="保存异常阈值"
              className="sho-threshold-save"
              disabled={!hasThresholdChange}
              loading={thresholdSaving}
              size="small"
              type="primary"
              onClick={onSave}
            >
              确定
            </Button>
          </span>
        </div>
      </div>
    </div>
  )
}

const thresholdShape = PropTypes.shape({
  severeAbnormalScoreThreshold: PropTypes.number,
  generalAbnormalScoreThreshold: PropTypes.number,
})

ThresholdSettingsContent.propTypes = {
  hasThresholdChange: PropTypes.bool.isRequired,
  normalizedDraft: thresholdShape.isRequired,
  onDraftChange: PropTypes.func.isRequired,
  onSave: PropTypes.func.isRequired,
  thresholdDraft: thresholdShape.isRequired,
  thresholdSaving: PropTypes.bool.isRequired,
}
