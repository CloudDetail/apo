import React from 'react'
import { theme, Tooltip, Collapse } from 'antd'
import {
  ExclamationCircleFilled,
  WarningOutlined,
  CloseCircleOutlined,
  ArrowRightOutlined,
  ThunderboltOutlined,
  ClockCircleOutlined,
  ApiOutlined,
  DatabaseOutlined,
  CloudOutlined,
} from '@ant-design/icons'
import { useTranslation } from 'react-i18next'

const { useToken } = theme

export interface TraceCall {
  target: string
  type: string
  count: number
  avgDuration: number
  maxDuration: number
  errorCount: number
  operations?: Array<{
    name: string
    count: number
    avgDuration: number
  }>
}

export interface TraceData {
  service: string
  calls: TraceCall[]
}

interface NodeTracesProps {
  traces: TraceData[]
}

const NodeTraces: React.FC<NodeTracesProps> = ({ traces }) => {
  const { token } = useToken()
  const { t } = useTranslation('pro/agui')

  const getTypeIcon = (type: string) => {
    const lowerType = type?.toLowerCase() || ''
    if (lowerType.includes('mysql') || lowerType.includes('database') || lowerType.includes('db')) {
      return <DatabaseOutlined style={{ fontSize: 14 }} />
    }
    if (lowerType.includes('http') || lowerType.includes('api')) {
      return <ApiOutlined style={{ fontSize: 14 }} />
    }
    return <CloudOutlined style={{ fontSize: 14 }} />
  }

  const getTypeColor = (type: string) => {
    const lowerType = type?.toLowerCase() || ''
    if (lowerType.includes('mysql') || lowerType.includes('database') || lowerType.includes('db')) {
      return token.colorInfo
    }
    if (lowerType.includes('http') || lowerType.includes('api')) {
      return token.colorSuccess
    }
    return token.colorTextSecondary
  }

  return (
    <div className="flex flex-col gap-4">
      {traces?.map((trace, traceIndex) => (
        <div key={traceIndex} className="space-y-3">
          {/* 服务名称 */}
          <div
            className="inline-flex items-center gap-2 px-3 py-2 rounded-lg text-sm font-semibold"
            style={{
              backgroundColor: token.colorPrimaryBg,
              color: token.colorPrimary,
              border: `1px solid ${token.colorPrimaryBorder}`,
            }}
          >
            <ApiOutlined />
            <span>{trace.service}</span>
          </div>

          {/* 调用链路 */}
          <div className="space-y-3">
            {trace.calls?.map((call, callIndex) => {
              const hasError = call.errorCount > 0
              const typeColor = getTypeColor(call.type)

              return (
                <div
                  key={callIndex}
                  className="rounded-lg border overflow-hidden"
                  style={{
                    borderColor: hasError ? token.colorErrorBorder : token.colorBorder,
                    backgroundColor: token.colorBgContainer,
                  }}
                >
                  {/* 调用关系主体 */}
                  <div
                    className="p-3"
                    style={{
                      backgroundColor: hasError ? token.colorErrorBg : token.colorFillQuaternary,
                    }}
                  >
                    <div className="flex items-center gap-3">
                      {/* 箭头指向目标 */}
                      <ArrowRightOutlined
                        style={{
                          fontSize: 16,
                          color: hasError ? token.colorError : token.colorPrimary,
                        }}
                      />

                      {/* 目标服务 */}
                      <div className="flex items-center gap-2 flex-1">
                        <div
                          className="flex items-center gap-1.5 px-2 py-1 rounded text-xs font-medium"
                          style={{
                            backgroundColor: token.colorBgContainer,
                            border: `1px solid ${hasError ? token.colorErrorBorder : token.colorBorder}`,
                            color: typeColor,
                          }}
                        >
                          {getTypeIcon(call.type)}
                          <span>{call.type.toUpperCase()}</span>
                        </div>
                        <span className="text-sm font-medium" style={{ color: token.colorText }}>
                          {call.target}
                        </span>
                      </div>

                      {/* 错误标识 */}
                      {hasError && (
                        <div
                          className="flex items-center gap-1 px-2 py-1 rounded text-xs font-medium"
                          style={{
                            backgroundColor: token.colorErrorBg,
                            color: token.colorError,
                            border: `1px solid ${token.colorErrorBorder}`,
                          }}
                        >
                          <ExclamationCircleFilled />
                          <span>
                            {call.errorCount} {t('evidencePanel.errorSuffix')}
                          </span>
                        </div>
                      )}
                    </div>

                    {/* 指标信息 */}
                    <div className="flex items-center gap-4 mt-3">
                      {/* 调用次数 */}
                      <Tooltip title={t('evidencePanel.callCount')}>
                        <div className="flex items-center gap-1.5">
                          <ThunderboltOutlined
                            style={{ fontSize: 14, color: token.colorTextSecondary }}
                          />
                          <span className="text-xs" style={{ color: token.colorTextSecondary }}>
                            {t('evidencePanel.calls')}:
                          </span>
                          <span
                            className="text-sm font-semibold"
                            style={{ color: token.colorText }}
                          >
                            {call.count.toLocaleString()}
                          </span>
                        </div>
                      </Tooltip>

                      {/* 平均耗时 */}
                      <Tooltip title={t('evidencePanel.avgDuration')}>
                        <div className="flex items-center gap-1.5">
                          <ClockCircleOutlined
                            style={{ fontSize: 14, color: token.colorTextSecondary }}
                          />
                          <span className="text-xs" style={{ color: token.colorTextSecondary }}>
                            {t('evidencePanel.average')}:
                          </span>
                          <span
                            className="text-sm font-semibold"
                            style={{ color: token.colorText }}
                          >
                            {call.avgDuration.toFixed(2)} ms
                          </span>
                        </div>
                      </Tooltip>

                      {/* 最大耗时 */}
                      <Tooltip title={t('evidencePanel.maxDuration')}>
                        <div className="flex items-center gap-1.5">
                          <WarningOutlined
                            style={{ fontSize: 14, color: token.colorTextSecondary }}
                          />
                          <span className="text-xs" style={{ color: token.colorTextSecondary }}>
                            {t('evidencePanel.max')}:
                          </span>
                          <span
                            className="text-sm font-semibold"
                            style={{ color: token.colorText }}
                          >
                            {call.maxDuration.toFixed(2)} ms
                          </span>
                        </div>
                      </Tooltip>

                      {/* 错误次数 */}
                      <Tooltip title={t('evidencePanel.errorCount')}>
                        <div className="flex items-center gap-1.5">
                          <CloseCircleOutlined
                            style={{
                              fontSize: 14,
                              color: hasError ? token.colorError : token.colorTextSecondary,
                            }}
                          />
                          <span className="text-xs" style={{ color: token.colorTextSecondary }}>
                            {t('evidencePanel.errors')}:
                          </span>
                          <span
                            className="text-sm font-semibold"
                            style={{ color: hasError ? token.colorError : token.colorText }}
                          >
                            {call.errorCount}
                          </span>
                        </div>
                      </Tooltip>
                    </div>
                  </div>

                  {/* 操作详情（可折叠） */}
                  {call.operations && call.operations.length > 0 && (
                    <Collapse
                      ghost
                      size="small"
                      items={[
                        {
                          key: '1',
                          label: (
                            <span
                              className="text-xs font-medium"
                              style={{ color: token.colorTextSecondary }}
                            >
                              {t('evidencePanel.operationDetails')} ({call.operations.length})
                            </span>
                          ),
                          children: (
                            <div className="space-y-2">
                              {call.operations.map((op, opIndex) => (
                                <div
                                  key={opIndex}
                                  className="flex items-center justify-between p-2 rounded"
                                  style={{
                                    backgroundColor: token.colorFillQuaternary,
                                    border: `1px solid ${token.colorBorder}`,
                                  }}
                                >
                                  <div className="flex items-center gap-2 flex-1">
                                    <span
                                      className="text-xs font-mono"
                                      style={{ color: token.colorText }}
                                    >
                                      {op.name}
                                    </span>
                                  </div>
                                  <div className="flex items-center gap-4">
                                    <div className="flex items-center gap-1">
                                      <span
                                        className="text-xs"
                                        style={{ color: token.colorTextSecondary }}
                                      >
                                        {t('evidencePanel.times')}:
                                      </span>
                                      <span
                                        className="text-xs font-medium"
                                        style={{ color: token.colorText }}
                                      >
                                        {op.count}
                                      </span>
                                    </div>
                                    <div className="flex items-center gap-1">
                                      <span
                                        className="text-xs"
                                        style={{ color: token.colorTextSecondary }}
                                      >
                                        {t('evidencePanel.average')}:
                                      </span>
                                      <span
                                        className="text-xs font-medium"
                                        style={{ color: token.colorText }}
                                      >
                                        {op.avgDuration.toFixed(2)} ms
                                      </span>
                                    </div>
                                  </div>
                                </div>
                              ))}
                            </div>
                          ),
                        },
                      ]}
                    />
                  )}
                </div>
              )
            })}
          </div>
        </div>
      ))}
    </div>
  )
}

export default NodeTraces
