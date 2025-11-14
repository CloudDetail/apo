import React from 'react'
import { Card, Empty, theme, Tooltip } from 'antd'
import {
  ExclamationCircleFilled,
  CheckCircleFilled,
  InfoCircleOutlined,
  ApiOutlined,
} from '@ant-design/icons'
import { useTranslation } from 'react-i18next'
import { FindingsNodeData } from '../FindingsTree/FindingsNode'
import StatusIcon, { useStatusLabel } from '../StatusIcon'
import NodeLogs from './NodeLogs'
import NodeTraces from './NodeTraces'
import NodeMetrics from './NodeMetrics'
import NodeTable from './NodeTable'

const { useToken } = theme

interface NodeDetailProps {
  selectedNode: FindingsNodeData | null
  onClose?: () => void // Reserved for future use
}
const NodeDetail: React.FC<NodeDetailProps> = ({ selectedNode }) => {
  const { token } = useToken()
  const { t } = useTranslation('pro/agui')
  const getStatusLabel = useStatusLabel()
  console.log('selectedNode', selectedNode)
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
          {/* 状态图标 */}
          {selectedNode.status && (
            <Tooltip title={getStatusLabel(selectedNode.status)} zIndex={100000}>
              <div
                style={{
                  color:
                    selectedNode.status === 'abnormal'
                      ? token.colorError
                      : selectedNode.status === 'warning'
                        ? token.colorWarning
                        : selectedNode.status === 'normal'
                          ? token.colorSuccess
                          : token.colorInfo,
                }}
              >
                <StatusIcon status={selectedNode.status} withContainer />
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
        {/* 徽标信息 */}
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

        {/* 主要内容 */}
        {selectedNode.summaryConclusion && (
          <div className="space-y-2">
            <div
              className="p-3 rounded"
              style={{
                backgroundColor:
                  selectedNode.status === 'abnormal'
                    ? token.colorErrorBg
                    : selectedNode.status === 'warning'
                      ? token.colorWarningBg
                      : selectedNode.status === 'normal'
                        ? token.colorSuccessBg
                        : token.colorInfoBg,
                border: `1px solid ${selectedNode.status === 'abnormal' ? token.colorErrorBorder : selectedNode.status === 'warning' ? token.colorWarningBorder : selectedNode.status === 'normal' ? token.colorSuccessBorder : token.colorInfoBorder}`,
                borderLeft: `5px solid ${selectedNode.status === 'abnormal' ? token.colorError : selectedNode.status === 'warning' ? token.colorWarning : selectedNode.status === 'normal' ? token.colorSuccess : token.colorInfo}`,
                borderRadius: token.borderRadiusLG,
              }}
            >
              {selectedNode.summaryConclusion}
              {/* <MarkdownRender content={selectedNode.summaryConclusion} theme="light" /> */}
            </div>
          </div>
        )}

        {/* 主要内容 */}
        {selectedNode.content && (
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

        {/* Abnormal Metrics */}
        {selectedNode.metrics?.abnormalMetrics.length > 0 && (
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

        {/* Normal Metrics */}
        {selectedNode.metrics?.normalMetrics.length > 0 && (
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

        {/* Table */}
        {selectedNode.table && <NodeTable table={selectedNode.table} />}

        {/* Logs */}
        {selectedNode.logs && selectedNode.logs.length > 0 && <NodeLogs logs={selectedNode.logs} />}
        {/* Traces */}
        {selectedNode.traces && selectedNode.traces.length > 0 && (
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
