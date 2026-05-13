import React from 'react'
import { Alert } from 'antd'
import {
  InfoCircleOutlined,
  AppstoreOutlined,
  ExclamationCircleOutlined,
  CheckCircleOutlined,
  LoadingOutlined,
  WarningOutlined,
} from '@ant-design/icons'
import { TrendingUp, Eye, EyeOff } from 'lucide-react'
import { useTranslation } from 'react-i18next'
import DelayLineChart from 'src/core/components/Chart/DelayLineChart'
import { convertTime } from 'src/core/utils/time'
import FindingsTree from '../FindingsTree/Tree'
import NodeDetail from './NodeDetail'
import StatusLegend from '../StatusLegend'
import { toFindingsNodeData, type FindingsNodeData } from '../FindingsTree/findingsContract'

// ==================== 类型定义 ====================
interface ServiceDetailProps {
  selectedService?: any
  loadingFindings?: boolean
  onRetryFindingsNode?: (nodeId: string) => void
  retryingFindingsNodeId?: string | null
}

interface ServiceIconProps {
  type?: string
  size?: 'small' | 'default' | 'large'
  showBackground?: boolean
}

interface MetricData {
  label: string
  type: string
  data: any[]
  value: number
  status: string
  hasData: boolean
}

// ==================== 工具函数 ====================
const convertYValue = (value: number, type: string): string => {
  const { t } = useTranslation('common')

  switch (type) {
    case 'logs':
      return Math.floor(value) + t('units.unit')

    case 'latency':
    case 'p90':
      if (value > 0 && value < 10) {
        return '< 0.01 ms'
      }
      return convertTime(value, 'ms', 2) + 'ms'

    case 'errorRate':
      if (value > 0 && value < 0.01) {
        return '< 0.01%'
      }
      return parseFloat(value.toFixed(2)) + '%'

    case 'tps':
      if (value > 0 && value < 0.01) {
        return '< 0.01' + t('units.timesPerMinute')
      }
      return parseFloat((Math.floor(value * 100) / 100).toString()) + t('units.timesPerMinute')

    default:
      return String(value)
  }
}

