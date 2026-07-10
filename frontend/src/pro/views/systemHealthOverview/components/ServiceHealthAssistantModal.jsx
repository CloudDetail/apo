import React from 'react'
import PropTypes from 'prop-types'
import { Modal, Tabs } from 'antd'

import { serviceAiAssistantActionFor } from '../detailViewState.mjs'
import { buildPendingIssueAnalysisKey } from '../pendingIssueAnalysis.mjs'
import { HealthRing, Lamp, statusMeta } from './HealthStatus.jsx'
import { PendingIssueAnalysis } from './PendingIssueQueue.jsx'

const serviceAssistantTitle = (service) => {
  const analysisTitle = serviceAiAssistantActionFor(service).label
  return service?.name ? `${service.name} ${analysisTitle}` : analysisTitle
}

const ServiceAssistantLoading = () => (
  <div aria-live="polite" className="sho-service-ai-loading" role="status">
    <div className="sho-service-ai-loading-header">
      <span aria-hidden="true" className="sho-service-ai-loading-mark">
        <span />
        <span />
        <span />
      </span>
      <div className="sho-service-ai-loading-copy">
        <strong>AI 分析中</strong>
        <span>正在生成指标判断、可能原因和处理建议</span>
      </div>
    </div>
    <div aria-hidden="true" className="sho-service-ai-loading-preview">
      <span />
      <span />
      <span />
      <span />
    </div>
  </div>
)

const ServiceAssistantEmpty = () => (
  <div className="sho-service-ai-empty">
    <strong>当前服务暂无红黄指标</strong>
    <span>各项指标处于健康或无数据状态，暂不生成异常或调优分析。</span>
  </div>
)

export default function ServiceHealthAssistantModal({
  activeMetricKey,
  issueAnalyses,
  issues,
  metricLabels,
  onClose,
  onMetricChange,
  open,
  queryParams,
  scoreThresholds,
  service,
}) {
  const activeIssue =
    issues.find((issue) => issue.metricKey === activeMetricKey) || issues[0] || null

  const renderServiceAssistantAnalysis = (issue) => {
    if (!issue) return <ServiceAssistantEmpty />

    const analysisKey = buildPendingIssueAnalysisKey(issue, queryParams)
    const analysisState = issueAnalyses[analysisKey] || {}
    if (analysisState.loading || (!analysisState.data && !analysisState.error)) {
      return <ServiceAssistantLoading />
    }

    return <PendingIssueAnalysis analysisState={analysisState} analysisTitle="事件" isExpanded />
  }

  const tabItems = issues.map((issue) => ({
    key: issue.metricKey,
    label: (
      <span className="sho-service-ai-tab-label">
        <Lamp status={issue.status} />
        <span>{issue.metricLabel}</span>
      </span>
    ),
    children: renderServiceAssistantAnalysis(issue),
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
        renderServiceAssistantAnalysis(null)
      )}
    </Modal>
  )
}

ServiceHealthAssistantModal.propTypes = {
  activeMetricKey: PropTypes.string,
  issueAnalyses: PropTypes.objectOf(PropTypes.object).isRequired,
  issues: PropTypes.arrayOf(PropTypes.object).isRequired,
  metricLabels: PropTypes.arrayOf(PropTypes.array).isRequired,
  onClose: PropTypes.func.isRequired,
  onMetricChange: PropTypes.func.isRequired,
  open: PropTypes.bool.isRequired,
  queryParams: PropTypes.object,
  scoreThresholds: PropTypes.shape({
    severeAbnormalScoreThreshold: PropTypes.number,
    generalAbnormalScoreThreshold: PropTypes.number,
  }).isRequired,
  service: PropTypes.object,
}
