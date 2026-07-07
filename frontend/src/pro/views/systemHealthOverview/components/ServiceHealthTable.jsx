import React from 'react'
import PropTypes from 'prop-types'
import { Input, Pagination, Select, Space, Tooltip } from 'antd'
import { Search } from 'lucide-react'

import { HealthRing, HealthTable, Lamp } from './HealthStatus.jsx'

export default function ServiceHealthTable({
  HealthScoreColumnTitle,
  filteredServices,
  metricLabels,
  onPageChange,
  onServiceHealthFilterChange,
  onServiceKeywordChange,
  scoreThresholds,
  serviceEmptyMessage,
  serviceHealthFilter,
  serviceHealthFilterOptions,
  serviceKeyword,
  servicePage,
  servicePageSize,
  servicePageTotal,
  serviceRowsError,
  serviceRowsLoading,
}) {
  return (
    <section className="sho-panel">
      <div className="sho-panel-head">
        <div className="sho-panel-title">服务健康度</div>
        <Space wrap size={8}>
          <Select
            className="sho-health-filter"
            options={serviceHealthFilterOptions}
            placeholder="筛选服务"
            value={serviceHealthFilter}
            onChange={onServiceHealthFilterChange}
          />
          <Input
            allowClear
            className="sho-search"
            placeholder="搜索服务"
            prefix={<Search size={14} />}
            value={serviceKeyword}
            onChange={(event) => onServiceKeywordChange(event.target.value)}
          />
        </Space>
      </div>
      <HealthTable className="sho-service-table">
        <thead>
          <tr>
            <th>服务</th>
            <th>
              <HealthScoreColumnTitle scoreThresholds={scoreThresholds} />
            </th>
            {metricLabels.map(([, label]) => (
              <th key={label}>{label}</th>
            ))}
          </tr>
        </thead>
        <tbody>
          {serviceRowsLoading && (
            <tr>
              <td className="sho-empty-cell" colSpan={9}>
                正在加载服务
              </td>
            </tr>
          )}
          {!serviceRowsLoading && serviceRowsError && (
            <tr>
              <td className="sho-empty-cell" colSpan={9}>
                {serviceRowsError}
              </td>
            </tr>
          )}
          {!serviceRowsLoading &&
            !serviceRowsError &&
            filteredServices.map((service) => (
              <tr key={service.name}>
                <td>
                  <div className="sho-service-name">
                    <Tooltip title={service.name}>
                      <strong>{service.name}</strong>
                    </Tooltip>
                    {service.desc ? <span>{service.desc}</span> : null}
                  </div>
                </td>
                <td>
                  <HealthRing
                    score={service.score}
                    scoreThresholds={scoreThresholds}
                    status={service.status}
                  />
                </td>
                {metricLabels.map(([key]) => (
                  <td key={key}>
                    <Lamp status={service.metrics[key]} />
                  </td>
                ))}
              </tr>
            ))}
          {!serviceRowsLoading && !serviceRowsError && !filteredServices.length && (
            <tr>
              <td className="sho-empty-cell" colSpan={9}>
                {serviceEmptyMessage}
              </td>
            </tr>
          )}
        </tbody>
      </HealthTable>
      {servicePageTotal > servicePageSize && (
        <div className="sho-service-pagination">
          <Pagination
            current={servicePage}
            pageSize={servicePageSize}
            pageSizeOptions={['20', '50', '100']}
            showSizeChanger
            showTotal={(total) => `共 ${total} 个服务`}
            size="small"
            total={servicePageTotal}
            onChange={onPageChange}
          />
        </div>
      )}
    </section>
  )
}

ServiceHealthTable.propTypes = {
  HealthScoreColumnTitle: PropTypes.elementType.isRequired,
  filteredServices: PropTypes.arrayOf(PropTypes.object).isRequired,
  metricLabels: PropTypes.arrayOf(PropTypes.array).isRequired,
  onPageChange: PropTypes.func.isRequired,
  onServiceHealthFilterChange: PropTypes.func.isRequired,
  onServiceKeywordChange: PropTypes.func.isRequired,
  scoreThresholds: PropTypes.shape({
    severeAbnormalScoreThreshold: PropTypes.number,
    generalAbnormalScoreThreshold: PropTypes.number,
  }).isRequired,
  serviceEmptyMessage: PropTypes.string.isRequired,
  serviceHealthFilter: PropTypes.string,
  serviceHealthFilterOptions: PropTypes.arrayOf(PropTypes.object).isRequired,
  serviceKeyword: PropTypes.string.isRequired,
  servicePage: PropTypes.number.isRequired,
  servicePageSize: PropTypes.number.isRequired,
  servicePageTotal: PropTypes.number.isRequired,
  serviceRowsError: PropTypes.string.isRequired,
  serviceRowsLoading: PropTypes.bool.isRequired,
}