// ==================== 子组件 ====================
const ServiceIcon: React.FC<ServiceIconProps> = ({
  type,
  size = 'default',
  showBackground = true,
}) => {
  // 服务状态配置 - 与 index.tsx filterOptions 保持一致
  const statusConfig = {
    suspectedRootCauseNodesDetails: {
      icon: <ExclamationCircleOutlined />,
      color: 'var(--ant-color-error)',
      bgColor: 'var(--ant-color-error-bg)',
      background: `linear-gradient(135deg, var(--ant-color-error), var(--ant-color-error-hover))`,
    },
    anomaliesExcludedNodesDetails: {
      icon: <InfoCircleOutlined />,
      color: 'var(--ant-color-warning)',
      bgColor: 'var(--ant-color-warning-bg)',
      background: `linear-gradient(135deg, var(--ant-color-warning), var(--ant-color-warning-hover))`,
    },
    normalNodesDetails: {
      icon: <CheckCircleOutlined />,
      color: 'var(--ant-color-success)',
      bgColor: 'var(--ant-color-success-bg)',
      background: `linear-gradient(135deg, var(--ant-color-success), var(--ant-color-success-hover))`,
    },
    default: {
      icon: <AppstoreOutlined />,
      color: 'var(--ant-color-info)',
      bgColor: 'var(--ant-color-info-bg)',
      background: `linear-gradient(135deg, var(--ant-color-info), var(--ant-color-info-hover))`,
    },
  }

  const config = statusConfig[type as keyof typeof statusConfig] || statusConfig.default

  // 根据尺寸设置样式
  const sizeConfig = {
    small: { width: '24px', height: '24px', fontSize: '12px' },
    default: { width: '40px', height: '40px', fontSize: '16px' },
    large: { width: '48px', height: '48px', fontSize: '20px' },
  }

  const currentSize = sizeConfig[size]

  if (!showBackground) {
    return (
      <span style={{ color: config.color, fontSize: currentSize.fontSize }}>{config.icon}</span>
    )
  }

  return (
    <div
      className="rounded-xl flex items-center justify-center shadow-lg"
      style={{
        width: currentSize.width,
        height: currentSize.height,
        background: config.background,
        color: 'var(--ant-color-white)',
      }}
    >
      <span style={{ fontSize: currentSize.fontSize }}>{config.icon}</span>
    </div>
  )
}
// ==================== 子组件 - 空状态 ====================
const EmptyState: React.FC = () => {
  const { t } = useTranslation('pro/agui')

  return (
    <div className="flex items-center justify-center h-full">
      <div className="text-center">
        <div
          className="w-32 h-32 rounded-3xl flex items-center justify-center mx-auto mb-6 shadow-inner"
          style={{
            background: `linear-gradient(135deg, var(--ant-color-info-bg), var(--ant-color-primary-bg))`,
          }}
        >
          <div
            className="w-20 h-20 rounded-2xl flex items-center justify-center"
            style={{
              background: `linear-gradient(135deg, var(--ant-color-fill-secondary), var(--ant-color-fill-tertiary))`,
            }}
          >
            <div className="text-5xl opacity-60">📊</div>
          </div>
        </div>
        <div className="space-y-3">
          <h4 className="text-xl font-semibold text-[var(--ant-color-text)]">
            {t('evidencePanel.readyForAnalysis')}
          </h4>
          <p className="text-[var(--ant-color-text-secondary)] max-w-md mx-auto leading-relaxed">
            {t('evidencePanel.selectServicePrompt')}
          </p>
          <div className="flex items-center justify-center space-x-2 text-sm text-[var(--ant-color-text-tertiary)] mt-6">
            <div
              className="w-2 h-2 rounded-full animate-pulse"
              style={{ backgroundColor: 'var(--ant-color-info)' }}
            ></div>
            <span>{t('evidencePanel.monitoringActive')}</span>
          </div>
        </div>
      </div>
    </div>
  )
}

// ==================== 子组件 - 服务头部 ====================
const ServiceHeader: React.FC<{ selectedService: any }> = ({ selectedService }) => {
  const { t } = useTranslation('pro/agui')

  return (
    <div
      className="px-4 py-2 border-b border-[var(--ant-color-border)] backdrop-blur-sm"
      style={{
        backgroundColor: 'var(--ant-color-bg-container)',
      }}
    >
      <div className="flex items-center space-x-3">
        <ServiceIcon type={selectedService?.type} size="default" showBackground={true} />
        <div>
          <div className="text-xl font-bold text-[var(--ant-color-text)]">
            {selectedService ? selectedService.serviceName : t('evidencePanel.serviceDetails')}
          </div>
          <p className="text-sm text-[var(--ant-color-text-secondary)]">
            {selectedService
              ? t('evidencePanel.serviceMonitoring')
              : t('evidencePanel.comprehensiveInfo')}
          </p>
        </div>
      </div>
    </div>
  )
}

// ==================== 子组件 - 指标卡片 ====================
const MetricCard: React.FC<{ metric: MetricData }> = ({ metric }) => {
  const { t } = useTranslation('pro/agui')

  return (
    <div
      key={metric.type}
      className="border border-[var(--ant-color-border)] rounded-lg p-2 bg-[var(--ant-color-bg-container)] shadow-sm hover:shadow-md transition-shadow"
    >
      <div className=" px-2 text-base font-medium text-[var(--ant-color-text-secondary)] mb-3 flex justify-between">
        {metric.label}
        {metric.hasData ? (
          <div
            className="text-base font-bold text-[var(--ant-color-text-secondary)] flex gap-2 items-center"
            style={{
              color:
                metric.status === 'abnormal'
                  ? 'rgb(248 113 113)'
                  : metric.status === 'warning'
                    ? 'var(--ant-color-warning)'
                    : 'rgb(52, 211, 153)',
            }}
          >
            {convertYValue(metric.value, metric.type)}
            <div className="text-sm">
              {metric.status === 'abnormal' && <ExclamationCircleOutlined />}
              {metric.status === 'warning' && <WarningOutlined />}
              {metric.status === 'normal' && <CheckCircleOutlined />}
            </div>
          </div>
        ) : (
          <div className="text-sm text-[var(--ant-color-text-tertiary)]">
            {t('evidencePanel.noData')}
          </div>
        )}
      </div>
      <div className="h-28">
        {metric.hasData ? (
          <DelayLineChart
            data={metric.data}
            type={metric.type}
            timeRange={null}
            needFillZero={false}
            allowTimeBrush={false}
          />
        ) : (
          <div className="h-full flex items-center justify-center text-sm text-[var(--ant-color-text-tertiary)]">
            {t('evidencePanel.noData')}
          </div>
        )}
      </div>
    </div>
  )
}

