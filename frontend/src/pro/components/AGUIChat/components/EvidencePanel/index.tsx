import { Collapse, CollapseProps, Input, Badge, Card, Button, Alert, Tooltip } from 'antd'
import {
  SearchOutlined,
  AppstoreOutlined,
  CheckCircleOutlined,
  ExclamationCircleOutlined,
  InfoCircleOutlined,
  FullscreenOutlined,
  FullscreenExitOutlined,
} from '@ant-design/icons'
import { useState, useMemo, useEffect, useCallback, useRef } from 'react'
import { useTranslation } from 'react-i18next'
import ServiceDetail from './ServiceDetail'
import { getFindingsDataApi } from '../../../../api/conversation'
import type { FindingsNodePayload } from '../../../../api/conversation'
import { useAGUIState } from '../../../../contexts/AGUIContext'
// @ts-ignore

const ServiceListChildren = ({
  services,
  onServiceSelect,
  selectedService,
  type,
}: {
  services: any[]
  onServiceSelect: (service: any) => void
  selectedService: any
  type: string
}) => {
  const { t } = useTranslation('pro/agui')

  return (
    <div className="px-4 py-2 space-y-3">
      {services.length === 0 ? (
        <div className="text-center text-[var(--ant-color-text-secondary)] py-8">
          <div className="text-3xl mb-2 opacity-50">🔍</div>
          <div className="text-sm">{t('evidencePanel.noServicesFound')}</div>
        </div>
      ) : (
        services.map((service, index) => {
          const isSelected = selectedService && selectedService.serviceName === service.serviceName
          return (
            <Card
              key={service.key || index}
              size="small"
              hoverable
              onClick={() =>
                onServiceSelect({
                  ...service,
                  type,
                })
              }
              className={`group transition-all duration-300 hover:shadow-lg hover:scale-[1.02] border-2 shadow-sm cursor-pointer ${
                isSelected ? 'ring-2 ring-offset-2' : ''
              }`}
              style={{
                background: isSelected
                  ? `linear-gradient(to right, var(--ant-color-primary-bg), var(--ant-color-primary-hover-bg))`
                  : `linear-gradient(to right, var(--ant-color-bg-container), var(--ant-color-fill-quaternary))`,
                borderColor: isSelected ? 'var(--ant-color-primary)' : 'transparent',
              }}
            >
              <div className="flex items-start space-x-3">
                {/* <div
                  className="flex-shrink-0 w-8 h-8 rounded-lg flex items-center justify-center shadow-sm"
                  style={{
                    background: `linear-gradient(135deg, var(--ant-color-primary), var(--ant-color-primary-hover))`,
                    color: 'var(--ant-color-white)',
                  }}
                >
                  <AppstoreOutlined className="text-sm" />
                </div> */}
                <div className="flex-1 min-w-0">
                  <div
                    className={`text-sm font-semibold mb-1 transition-colors ${
                      isSelected
                        ? 'text-[var(--ant-color-primary)]'
                        : 'text-[var(--ant-color-text)] group-hover:text-[var(--ant-color-primary)]'
                    }`}
                  >
                    {service.serviceName}
                  </div>
                  <div className="text-xs text-[var(--ant-color-text-secondary)] leading-relaxed line-clamp-2">
                    {service.description}
                  </div>
                </div>
              </div>
            </Card>
          )
        })
      )}
    </div>
  )
}
const ServiceList = ({
  onServiceSelect,
  selectedService,
  evidenceData,
}: {
  onServiceSelect: (service: any) => void
  selectedService: any
  evidenceData: any
}) => {
  const { t } = useTranslation('pro/agui')
  const [activeFilters, setActiveFilters] = useState<string[]>([
    'suspectedRootCauseNodesDetails',
    'anomaliesExcludedNodesDetails',
    'normalNodesDetails',
  ])
  const [searchText, setSearchText] = useState('')

  const filterOptions = [
    {
      key: 'suspectedRootCauseNodesDetails',
      label: t('evidencePanel.suspected'),
      color: 'var(--ant-color-error)',
      bgColor: 'var(--ant-color-error-bg)',
      icon: <ExclamationCircleOutlined />,
      count: evidenceData?.suspectedRootCauseNodesDetails?.length || 0,
      tooltip: t('evidencePanel.suspectedTooltip'),
    },
    {
      key: 'anomaliesExcludedNodesDetails',
      label: t('evidencePanel.impacted'),
      color: 'var(--ant-color-warning)',
      bgColor: 'var(--ant-color-warning-bg)',
      icon: <InfoCircleOutlined />,
      count: evidenceData?.anomaliesExcludedNodesDetails?.length || 0,
      tooltip: t('evidencePanel.impactedTooltip'),
    },
    {
      key: 'normalNodesDetails',
      label: t('evidencePanel.normal'),
      color: 'var(--ant-color-success)',
      bgColor: 'var(--ant-color-success-bg)',
      icon: <CheckCircleOutlined />,
      count: evidenceData?.normalNodesDetails?.length || 0,
      tooltip: t('evidencePanel.normalTooltip'),
    },
  ]

  const toggleFilter = (filterKey: string) => {
    setActiveFilters((prev) =>
      prev.includes(filterKey) ? prev.filter((f) => f !== filterKey) : [...prev, filterKey],
    )
  }

  const filteredData = useMemo(() => {
    const filtered: any = {}

    filterOptions.forEach((option) => {
      if (activeFilters.includes(option.key)) {
        const services = evidenceData?.[option.key as keyof typeof evidenceData] || []
        filtered[option.key] = services.filter((service: any) =>
          service.serviceName?.toLowerCase().includes(searchText.toLowerCase()),
        )
      }
    })

    return filtered
  }, [activeFilters, searchText])

  const items: CollapseProps['items'] = filterOptions
    .filter((option) => activeFilters.includes(option.key))
    .map((option) => ({
      key: option.key,
      classNames: {
        header: '!items-center ',
      },
      label: (
        <div className="flex items-center justify-between py-1">
          <div className="flex items-center space-x-3">
            <div
              className="w-8 h-8 rounded-xl flex items-center justify-center shadow-sm"
              style={{ backgroundColor: option.bgColor }}
            >
              <span style={{ color: option.color }} className="text-sm">
                {option.icon}
              </span>
            </div>
            <div>
              <span className="font-semibold text-[var(--ant-color-text)]">
                {option.label} {t('evidencePanel.services')}
              </span>
              {/* <div className="text-xs text-[var(--ant-color-text-secondary)] mt-0.5">
                Active monitoring services
              </div> */}
            </div>
          </div>
          <Badge
            count={filteredData[option.key]?.length || 0}
            style={{ backgroundColor: option.color }}
            className="shadow-sm"
          />
        </div>
      ),
      children: (
        <ServiceListChildren
          services={filteredData[option.key] || []}
          onServiceSelect={onServiceSelect}
          selectedService={selectedService}
          type={option.key}
        />
      ),
    }))

  return (
    <div
      className="h-full flex flex-col"
      style={{
        background: `linear-gradient(to bottom, var(--ant-color-bg-container), var(--ant-color-fill-quaternary))`,
      }}
    >
      {/* Header */}
      <div
        className="px-4 py-2 border-b border-[var(--ant-color-border)] backdrop-blur-sm"
        style={{
          backgroundColor: 'var(--ant-color-bg-container)',
          opacity: 0.95,
        }}
      >
        <div className="flex items-center space-x-3 mb-4">
          <div
            className="w-10 h-10 rounded-xl flex items-center justify-center shadow-lg"
            style={{
              background: `linear-gradient(135deg, var(--ant-color-primary), var(--ant-color-primary-hover))`,
              color: 'var(--ant-color-white)',
            }}
          >
            <AppstoreOutlined className="text-lg" />
          </div>
          <div>
            <div
              className="text-xl font-bold bg-clip-text text-transparent"
              style={{
                background: `linear-gradient(to right, var(--ant-color-primary), var(--ant-color-primary-hover))`,
                WebkitBackgroundClip: 'text',
                backgroundClip: 'text',
              }}
            >
              {t('evidencePanel.serviceMonitor')}
            </div>
          </div>
        </div>

        {/* Search */}
        <div className="mb-6">
          <Input
            placeholder={t('evidencePanel.searchPlaceholder')}
            prefix={<SearchOutlined className="text-[var(--ant-color-text-secondary)]" />}
            value={searchText}
            onChange={(e) => setSearchText(e.target.value)}
            className="rounded-xl shadow-sm border-0"
            style={{
              backgroundColor: 'var(--ant-color-fill-quaternary)',
            }}
            size="large"
            allowClear
          />
        </div>

        {/* Filter Indicators */}
        <div className="flex flex-wrap gap-2">
          {filterOptions.map((option) => (
            <Tooltip key={option.key} title={option.tooltip} placement="top" zIndex={100000}>
              <button
                onClick={() => toggleFilter(option.key)}
                className={`w-full group flex items-center space-x-3 p-2 rounded-xl border-2 transition-all duration-300 shadow-sm hover:shadow-md hover:scale-105 ${
                  activeFilters.includes(option.key)
                    ? 'border-transparent shadow-lg transform scale-105'
                    : 'hover:border-[var(--ant-color-primary-hover)]'
                }`}
                style={{
                  backgroundColor: activeFilters.includes(option.key)
                    ? option.bgColor
                    : 'var(--ant-color-bg-container)',
                  borderColor: activeFilters.includes(option.key)
                    ? option.color
                    : 'var(--ant-color-border)',
                }}
              >
                <div
                  className="w-3 h-3 rounded-lg flex items-center justify-center shadow-sm"
                  style={{
                    backgroundColor: activeFilters.includes(option.key)
                      ? option.color
                      : option.bgColor,
                    color: activeFilters.includes(option.key)
                      ? 'rgba(255,255,255,0.7)'
                      : option.color,
                  }}
                >
                  {option.icon}
                </div>
                <div className="flex items-center justify-between w-full">
                  <span
                    className="text-sm font-semibold"
                    style={{
                      color: activeFilters.includes(option.key)
                        ? option.color
                        : 'var(--ant-color-text)',
                    }}
                  >
                    {option.label}
                  </span>
                  <span className="text-xs text-[var(--ant-color-text-secondary)]">
                    {option.count} {t('evidencePanel.servicesCount')}
                  </span>
                </div>
              </button>
            </Tooltip>
          ))}
        </div>
      </div>

      {/* Content */}
      <div className="flex-1 overflow-y-auto">
        {items.length === 0 ? (
          <div className="text-center text-[var(--ant-color-text-secondary)] py-16">
            <div
              className="w-20 h-20 rounded-full flex items-center justify-center mx-auto mb-4 shadow-inner"
              style={{
                background: `linear-gradient(135deg, var(--ant-color-fill-secondary), var(--ant-color-fill-tertiary))`,
              }}
            >
              <div className="text-4xl opacity-50">🔍</div>
            </div>
            <div className="text-lg font-medium mb-2">{t('evidencePanel.noServicesFound')}</div>
            <div className="text-sm opacity-75">{t('evidencePanel.tryAdjustFilters')}</div>
          </div>
        ) : (
          <div className="p-2">
            {items.map((item) => (
              <div key={item.key.toString()} className="mb-4">
                <Collapse
                  bordered={false}
                  items={[item]}
                  ghost
                  expandIconPosition="end"
                  defaultActiveKey={[item.key.toString()]}
                  className="backdrop-blur-sm rounded-xl shadow-sm border overflow-hidden"
                  style={{
                    backgroundColor: 'var(--ant-color-bg-container)',
                    borderColor: 'var(--ant-color-border)',
                    opacity: 0.9,
                  }}
                />
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  )
}

const EvidencePanel = ({ data }: { data: any }) => {
  const { t } = useTranslation('pro/agui')
  const [selectedService, setSelectedService] = useState<any>(null)
  const [isFullscreen, setIsFullscreen] = useState<boolean>(false) // 初始非全屏
  const [loadingFindings, setLoadingFindings] = useState<boolean>(false) // findings 数据加载状态
  const [isAnimating, setIsAnimating] = useState<boolean>(false) // 动画状态

  // 获取当前 threadId
  const { currentThreadId } = useAGUIState()

  // 组件挂载后延迟展开到全屏，产生动画效果
  useEffect(() => {
    // 先显示组件，让用户看到初始状态
    const showTimer = setTimeout(() => {
      setIsAnimating(true)
    }, 100)

    // 然后开始全屏动画
    const fullscreenTimer = setTimeout(() => {
      setIsFullscreen(true)
      // 动画完成后移除脉冲效果
      setTimeout(() => {
        setIsAnimating(false)
      }, 700)
    }, 800)

    return () => {
      clearTimeout(showTimer)
      clearTimeout(fullscreenTimer)
    }
  }, [])

  // findings 数据缓存：key 为 serviceName，value 为 findings 数据
  const findingsCache = useRef<Map<string, FindingsNodePayload[] | null>>(new Map())
  // 正在请求的服务列表，避免重复请求
  const loadingServices = useRef<Set<string>>(new Set())

  // 获取服务的 findings 数据（带缓存）
  const fetchServiceFindingsData = useCallback(
    async (serviceName: string): Promise<FindingsNodePayload[] | null> => {
      // 检查是否有 threadId
      if (!currentThreadId) {
        console.warn('无法获取 findings 数据：缺少 threadId')
        return null
      }

      // 检查缓存
      if (findingsCache.current.has(serviceName)) {
        return findingsCache.current.get(serviceName)
      }

      // 检查是否正在请求中
      if (loadingServices.current.has(serviceName)) {
        return null
      }

      try {
        loadingServices.current.add(serviceName)
        setLoadingFindings(true)

        const response = await getFindingsDataApi({
          threadId: currentThreadId,
          serviceName: serviceName,
        })

        if (response?.nodes?.length > 0) {
          const findingsData = response.nodes
          findingsCache.current.set(serviceName, findingsData)
          return findingsData
        }
      } catch (error) {
        console.error(`获取服务 ${serviceName} 的 findings 数据失败:`, error)
        // 缓存失败结果，避免重复请求
        findingsCache.current.set(serviceName, null)
      } finally {
        loadingServices.current.delete(serviceName)
        setLoadingFindings(false)
      }

      return null
    },
    [currentThreadId],
  )

  // 解析传入 data：字符串 -> JSON.parse；对象则直接使用；否则报错
  const { parsedData, parseError, errorMessage } = useMemo(() => {
    let parsed: any = undefined
    let err = false
    let msg = ''
    try {
      if (data == null) {
        // 允许为空时使用 mock 数据
        return
      } else if (typeof data === 'string') {
        parsed = JSON.parse(data)
      } else if (typeof data === 'object') {
        parsed = data
      } else {
        err = true
        msg = '无效的数据类型，期望为 JSON 字符串或对象'
      }
    } catch (e) {
      err = true
      msg = 'JSON 解析失败：' + (e as Error).message
    }

    if (!err && (parsed === null || typeof parsed !== 'object' || Array.isArray(parsed))) {
      err = true
      msg = '解析结果不是对象类型'
    }

    return { parsedData: err ? undefined : parsed, parseError: err, errorMessage: msg }
  }, [data])
  console.log('parsedData', parsedData)

  const handleServiceSelect = useCallback(
    async (service: any) => {
      setSelectedService(service)

      const shouldFetchFindings = service && service.serviceName
      if (shouldFetchFindings) {
        const findingsData = await fetchServiceFindingsData(service.serviceName)
        console.log('findingsData', findingsData)
        // 如果获取到数据，更新选中服务的 evidence
        if (findingsData) {
          setSelectedService((prev) => ({
            ...prev,
            evidence: {
              ...prev?.evidence,
              findingsData,
            },
          }))
        }
      }
    },
    [fetchServiceFindingsData],
  )

  // 默认选择第一个 suspected 的服务节点（在解析成功时）
  useEffect(() => {
    if (!parseError) {
      const source = parsedData
      if (
        source?.suspectedRootCauseNodesDetails &&
        source.suspectedRootCauseNodesDetails.length > 0
      ) {
        const service = source.suspectedRootCauseNodesDetails[0]
        service.type = 'suspectedRootCauseNodesDetails'
        // 使用 handleServiceSelect 来选择服务，这样会自动获取 findings 数据
        handleServiceSelect(service)
      }
    }
  }, [parseError, parsedData, handleServiceSelect])

  const toggleFullscreen = () => {
    setIsFullscreen(!isFullscreen)
  }

  // 清理缓存的函数
  const clearFindingsCache = useCallback(() => {
    findingsCache.current.clear()
    loadingServices.current.clear()
  }, [])

  // 组件卸载时清理缓存
  useEffect(() => {
    return () => {
      clearFindingsCache()
    }
  }, [clearFindingsCache])

  // 解析失败：展示错误信息且不需要放大（不显示放大按钮、容器采用正常高度）
  if (parseError) {
    return (
      <div
        className="w-full h-[500px] rounded-2xl border-2 overflow-hidden shadow-xl"
        style={{
          background: `linear-gradient(135deg, var(--ant-color-bg-container), var(--ant-color-fill-quaternary))`,
          borderColor: 'var(--ant-color-border)',
        }}
      >
        <div className="h-full flex items-center justify-center p-6">
          <Alert
            message={t('evidencePanel.dataError')}
            description={errorMessage || t('evidencePanel.parseError')}
            type="error"
            showIcon
          />
        </div>
      </div>
    )
  }

  const effectiveData = parsedData

  return (
    <div
      className={`flex overflow-hidden shadow-xl backdrop-blur-sm transition-all duration-700 ease-in-out transform ${
        isFullscreen
          ? 'fixed inset-0 z-50 rounded-none border-0 scale-100'
          : 'w-full h-[500px] rounded-2xl border-2 scale-95'
      }`}
      style={{
        background: isFullscreen
          ? 'var(--ant-color-bg-container)'
          : `linear-gradient(135deg, var(--ant-color-bg-container), var(--ant-color-fill-quaternary))`,
        borderColor: isFullscreen ? 'transparent' : 'var(--ant-color-border)',
        zIndex: isFullscreen ? 99999 : 1,
        transformOrigin: 'center center',
        opacity: isAnimating ? 1 : 1,
      }}
    >
      <div
        className="w-1/6 min-w-[250px] h-full border-r backdrop-blur-sm flex-0"
        style={{
          borderColor: 'var(--ant-color-border)',
          backgroundColor: 'var(--ant-color-bg-container)',
          opacity: 0.9,
        }}
      >
        <ServiceList
          onServiceSelect={handleServiceSelect}
          selectedService={selectedService}
          evidenceData={effectiveData}
        />
      </div>
      <div
        className="flex-1 h-full backdrop-blur-sm relative"
        style={{
          backgroundColor: 'var(--ant-color-bg-container)',
        }}
      >
        {/* 页面全屏切换按钮（解析成功时才展示） */}
        <Button
          type="text"
          icon={isFullscreen ? <FullscreenExitOutlined /> : <FullscreenOutlined />}
          onClick={toggleFullscreen}
          className="absolute top-4 right-4 z-10 hover:bg-[var(--ant-color-fill-secondary)] transition-all duration-200"
          style={{
            backgroundColor: 'var(--ant-color-bg-container)',
            border: '1px solid var(--ant-color-border)',
            borderRadius: '8px',
            boxShadow: '0 2px 8px rgba(0,0,0,0.1)',
          }}
          size="large"
          title={
            isFullscreen ? t('evidencePanel.exitFullscreen') : t('evidencePanel.enterFullscreen')
          }
        />
        <ServiceDetail selectedService={selectedService} loadingFindings={loadingFindings} />
      </div>
    </div>
  )
}

export default EvidencePanel
