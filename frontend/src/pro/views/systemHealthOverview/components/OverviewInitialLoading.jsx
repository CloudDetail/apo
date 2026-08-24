import React from 'react'
import PropTypes from 'prop-types'
import { Spin } from 'antd'

export default function OverviewInitialLoading({ metricLabels }) {
  return (
    <div className="sho-overview-loading" role="status" aria-live="polite">
      <div className="sho-loading-summary" aria-hidden="true">
        {[0, 1, 2].map((item) => (
          <div className="sho-loading-card" key={item}>
            <span className="sho-loading-line sho-loading-line-title" />
            <span className="sho-loading-line sho-loading-line-value" />
          </div>
        ))}
      </div>

      <section className="sho-panel sho-overview-loading-panel">
        <div className="sho-panel-head">
          <div className="sho-panel-title">系统健康度</div>
          <span className="sho-muted">正在同步指标状态</span>
        </div>
        <div className="sho-overview-loading-body">
          <Spin />
          <div>
            <strong>正在加载健康度</strong>
            <span>获取系统、指标和待处理事项，完成后展示当前状态。</span>
          </div>
        </div>
        <div className="sho-loading-table" aria-hidden="true">
          {[0, 1, 2, 3].map((row) => (
            <div className="sho-loading-table-row" key={row}>
              <span className="sho-loading-table-cell sho-loading-table-cell-system" />
              <span className="sho-loading-table-cell sho-loading-table-cell-score" />
              <span className="sho-loading-table-cell sho-loading-table-cell-trend" />
              {metricLabels.map(([key]) => (
                <span className="sho-loading-table-cell sho-loading-table-cell-metric" key={key} />
              ))}
            </div>
          ))}
        </div>
      </section>

      <section className="sho-panel sho-overview-loading-pending">
        <div className="sho-panel-head">
          <div className="sho-panel-title">待处理事项</div>
          <span className="sho-muted">等待风险排序</span>
        </div>
        <div className="sho-loading-pending-table" aria-hidden="true">
          {[0, 1, 2].map((row) => (
            <div className="sho-loading-pending-row" key={row}>
              <span />
              <span />
              <span />
              <span />
              <span />
            </div>
          ))}
        </div>
      </section>
    </div>
  )
}

OverviewInitialLoading.propTypes = {
  metricLabels: PropTypes.arrayOf(PropTypes.array).isRequired,
}