// ==================== 子组件 - RED 指标区域 ====================
const RedMetricsSection: React.FC<{ redMetrics: any }> = ({ redMetrics }) => {
  const { t } = useTranslation('pro/agui')
  const [isExpanded, setIsExpanded] = React.useState(true)

  // 始终显示三个指标，即使为 null 也显示
  const metricsData: MetricData[] = [
    {
      label: t('evidencePanel.requestDuration'),
      type: 'latency',
      data: redMetrics?.latency?.chartData || [],
      value: redMetrics?.latency?.value || 0,
      status: redMetrics?.latency?.status || 'normal',
      hasData: redMetrics?.latency != null,
    },
    {
      label: t('evidencePanel.errorRate'),
      type: 'errorRate',
      data: redMetrics?.errorRate?.chartData || [],
      value: redMetrics?.errorRate?.value || 0,
      status: redMetrics?.errorRate?.status || 'normal',
      hasData: redMetrics?.errorRate != null,
    },
    {
      label: t('evidencePanel.requestRate'),
      type: 'tps',
      data: redMetrics?.tps?.chartData || [],
      value: redMetrics?.tps?.value || 0,
      status: redMetrics?.tps?.status || 'normal',
      hasData: redMetrics?.tps != null,
    },
  ]

  return (
    <div className="space-y-4">
      <div className="flex items-center space-x-2">
        <div
          className="w-1 h-5 rounded-full"
          style={{ backgroundColor: 'var(--ant-color-primary)' }}
        ></div>
        <div className="text-base font-semibold text-[var(--ant-color-text)]">
          {t('evidencePanel.requestMetrics')}
        </div>
        <button
          onClick={() => setIsExpanded(!isExpanded)}
          className="ml-2 p-1 hover:bg-[var(--ant-color-fill-secondary)] rounded transition-colors cursor-pointer"
          style={{ border: 'none', background: 'transparent' }}
        >
          {isExpanded ? (
            <Eye className="w-4 h-4 text-[var(--ant-color-text-secondary)]" />
          ) : (
            <EyeOff className="w-4 h-4 text-[var(--ant-color-text-secondary)]" />
          )}
        </button>
      </div>
      {isExpanded && (
        <div className="grid grid-cols-3 gap-4">
          {metricsData.map((metric) => (
            <MetricCard key={metric.type} metric={metric} />
          ))}
        </div>
      )}
    </div>
  )
}

