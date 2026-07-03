import React from 'react'
import PropTypes from 'prop-types'
import { Input, Pagination, Select, Space, Tooltip } from 'antd'
import { HelpCircle, Search } from 'lucide-react'

import { HealthRing, HealthTable, Lamp, Legend } from './HealthStatus.jsx'
import SystemHealthScoreTrendCell from './SystemHealthScoreTrendCell.jsx'

export default function OverviewSystemTable({
  HealthScoreCalculationTooltip,
  HealthScoreColumnTitle,
  metricLabels,
  onOpenDetail,
  onPageChange,
  onSystemHealthFilterChange,
  onSystemKeywordChange,
  overviewState,
  overviewSystems,
  scoreThresholds,
  systemHealthFilter,
  systemHealthFilterOptions,
  systemKeyword,
  systemPage,
  systemPageSize,
  systemPageTotal,
  systemTrendError,
  systemTrendGranularity,
  systemTrendLoading,
  systemTrendRangeLabel,
  systemTrendRows,
}) {
  return (
    <section className="sho-panel">
      <div className="sho-panel-head">
        <div className="sho-panel-title">
          系统健康度
          <Tooltip
            title={<HealthScoreCalculationTooltip />}
            overlayClassName="sho-health-score-popover"
          >
            <HelpCircle size={15} className="sho-help-icon" />
          </Tooltip>
        </div>
        <Space wrap size={8}>
          <Select
            allowClear
            className="sho-health-filter"
            options={systemHealthFilterOptions}
            placeholder="筛选系统"
            value={systemHealthFilter}
            onChange={onSystemHealthFilterChange}
          />
          <Input
            allowClear
            className="sho-search"
            placeholder="搜索系统"
            prefix={<Search size={14} />}
            value={systemKeyword}
            onChange={(event) => onSystemKeywordChange(event.target.value)}
          />
          <Legend />
        </Space>
      </div>
      {overviewState || (
        <HealthTable className="sho-system-table">
          <thead>
            <tr>
              <th>系统</th>
              <th>
                <HealthScoreColumnTitle scoreThresholds={scoreThresholds} />
              </th>
              <th>历史值</th>
              {metricLabels.map(([, label]) => (
                <th key={label}>{label}</th>
              ))}
            </tr>
          </thead>
          <tbody>
            {overviewSystems.map((system) => (
              <tr key={system.groupId} onClick={() => onOpenDetail(system)}>
                <td>
                  <div className="sho-name-cell">
                    <button
                      type="button"
                      className="sho-system-name-button"
                      aria-label={`查看${system.name}健康度详情`}
                      onClick={(event) => {
                        event.stopPropagation()
                        onOpenDetail(system)
                      }}
                    >
                      <strong>{system.name}</strong>
                    </button>
                  </div>
                </td>
                <td>
                  <HealthRing
                    score={system.score}
                    scoreThresholds={scoreThresholds}
                    status={system.status}
                  />
                </td>
                <td className="sho-system-trend-cell-td">
                  <SystemHealthScoreTrendCell
                    error={systemTrendError}
                    granularity={systemTrendGranularity}
                    loading={systemTrendLoading}
                    rows={systemTrendRows[String(system.groupId)]}
                    scoreThresholds={scoreThresholds}
                    systemName={system.name}
                    trendRangeLabel={systemTrendRangeLabel}
                  />
                </td>
                {metricLabels.map(([key]) => (
                  <td key={key}>
                    <Lamp status={system.metrics[key]} />
                  </td>
                ))}
              </tr>
            ))}
          </tbody>
        </HealthTable>
      )}
      {systemPageTotal > systemPageSize && (
        <div className="sho-system-pagination">
          <Pagination
            current={systemPage}
            pageSize={systemPageSize}
            pageSizeOptions={['20', '50', '100']}
            showSizeChanger
            showTotal={(total) => `共 ${total} 个系统`}
            size="small"
            total={systemPageTotal}
            onChange={onPageChange}
          />
        </div>
      )}
    </section>
  )
}

OverviewSystemTable.propTypes = {
  HealthScoreCalculationTooltip: PropTypes.elementType.isRequired,
  HealthScoreColumnTitle: PropTypes.elementType.isRequired,
  metricLabels: PropTypes.arrayOf(PropTypes.array).isRequired,
  onOpenDetail: PropTypes.func.isRequired,
  onPageChange: PropTypes.func.isRequired,
  onSystemHealthFilterChange: PropTypes.func.isRequired,
  onSystemKeywordChange: PropTypes.func.isRequired,
  overviewState: PropTypes.node,
  overviewSystems: PropTypes.arrayOf(PropTypes.object).isRequired,
  scoreThresholds: PropTypes.shape({
    severeAbnormalScoreThreshold: PropTypes.number,
    generalAbnormalScoreThreshold: PropTypes.number,
  }).isRequired,
  systemHealthFilter: PropTypes.string,
  systemHealthFilterOptions: PropTypes.arrayOf(PropTypes.object).isRequired,
  systemKeyword: PropTypes.string.isRequired,
  systemPage: PropTypes.number.isRequired,
  systemPageSize: PropTypes.number.isRequired,
  systemPageTotal: PropTypes.number.isRequired,
  systemTrendError: PropTypes.string.isRequired,
  systemTrendGranularity: PropTypes.string.isRequired,
  systemTrendLoading: PropTypes.bool.isRequired,
  systemTrendRangeLabel: PropTypes.string.isRequired,
  systemTrendRows: PropTypes.objectOf(PropTypes.array).isRequired,
}
