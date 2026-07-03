import React from 'react'
import PropTypes from 'prop-types'

export const DetailLoadingStatus = ({ className = '' }) => (
  <span
    className={`sho-detail-loading-status${className ? ` ${className}` : ''}`}
    role="status"
    aria-live="polite"
  >
    <span className="sho-detail-loading-pulse" aria-hidden="true">
      <span />
      <span />
      <span />
    </span>
    <span>正在加载系统详情</span>
  </span>
)

DetailLoadingStatus.propTypes = {
  className: PropTypes.string,
}

export const DetailLoadingPanel = () => (
  <section className="sho-panel sho-detail-loading" role="status" aria-live="polite">
    <div className="sho-panel-head">
      <div className="sho-panel-title">系统详情</div>
      <DetailLoadingStatus className="sho-detail-loading-status-muted" />
    </div>
    <div className="sho-detail-loading-body" aria-hidden="true">
      <div className="sho-detail-loading-metrics">
        {[0, 1, 2, 3].map((item) => (
          <div className="sho-detail-loading-metric" key={item}>
            <div className="sho-detail-loading-metric-head">
              <span className="sho-loading-line sho-detail-loading-title-line" />
              <span className="sho-detail-loading-dot" />
            </div>
            <span className="sho-loading-line sho-detail-loading-value-line" />
            <span className="sho-loading-line sho-detail-loading-meter-line" />
          </div>
        ))}
      </div>
      <div className="sho-detail-loading-table">
        {[0, 1, 2, 3].map((row) => (
          <div className="sho-detail-loading-row" key={row}>
            <span className="sho-loading-line" />
            <span className="sho-loading-line" />
            <span className="sho-loading-line" />
            <span className="sho-loading-line" />
            <span className="sho-loading-line" />
            <span className="sho-loading-line" />
          </div>
        ))}
      </div>
    </div>
  </section>
)