// ==================== 子组件 - 推理树区域 ====================
const FindingsSection: React.FC<{
  evidence: any
  loadingFindings: boolean
  hydrationPhase?: string
  errorMessage?: string
  onNodeClick?: (nodeData: any) => void
}> = ({ evidence, loadingFindings, hydrationPhase, errorMessage, onNodeClick }) => {
  const { t } = useTranslation('pro/agui')
  const hasFindingsData = Boolean(evidence?.findingsData?.length)
  const isInitialLoading = loadingFindings && !hasFindingsData

  return (
    <div className="flex-1 h-0 flex flex-col">
      <div className="mb-2 flex items-center justify-between flex-0">
        <div className="flex items-center space-x-2">
          <TrendingUp className="w-5 h-5 text-cyan-400" />
          <div className="text-base font-semibold text-[var(--ant-color-text)]">
            {t('evidencePanel.reasoningTree')}
          </div>
          {loadingFindings && (
            <LoadingOutlined className="text-[var(--ant-color-primary)] animate-spin" />
          )}
        </div>
        {hasFindingsData && (
          <StatusLegend
            className="ml-4"
            showTitle={false}
            statuses={['normal', 'warning', 'abnormal']}
          />
        )}
      </div>

      {isInitialLoading ? (
        <div className="flex items-center justify-center py-8">
          <div className="text-center">
            <LoadingOutlined className="text-2xl text-[var(--ant-color-primary)] animate-spin mb-2" />
            <div className="text-sm text-[var(--ant-color-text-secondary)]">
              {t('evidencePanel.loadingFindings')}
            </div>
          </div>
        </div>
      ) : hasFindingsData ? (
        <div className="w-full flex flex-1 h-0">
          <FindingsTree
            nodes={evidence.findingsData}
            onNodeClick={onNodeClick}
            findingsId="service-findings"
          />
        </div>
      ) : hydrationPhase === 'failed' ? (
        <Alert
          type="warning"
          showIcon
          message={errorMessage || t('evidencePanel.noFindings')}
        />
      ) : (
        <div className="text-center text-[var(--ant-color-text-secondary)] py-8">
          <div className="text-3xl mb-2 opacity-50">📊</div>
          <div className="text-sm">{t('evidencePanel.noFindings')}</div>
        </div>
      )}
    </div>
  )
}

// ==================== 主组件 ====================
const ServiceDetail: React.FC<ServiceDetailProps> = ({
  selectedService,
  loadingFindings = false,
  onRetryFindingsNode,
  retryingFindingsNodeId,
}) => {
  const { evidence, redMetrics } = selectedService || {}
  const hydrationPhase = selectedService?.findingsHydrationPhase
  const findingsErrorMessage = selectedService?.findingsErrorMessage

  const [selectedNode, setSelectedNode] = React.useState<FindingsNodeData | null>(null)

  const handleNodeClick = React.useCallback((nodeData: FindingsNodeData) => {
    setSelectedNode(nodeData)
  }, [])

  const handleClearSelection = React.useCallback(() => {
    setSelectedNode(null)
  }, [])

  React.useEffect(() => {
    setSelectedNode(null)
  }, [selectedService?.serviceName])

  React.useEffect(() => {
    setSelectedNode((current) => {
      if (!current) {
        return current
      }
      const latestNode = evidence?.findingsData?.find((node: any) => node.nodeId === current.nodeId)
      return latestNode ? toFindingsNodeData(latestNode) : null
    })
  }, [evidence?.findingsData])

  return (
    <div
      className="h-full flex flex-col"
      style={{
        background: `linear-gradient(135deg, var(--ant-color-bg-container), var(--ant-color-fill-quaternary))`,
      }}
    >
      {/* 服务头部 */}
      <ServiceHeader selectedService={selectedService} />

      {/* 主内容区域 */}
      <div className="flex-1 p-2 overflow-y-hidden">
        {!selectedService ? (
          <EmptyState />
        ) : (
          <div className="flex gap-2 h-full">
            <div className="flex-1 space-y-3 flex flex-col">
              {/* RED Metrics 区域 */}
              <RedMetricsSection redMetrics={redMetrics} />

              <FindingsSection
                evidence={evidence}
                loadingFindings={loadingFindings}
                hydrationPhase={hydrationPhase}
                errorMessage={findingsErrorMessage}
                onNodeClick={handleNodeClick}
              />

              {/* Normal 服务的提示信息 */}
              {/* {selectedService?.type === 'normalNodesDetails' && <NormalServiceState />} */}
            </div>
            <div className="w-[300px] shrink-0">
              <NodeDetail
                selectedNode={selectedNode}
                onClose={handleClearSelection}
                onRetry={onRetryFindingsNode}
                retryingNodeId={retryingFindingsNodeId}
              />
            </div>
          </div>
        )}
      </div>
    </div>
  )
}
export default ServiceDetail
