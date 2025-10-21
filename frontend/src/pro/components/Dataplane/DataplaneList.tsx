import { useEffect, useState, useMemo } from 'react'
import { useSearchParams } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import { getDataplaneTypeListApi, getCapabilitiesApi } from 'src/pro/api/dataplane'
import { DataplaneTypeItem } from 'src/pro/types/datasource'
import DataplaneCard from './DataplaneCard'
import DataCoverageOverview from './DataCoverageOverview'
import styles from '../StatusIndicator/index.module.scss'
import { SettingsIcon } from 'lucide-react'
import { Button, Tooltip } from 'antd'

const CAPABILITY_OPTIONS = [
  { value: 'Metrics', label: 'Metrics' },
  { value: 'Log', label: 'Log' },
  { value: 'Trace', label: 'Trace' },
  { value: 'RedMetric', label: 'RedMetric' },
  { value: 'Topology', label: 'Topology' },
  { value: 'ServiceInstance', label: 'ServiceInstance' },
]

const DataplaneList = ({
  onClick,
  refreshKey,
}: {
  onClick: (type: DataplaneTypeItem) => void
  refreshKey?: number
}) => {
  const { t } = useTranslation('pro/agent')
  const [searchParams, setSearchParams] = useSearchParams()
  const [dataplaneList, setDataplaneList] = useState<DataplaneTypeItem[]>([])
  const [capabilityStatus, setCapabilityStatus] = useState<Record<string, boolean>>({})
  const [capabilityDescriptions, setCapabilityDescriptions] = useState<Record<string, string>>({})

  // 从 URL 参数中获取 showSettings 状态
  const showSettingsParam = searchParams.get('showSettings')
  const [showSettings, setShowSettings] = useState(showSettingsParam === 'true')

  // 从 URL 参数中获取筛选条件
  const selectedCapabilities = useMemo(() => {
    const filterParam = searchParams.get('filter')
    return filterParam ? filterParam.split(',') : []
  }, [searchParams])

  // 监听 URL 参数变化，更新 showSettings
  useEffect(() => {
    setShowSettings(showSettingsParam === 'true')
  }, [showSettingsParam])

  const fetchCapabilities = () => {
    getCapabilitiesApi().then((res: any) => {
      const statusMap: Record<string, boolean> = {}
      const descriptionMap: Record<string, string> = {}
      res?.totalSupported?.forEach((item: any) => {
        statusMap[item.type] = item.enabled
        // 优先使用API返回的description，如果没有则使用国际化文件中的描述
        descriptionMap[item.type] = item.description || t(`capabilityDescriptions.${item.type}`)
      })
      setCapabilityStatus(statusMap)
      setCapabilityDescriptions(descriptionMap)
    })
  }

  useEffect(() => {
    getDataplaneTypeListApi().then((res: any) => {
      setDataplaneList(res?.dpTypes || [])
    })

    fetchCapabilities()
  }, [refreshKey])

  const filteredDataplaneList = useMemo(() => {
    if (selectedCapabilities.length === 0) {
      return dataplaneList
    }
    return dataplaneList.filter((item) =>
      selectedCapabilities.every((capability) => item.capabilitySpec.includes(capability)),
    )
  }, [dataplaneList, selectedCapabilities])

  const toggleCapability = (capability: string) => {
    const newFilters = selectedCapabilities.includes(capability)
      ? selectedCapabilities.filter((c) => c !== capability)
      : [...selectedCapabilities, capability]

    // 更新 URL 参数
    if (newFilters.length > 0) {
      setSearchParams({ filter: newFilters.join(',') })
    } else {
      setSearchParams({})
    }
  }

  const handleQuickSetup = (type: string) => {
    // 设置筛选条件为对应的 type，并关闭弹窗
    setSearchParams({ filter: type })
  }

  const handleOpenSettings = () => {
    const params = new URLSearchParams(searchParams)
    params.set('showSettings', 'true')
    setSearchParams(params)
  }

  const handleCloseSettings = () => {
    const params = new URLSearchParams(searchParams)
    params.delete('showSettings')
    setSearchParams(params)
    // 关闭弹窗后刷新能力状态
    fetchCapabilities()
  }

  return (
    <div className="w-full overflow-y-auto h-full flex flex-col gap-4">
      {/* Filter Section */}
      <div className="flex flex-col gap-3">
        <div className="flex items-center justify-between">
          <div className="text-sm font-medium text-[var(--ant-color-text-secondary)]">
            {t('dataplane.filterByCapability')}
          </div>
          <div className="flex items-center gap-4 text-xs text-[var(--ant-color-text-tertiary)]">
            <div className="flex items-center gap-1">
              <div
                className={`w-2 h-2 rounded-full ${styles.loader}`}
                style={{ boxShadow: '0 0 0 2px rgba(77, 217, 0, 0.1)' }}
              ></div>
              <span>{t('dataCapability.configured')}</span>
            </div>
            <div className="flex items-center gap-1">
              <div
                className={`w-2 h-2 rounded-full ${styles.grayLoader}`}
                style={{ boxShadow: '0 0 0 2px rgba(140, 140, 140, 0.1)' }}
              ></div>
              <span>{t('dataCapability.unconfigured')}</span>
            </div>
            <Button
              type="text"
              size="small"
              onClick={handleOpenSettings}
              className="flex items-center justify-center text-xs"
            >
              <SettingsIcon className="w-4 h-4" /> {t('dataplane.seeTheDetail')}
            </Button>
          </div>
        </div>
        <div className="flex flex-wrap gap-2">
          {CAPABILITY_OPTIONS.map((option) => {
            const isSelected = selectedCapabilities.includes(option.value)
            const isEnabled = capabilityStatus[option.value] ?? false
            const baseDescription = t(`capabilityDescriptions.${option.value}`)
            const apiDescription = capabilityDescriptions[option.value]
            // 如果API返回了description且与基础描述不同，则一起展示
            const tooltipContent =
              apiDescription && apiDescription !== baseDescription
                ? `${baseDescription}\n\n${apiDescription}`
                : baseDescription
            return (
              <Tooltip
                key={option.value}
                title={<div style={{ whiteSpace: 'pre-line' }}>{tooltipContent}</div>}
                placement="top"
                overlayStyle={{ maxWidth: '300px' }}
              >
                <button
                  onClick={() => toggleCapability(option.value)}
                  className={`group/filter flex items-center gap-2 px-3 py-1.5 rounded-md text-sm font-medium transition-all duration-200 ${
                    isSelected
                      ? 'bg-[var(--ant-color-primary)] text-white hover:bg-[var(--ant-color-primary-hover)] shadow-sm'
                      : 'bg-[var(--ant-color-fill-quaternary)] text-[var(--ant-color-text-secondary)] hover:bg-[var(--ant-color-fill-quinary)]'
                  }`}
                >
                  <div className="relative flex items-center">
                    <div
                      className={isEnabled ? styles.loader : styles.grayLoader}
                      style={{
                        boxShadow: isSelected
                          ? '0 0 0 4px rgba(255, 255, 255, 0.3)'
                          : isEnabled
                            ? '0 0 0 4px rgba(77, 217, 0, 0.1)'
                            : '0 0 0 4px rgba(140, 140, 140, 0.1)',
                      }}
                    ></div>
                  </div>
                  <span>{option.label}</span>
                </button>
              </Tooltip>
            )
          })}
          {selectedCapabilities.length > 0 && (
            <button
              onClick={() => setSearchParams({})}
              className="px-3 py-1.5 rounded-md text-sm font-medium bg-[var(--ant-color-error-bg)] text-[var(--ant-color-error-text)] hover:bg-[var(--ant-color-error-hover)] transition-all duration-200"
            >
              {t('dataplane.clearAll')}
            </button>
          )}
        </div>
      </div>

      {/* Dataplane List */}
      <div className="flex flex-col gap-3 flex-1 overflow-hidden">
        {/* Results Summary */}
        <div className="flex items-center justify-between px-1 flex-0">
          <span className="text-xs text-[var(--ant-color-text-tertiary)]">
            {filteredDataplaneList.length > 0 ? (
              <>
                {t('dataplane.showing')}{' '}
                <span className="font-semibold text-[var(--ant-color-primary)]">
                  {filteredDataplaneList.length}
                </span>{' '}
                {t('dataplane.of')} <span className="font-semibold">{dataplaneList.length}</span>{' '}
                {t('dataplane.dataplanes')}
              </>
            ) : (
              <span className="text-[var(--ant-color-warning-text)]">
                {t('dataplane.noResultsFound')}
              </span>
            )}
          </span>
        </div>

        {/* Grid */}
        <div className="flex flex-col gap-2 w-full  overflow-y-auto flex-1">
          {filteredDataplaneList.length > 0 ? (
            filteredDataplaneList.map((option) => (
              <div key={option.typeName} onClick={() => onClick(option)}>
                <DataplaneCard info={option} />
              </div>
            ))
          ) : (
            <div className="col-span-full text-center py-12">
              <div className="flex flex-col items-center gap-3">
                <div className="text-4xl">🔍</div>
                <div className="text-[var(--ant-color-text-tertiary)]">
                  {t('dataplane.noDataplaneMatches')}
                </div>
                <button
                  onClick={() => setSearchParams({})}
                  className="mt-2 px-4 py-2 rounded-md text-sm font-medium bg-[var(--ant-color-primary)] text-white hover:bg-[var(--ant-color-primary-hover)] transition-colors"
                >
                  {t('dataplane.clearFilters')}
                </button>
              </div>
            </div>
          )}
        </div>
      </div>
      <DataCoverageOverview
        visible={showSettings}
        onClose={handleCloseSettings}
        onQuickSetup={handleQuickSetup}
      />
    </div>
  )
}

export default DataplaneList
