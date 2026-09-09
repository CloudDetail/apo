import React from 'react'
import PropTypes from 'prop-types'
import { Button, Modal, Tabs } from 'antd'
import atomicIcon from 'src/pro/assets/atomic.png'

import { serviceAiAssistantActionFor } from '../detailViewState.mjs'
import { describePendingIssue } from '../pendingIssueDescription.mjs'
import { HealthRing, Lamp, statusMeta } from './HealthStatus.jsx'

const serviceAssistantTitle = (service) => {
  const analysisTitle = serviceAiAssistantActionFor(service).label
  return service?.name ? `${service.name} ${analysisTitle}` : analysisTitle
}

const ServiceAssistantEmpty = () => (
  <div className="sho-service-ai-empty">
    <strong>当前服务暂无可分析的异常指标</strong>
    <span>各项指标处于健康或无数据状态。</span>
  </div>
)

const ServiceAssistantIssue = ({ issue, onAnalyzeIssue }) => (
  <div className="sho-service-ai-launch">
    <div className="sho-service-ai-launch-summary">
      <div className="sho-service-ai-launch-title">
        <Lamp status={issue.status} />
        <strong>{issue.metricLabel}</strong>
      </div>
      <span>{describePendingIssue(issue)}</span>
    </div>
    <Button
      icon={<img alt="" aria-hidden="true" className="sho-ai-assistant-button-icon" src={atomicIcon} />}
      type="primary"
      onClick={() => onAnalyzeIssue(issue)}
    >
      在 Agent 中分析
    </Button>
  </div>
)

export default function ServiceHealthAssistantModal({
  activeMetricKey,
  issues,
  metricLabels,
  onClose,
  onAnalyzeIssue,
  onMetricChange,
  open,
  scoreThresholds,
  service,
}) {
  const activeIssue =
    issues.find((issue) => issue.metricKey === activeMetricKey) || issues[0] || null

  const tabItems = issues.map((issue) => ({
    key: issue.metricKey,
    label: (
      <span className="sho-service-ai-tab-label">
        <Lamp status={issue.status} />
        <span>{issue.metricLabel}</span>
      </span>
    ),
    children: <ServiceAssistantIssue issue={issue} onAnalyzeIssue={onAnalyzeIssue} />,
  }))

  return (
    <Modal
      className="sho-service-ai-modal"
      destroyOnClose
      footer={null}
      open={open}
      title={serviceAssistantTitle(service)}
      width={1040}
      wrapClassName="sho-service-ai-modal-wrap"
      onCancel={onClose}
    >
      <div className="sho-service-ai-summary">
        <div className="sho-service-ai-score">
          <span>健康度</span>
          <HealthRing
            score={service?.score}
            scoreThresholds={scoreThresholds}
            size="mini"
            status={service?.status}
          />
          <b>{statusMeta[service?.status]?.label || '无数据'}</b>
        </div>
        <div className="sho-service-ai-lamps">
          {metricLabels.map(([key, label]) => (
            <span className="sho-service-ai-lamp-item" key={key}>
              <Lamp status={service?.metrics?.[key] || 'gray'} />
              <b>{label}</b>
            </span>
          ))}
        </div>
      </div>

      {issues.length ? (
        <Tabs
          activeKey={activeIssue?.metricKey}
          className="sho-service-ai-tabs"
          items={tabItems}
          size="small"
          onChange={onMetricChange}
        />
      ) : (
        <ServiceAssistantEmpty />
      )}
    </Modal>
  )
}

ServiceHealthAssistantModal.propTypes = {
  activeMetricKey: PropTypes.string,
  issues: PropTypes.arrayOf(PropTypes.object).isRequired,
  metricLabels: PropTypes.arrayOf(PropTypes.array).isRequired,
  onClose: PropTypes.func.isRequired,
  onAnalyzeIssue: PropTypes.func.isRequired,
  onMetricChange: PropTypes.func.isRequired,
  open: PropTypes.bool.isRequired,
  scoreThresholds: PropTypes.shape({
    severeAbnormalScoreThreshold: PropTypes.number,
    generalAbnormalScoreThreshold: PropTypes.number,
  }).isRequired,
  service: PropTypes.object,
}
