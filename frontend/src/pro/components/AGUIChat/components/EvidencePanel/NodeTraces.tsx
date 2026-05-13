import React from 'react'
import { theme, Tooltip } from 'antd'
import {
  ExclamationCircleFilled,
  ArrowRightOutlined,
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

interface TraceMetricProps {
  label: string
  value: string | number
  title: string
  danger?: boolean
}

const formatDuration = (value: number) => `${value.toFixed(2)} ms`

const NodeTraces: React.FC<NodeTracesProps> = ({ traces }) => {
  const { token } = useToken()
  const { t } = useTranslation('pro/agui')

  const getTypeMeta = (type: string) => {
    const lowerType = type?.toLowerCase() || ''

    if (lowerType.includes('mysql') || lowerType.includes('database') || lowerType.includes('db')) {
      return {
        icon: <DatabaseOutlined style={{ fontSize: 12 }} />,
        label: 'DB',
        color: token.colorInfo,
      }
    }

    if (lowerType.includes('http') || lowerType.includes('api')) {
      return {
        icon: <ApiOutlined style={{ fontSize: 12 }} />,
        label: lowerType.includes('http') ? 'HTTP' : 'API',
        color: token.colorSuccess,
      }
    }

    return {
      icon: <CloudOutlined style={{ fontSize: 12 }} />,
      label: type ? type.toUpperCase() : 'TRACE',
      color: token.colorTextSecondary,
    }
  }

  const TraceMetric: React.FC<TraceMetricProps> = ({ label, value, title, danger }) => (
    <Tooltip title={title}>
      <div className="min-w-0">
        <div
          className="text-[11px] leading-4 whitespace-nowrap"
          style={{ color: token.colorTextTertiary }}
        >
          {label}
        </div>
        <div
          className="text-xs font-semibold leading-5 truncate"
          style={{ color: danger ? token.colorError : token.colorText }}
        >
          {value}
        </div>
      </div>
    </Tooltip>
  )

  return (
    <div className="flex flex-col gap-4">
      {traces?.map((trace, traceIndex) => (
        <div key={traceIndex} className="space-y-2">
          <div className="flex items-center justify-between gap-2 px-1">
            <Tooltip title={trace.service} placement="topLeft">
              <div className="min-w-0 flex items-center gap-2">
                <span
                  className="flex h-6 w-6 shrink-0 items-center justify-center rounded-md"
                  style={{
                    color: token.colorPrimary,
                    backgroundColor: token.colorPrimaryBg,
                  }}
                >
                  <ApiOutlined style={{ fontSize: 13 }} />
                </span>
                <span className="truncate text-sm font-semibold" style={{ color: token.colorText }}>
                  {trace.service}
                </span>
              </div>
            </Tooltip>
            <span
              className="shrink-0 rounded px-1.5 py-0.5 text-[11px] leading-4"
              style={{
                color: token.colorTextSecondary,
                backgroundColor: token.colorFillQuaternary,
              }}
            >
              {trace.calls?.length ?? 0}
            </span>
          </div>

          <div className="space-y-2">
            {trace.calls?.map((call, callIndex) => {
              const hasError = call.errorCount > 0
              const typeMeta = getTypeMeta(call.type)
              const operations = call.operations ?? []
              const showOperations =
                operations.length > 1 ||
                (operations.length === 1 && operations[0]?.name !== call.target)

              return (
                <div
                  key={callIndex}
                  className="rounded-lg border px-2.5 py-2.5"
                  style={{
                    borderColor: hasError ? token.colorErrorBorder : token.colorBorderSecondary,
                    backgroundColor: hasError ? token.colorErrorBg : token.colorFillQuaternary,
                  }}
                >
                  <div className="flex items-start gap-2">
                    <span
                      className="mt-0.5 flex h-5 w-5 shrink-0 items-center justify-center rounded"
                      style={{
                        color: hasError ? token.colorError : token.colorPrimary,
                        backgroundColor: token.colorBgContainer,
                      }}
                    >
                      <ArrowRightOutlined style={{ fontSize: 12 }} />
                    </span>

                    <div className="min-w-0 flex-1">
                      <div className="flex min-w-0 items-center gap-1.5">
                        <span
                          className="inline-flex shrink-0 items-center gap-1 rounded px-1.5 py-0.5 text-[11px] font-medium leading-4"
                          style={{
                            color: typeMeta.color,
                            backgroundColor: token.colorBgContainer,
                            border: `1px solid ${hasError ? token.colorErrorBorder : token.colorBorder}`,
                          }}
                        >
                          {typeMeta.icon}
                          <span>{typeMeta.label}</span>
                        </span>

                        {hasError && (
                          <span
                            className="inline-flex shrink-0 items-center gap-1 rounded px-1.5 py-0.5 text-[11px] font-medium leading-4"
                            style={{
                              color: token.colorError,
                              backgroundColor: token.colorErrorBg,
                              border: `1px solid ${token.colorErrorBorder}`,
                            }}
                          >
                            <ExclamationCircleFilled style={{ fontSize: 11 }} />
                            <span>
                              {call.errorCount} {t('evidencePanel.errorSuffix')}
                            </span>
                          </span>
                        )}
                      </div>

                      <Tooltip title={call.target} placement="topLeft">
                        <div
                          className="mt-1 text-sm font-semibold leading-snug"
                          style={{
                            color: token.colorText,
                            overflowWrap: 'anywhere',
                          }}
                        >
                          {call.target}
                        </div>
                      </Tooltip>
                    </div>
                  </div>

                  <div className="mt-2.5 grid grid-cols-2 gap-x-4 gap-y-1.5">
                    <TraceMetric
                      label={t('evidencePanel.calls')}
                      title={t('evidencePanel.callCount')}
                      value={call.count.toLocaleString()}
                    />
                    <TraceMetric
                      label={t('evidencePanel.average')}
                      title={t('evidencePanel.avgDuration')}
                      value={formatDuration(call.avgDuration)}
                    />
                    <TraceMetric
                      label={t('evidencePanel.max')}
                      title={t('evidencePanel.maxDuration')}
                      value={formatDuration(call.maxDuration)}
                    />
                    <TraceMetric
                      label={t('evidencePanel.errors')}
                      title={t('evidencePanel.errorCount')}
                      value={call.errorCount}
                      danger={hasError}
                    />
                  </div>

                  {showOperations && (
                    <div
                      className="mt-2 border-t pt-2"
                      style={{ borderColor: hasError ? token.colorErrorBorder : token.colorBorder }}
                    >
                      <div
                        className="mb-1.5 text-[11px] font-medium leading-4"
                        style={{ color: token.colorTextSecondary }}
                      >
                        {t('evidencePanel.operationDetails')} ({operations.length})
                      </div>

                      <div className="space-y-1">
                        {operations.map((op, opIndex) => (
                          <div key={opIndex} className="min-w-0">
                            <Tooltip title={op.name} placement="topLeft">
                              <div
                                className="truncate text-xs font-mono leading-5"
                                style={{ color: token.colorText }}
                              >
                                {op.name}
                              </div>
                            </Tooltip>
                            <div
                              className="flex flex-wrap gap-x-3 gap-y-0.5 text-[11px] leading-4"
                              style={{ color: token.colorTextTertiary }}
                            >
                              <span className="whitespace-nowrap">
                                {t('evidencePanel.times')}: {op.count.toLocaleString()}
                              </span>
                              <span className="whitespace-nowrap">
                                {t('evidencePanel.average')}: {formatDuration(op.avgDuration)}
                              </span>
                            </div>
                          </div>
                        ))}
                      </div>
                    </div>
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
