import React from 'react'
import { Alert, Button, Card, Empty, theme, Tooltip } from 'antd'
import {
  ExclamationCircleFilled,
  CheckCircleFilled,
  InfoCircleOutlined,
  ApiOutlined,
  LoadingOutlined,
} from '@ant-design/icons'
import { useTranslation } from 'react-i18next'
import { RefreshCw } from 'lucide-react'
import StatusIcon, { useStatusLabel } from '../StatusIcon'
import NodeLogs from './NodeLogs'
import NodeTraces from './NodeTraces'
import NodeMetrics from './NodeMetrics'
import NodeTable from './NodeTable'
import RelatedReports from './RelatedReports'
import { getFindingsDetailSections, type FindingsNodeData } from '../FindingsTree/findingsContract'

const { useToken } = theme

interface NodeDetailProps {
  selectedNode: FindingsNodeData | null
  onClose?: () => void // Reserved for future use
  onRetry?: (nodeId: string) => void
  retryingNodeId?: string | null
}
const NodeDetail: React.FC<NodeDetailProps> = ({ selectedNode, onRetry, retryingNodeId }) => {
  const { token } = useToken()
  const { t } = useTranslation('pro/agui')
  const getStatusLabel = useStatusLabel()

  if (!selectedNode) {
    return (
      <div
        className="w-full h-full flex items-center justify-center p-4"
        style={{
          backgroundColor: token.colorBgContainer,
          borderRadius: token.borderRadiusLG,
          border: `1px solid ${token.colorBorder}`,
        }}
      >
        <Empty
          image={Empty.PRESENTED_IMAGE_SIMPLE}
          description={
            <div className="space-y-2">
              <div className="text-[var(--ant-color-text-secondary)]">
                {t('evidencePanel.noNodeSelected')}
              </div>
              <div className="text-xs text-[var(--ant-color-text-tertiary)]">
                {t('evidencePanel.clickNodeToView')}
              </div>
            </div>
          }
        />
      </div>
    )
  }

  const detailSections = getFindingsDetailSections(selectedNode)
  const loadingNodeDetailsText = t('evidencePanel.loadingNodeDetails')
  const visualStatus =
    selectedNode.visualStatus ??
    (selectedNode.nodeKind === 'unsupported' ? 'normal' : selectedNode.status)
  const statusColor =
    visualStatus === 'abnormal'
      ? token.colorError
      : visualStatus === 'warning'
        ? token.colorWarning
        : token.colorSuccess
  const statusBg =
    visualStatus === 'abnormal'
      ? token.colorErrorBg
      : visualStatus === 'warning'
        ? token.colorWarningBg
        : token.colorSuccessBg
  const statusBorder =
    visualStatus === 'abnormal'
      ? token.colorErrorBorder
      : visualStatus === 'warning'
        ? token.colorWarningBorder
        : token.colorSuccessBorder
  const isRetrying =
    retryingNodeId === selectedNode.nodeId || selectedNode.hydrationStatus === 'running'
  const canRetry =
    selectedNode.hydrationStatus === 'failed' && selectedNode.retryable && Boolean(onRetry)

  return (
    <Card
      className="w-full h-full"
      style={{
        borderRadius: token.borderRadiusLG,
        boxShadow: token.boxShadowTertiary,
      }}
      styles={{
        body: {
          padding: 0,
          height: '100%',
          display: 'flex',
          flexDirection: 'column',
        },
      }}
    >
      {/* 头部 */}
      <div
        className="flex items-center justify-between p-2 border-b"
        style={{
          borderColor: token.colorBorder,
          backgroundColor: token.colorBgContainer,
        }}
      >
        <div className="flex items-center gap-2 flex-1 min-w-0">
          {detailSections.loading ? (
            <Tooltip title={loadingNodeDetailsText} zIndex={100000}>
              <div
                className="inline-flex items-center justify-center shrink-0 w-7 h-7 rounded-xl"
                style={{
                  backgroundColor: token.colorPrimaryBg,
                  color: token.colorPrimary,
                }}
              >
                <LoadingOutlined className="animate-spin" />
              </div>
            </Tooltip>
          ) : (
            <Tooltip title={getStatusLabel(visualStatus)} zIndex={100000}>
              <div style={{ color: statusColor }}>
                <StatusIcon status={visualStatus} withContainer />
              </div>
            </Tooltip>
          )}

          <div className="flex-1 min-w-0">
            <div
              className="text-base font-semibold truncate"
              style={{ color: token.colorText }}
              title={selectedNode.title || selectedNode.content}
            >
              {selectedNode.title || t('evidencePanel.nodeDetails')}
            </div>
            {selectedNode.subtitle && (
              <div
                className="text-xs mt-1 truncate"
                style={{ color: token.colorTextSecondary }}
                title={selectedNode.subtitle}
              >
                {selectedNode.subtitle}
              </div>
            )}
          </div>
        </div>
        {/* {onClose && <Button type="text" icon={<CloseOutlined />} onClick={onClose} size="small" />} */}
      </div>

      {/* 内容区域 */}
      <div className="flex-1 overflow-y-auto p-4 space-y-4">
        {detailSections.loading && <Alert type="info" showIcon message={loadingNodeDetailsText} />}

        {selectedNode.hydrationStatus === 'failed' && (
          <Alert
            type="warning"
            showIcon
            message={selectedNode.errorMessage || t('evidencePanel.loadingFindings')}
            description={[
              selectedNode.errorType ? `type: ${selectedNode.errorType}` : null,
              typeof selectedNode.attempt === 'number' ? `attempt: ${selectedNode.attempt}` : null,
              selectedNode.retryable ? 'retryable' : null,
            ]
              .filter(Boolean)
              .join(' · ')}
            action={
              canRetry ? (
                <Button
                  size="small"
                  type="primary"
                  loading={isRetrying}
                  icon={isRetrying ? undefined : <RefreshCw size={14} />}
                  onClick={() => onRetry?.(selectedNode.nodeId)}
                >
                  {isRetrying ? t('evidencePanel.retryingNode') : t('evidencePanel.retryNode')}
                </Button>
              ) : undefined
            }
          />
        )}

        {selectedNode.badge && (
          <div className="flex items-center gap-2">
            <span className="text-xs text-[var(--ant-color-text-secondary)]">
              {t('evidencePanel.step')}
            </span>
            <span
              className="inline-flex items-center px-2 py-1 rounded text-xs font-medium"
              style={{
                backgroundColor: 'var(--ant-color-primary-bg)',
                color: 'var(--ant-color-primary)',
              }}
            >
              {selectedNode.badge}
            </span>
          </div>
        )}

        {detailSections.summaryConclusion && (
          <div className="space-y-2">
            <div
              className="p-3 rounded"
              style={{
                backgroundColor: statusBg,
                border: `1px solid ${statusBorder}`,
                borderLeft: `5px solid ${statusColor}`,
                borderRadius: token.borderRadiusLG,
              }}
            >
              {selectedNode.summaryConclusion}
            </div>
          </div>
        )}

        {detailSections.content && (
          <div className="space-y-2">
            <div className="flex items-center gap-2">
              <InfoCircleOutlined
                style={{
                  fontSize: 14,
                  color: token.colorTextSecondary,
                }}
              />
              <span className="text-xs font-medium text-[var(--ant-color-text-secondary)]">
                {t('evidencePanel.details')}
              </span>
            </div>
            <div
              className="p-3 rounded-lg text-sm leading-relaxed"
              style={{
                backgroundColor: token.colorFillQuaternary,
                border: `1px solid ${token.colorBorder}`,
                color: token.colorText,
              }}
            >
              {selectedNode.content}
            </div>
          </div>
        )}

        {!detailSections.loading && <RelatedReports reportLinks={selectedNode.reportLinks} />}
        {detailSections.abnormalMetrics && (
          <div className="space-y-3">
            <div className="space-y-2">
              <div className="flex items-center gap-2">
                <ExclamationCircleFilled
                  style={{
                    fontSize: 16,
                    color: token.colorError,
                  }}
                />
                <span
                  className="text-base font-medium"
                  style={{
                    color: token.colorTextBase,
                  }}
                >
                  {t('evidencePanel.abnormalMetrics')}
                </span>
              </div>
              <div
                style={{
                  height: 1,
                  backgroundColor: token.colorBorder,
                }}
              />
            </div>
            <div className="">
              <NodeMetrics metrics={selectedNode.metrics.abnormalMetrics} type="abnormal" />
            </div>
          </div>
        )}

        {detailSections.normalMetrics && (
          <div className="space-y-3">
            <div className="space-y-2">
              <div className="flex items-center gap-2">
                <CheckCircleFilled
                  style={{
                    fontSize: 16,
                    color: token.colorSuccess,
                  }}
                />
                <span
                  className="text-base font-medium"
                  style={{
                    color: token.colorTextBase,
                  }}
                >
                  {t('evidencePanel.normalMetrics')}
                </span>
              </div>
              <div
                style={{
                  height: 1,
                  backgroundColor: token.colorBorder,
                }}
              />
            </div>
            <div className="">
              <NodeMetrics metrics={selectedNode.metrics.normalMetrics} type="normal" />
            </div>
          </div>
        )}

        {detailSections.table && selectedNode.table && <NodeTable table={selectedNode.table} />}

        {detailSections.logs && selectedNode.logs && <NodeLogs logs={selectedNode.logs} />}
        {detailSections.traces && selectedNode.traces && (
          <div className="space-y-3">
            <div className="space-y-2">
              <div className="flex items-center gap-2">
                <ApiOutlined
                  style={{
                    fontSize: 16,
                    color: token.colorTextSecondary,
                  }}
                />
                <span
                  className="text-base font-medium"
                  style={{
                    color: token.colorTextBase,
                  }}
                >
                  {t('evidencePanel.traces')}
                </span>
              </div>
              <div
                style={{
                  height: 1,
                  backgroundColor: token.colorBorder,
                }}
              />
            </div>
            <div className="">
              <NodeTraces traces={selectedNode.traces} />
            </div>
          </div>
        )}
      </div>
    </Card>
  )
}

export default NodeDetail
