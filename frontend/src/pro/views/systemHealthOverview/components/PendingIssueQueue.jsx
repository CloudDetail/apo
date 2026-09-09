import React from 'react'
import PropTypes from 'prop-types'
import { Alert, Button, Pagination, Spin, Tooltip } from 'antd'
import atomicIcon from 'src/pro/assets/atomic.png'
import { HelpCircle } from 'lucide-react'

import { describePendingIssue } from '../pendingIssueDescription.mjs'
import { Lamp } from './HealthStatus.jsx'

const pendingIssueSystemDetailTarget = (system = {}) => {
  const score = system.score ?? system.systemScore
  return {
    groupId: system.groupId,
    name: system.name || system.systemName || '未命名系统',
    score: Number.isFinite(Number(score)) ? Number(score) : 0,
    status: system.status || system.systemStatus || 'gray',
    metrics: system.metrics || {},
    metricDetails: system.metricDetails || {},
    services: Array.isArray(system.services) ? system.services : [],
    serviceTotal: Number.isFinite(Number(system.serviceTotal)) ? Number(system.serviceTotal) : 0,
    abnormalServiceCount: Number.isFinite(Number(system.abnormalServiceCount))
      ? Number(system.abnormalServiceCount)
      : 0,
    attentionServiceCount: Number.isFinite(Number(system.attentionServiceCount))
      ? Number(system.attentionServiceCount)
      : 0,
  }
}

const pendingIssueSystemTargets = (issue = {}) => {
  const systems = Array.isArray(issue.systems) && issue.systems.length ? issue.systems : [issue]
  return systems.map(pendingIssueSystemDetailTarget)
}

export default function PendingIssueQueue({
  onOpenAgent,
  onOpenSystemDetail,
  onPageChange,
  pendingIssueError,
  pendingIssueLoading,
  pendingIssuePage,
  pendingIssuePageSize,
  pendingIssueRows,
  pendingIssueTotal,
  pendingServiceCount,
}) {
  return (
    <section className="sho-panel sho-pending-panel">
      <div className="sho-panel-head">
        <div className="sho-panel-title">
          待处理事项
          <Tooltip title="当前用户有权访问的所有系统中，按风险排序的红灯和黄灯服务指标">
            <HelpCircle size={15} className="sho-help-icon" />
          </Tooltip>
        </div>
        <span className="sho-muted">
          {pendingServiceCount} 个服务 / 本页 {pendingIssueRows.length} 类异常 / 共{' '}
          {pendingIssueTotal} 条
        </span>
      </div>
      {pendingIssueLoading ? (
        <div className="sho-state-panel">
          <Spin size="small" />
          <span>正在加载待处理事项</span>
        </div>
      ) : pendingIssueError ? (
        <Alert message="待处理事项加载失败" description={pendingIssueError} type="error" showIcon />
      ) : pendingIssueRows.length ? (
        <div className="sho-pending-table-wrap">
          <table className="sho-pending-table">
            <colgroup>
              <col className="sho-pending-col-service" />
              <col className="sho-pending-col-issue" />
              <col className="sho-pending-col-system" />
              <col className="sho-pending-col-value" />
              <col className="sho-pending-col-action" />
            </colgroup>
            <thead>
              <tr>
                <th>服务</th>
                <th>异常</th>
                <th>系统</th>
                <th>当前态</th>
                <th>AI助手</th>
              </tr>
            </thead>
            <tbody>
              {pendingIssueRows.map((issue) => {
                return (
                  <React.Fragment key={issue.queueId}>
                    <tr className="sho-pending-row">
                      <td className="sho-pending-service-cell">
                        <div className="sho-pending-service-inline">
                          <span className="sho-pending-service-name" title={issue.serviceName}>
                            {issue.serviceName}
                          </span>
                        </div>
                      </td>
                      <td className="sho-pending-issue-cell">
                        <div className="sho-pending-issue-content">
                          <div className="sho-pending-issue-title">
                            <Lamp status={issue.status} />
                            <strong>{issue.metricLabel}</strong>
                          </div>
                          <small title={issue.note}>{describePendingIssue(issue)}</small>
                        </div>
                      </td>
                      <td className="sho-pending-system-cell">
                        <span className="sho-pending-system-list" title={issue.systemName}>
                          {pendingIssueSystemTargets(issue).map((system, index) => (
                            <React.Fragment key={`${system.groupId || system.name}-${index}`}>
                              {index > 0 && (
                                <span className="sho-pending-system-separator" aria-hidden="true">
                                  、
                                </span>
                              )}
                              <button
                                type="button"
                                className="sho-pending-system"
                                title={system.name}
                                aria-label={`查看${system.name}健康度详情`}
                                onClick={() =>
                                  onOpenSystemDetail(pendingIssueSystemDetailTarget(system))
                                }
                              >
                                {system.name}
                              </button>
                            </React.Fragment>
                          ))}
                        </span>
                      </td>
                      <td>
                        <span className="sho-pending-value">{issue.value}</span>
                      </td>
                      <td className="sho-pending-action-cell">
                        <Button
                          icon={
                            <img
                              alt=""
                              aria-hidden="true"
                              className="sho-ai-assistant-button-icon sho-pending-analysis-button-icon"
                              src={atomicIcon}
                            />
                          }
                          size="small"
                          onClick={() => onOpenAgent(issue)}
                        >
                          异常分析
                        </Button>
                      </td>
                    </tr>
                  </React.Fragment>
                )
              })}
            </tbody>
          </table>
        </div>
      ) : (
        <div className="sho-empty-block">暂无需要优先处理的系统问题</div>
      )}
      {pendingIssueTotal > pendingIssuePageSize && (
        <div className="sho-pending-pagination">
          <Pagination
            current={pendingIssuePage}
            pageSize={pendingIssuePageSize}
            pageSizeOptions={['10', '20', '50', '100']}
            showSizeChanger
            showTotal={(total) => `共 ${total} 条`}
            size="small"
            total={pendingIssueTotal}
            onChange={onPageChange}
          />
        </div>
      )}
    </section>
  )
}

PendingIssueQueue.propTypes = {
  onOpenAgent: PropTypes.func.isRequired,
  onOpenSystemDetail: PropTypes.func.isRequired,
  onPageChange: PropTypes.func.isRequired,
  pendingIssueError: PropTypes.string.isRequired,
  pendingIssueLoading: PropTypes.bool.isRequired,
  pendingIssuePage: PropTypes.number.isRequired,
  pendingIssuePageSize: PropTypes.number.isRequired,
  pendingIssueRows: PropTypes.arrayOf(PropTypes.object).isRequired,
  pendingIssueTotal: PropTypes.number.isRequired,
  pendingServiceCount: PropTypes.number.isRequired,
}
