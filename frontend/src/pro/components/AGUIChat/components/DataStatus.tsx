import React, { useEffect, useState } from 'react'
import { Tooltip } from 'antd'
import { useTranslation } from 'react-i18next'
import { IoWarning, IoSettingsOutline } from 'react-icons/io5'
import { IoMdCheckmarkCircle, IoMdRadioButtonOff } from 'react-icons/io'
import { getTimestampRange } from 'src/core/store/reducers/timeRangeReducer'
import { getDataplaneDatasourcesApi, getCapabilitiesApi } from 'src/pro/api/dataplane'
import styles from './DataStatus.module.scss'

interface DataSource {
  type: string
  name: string
}

interface CapabilityItem {
  type: string
  enabled: boolean
  sources: DataSource[]
  description: string
}

const DataStatus: React.FC = () => {
  const { t } = useTranslation('pro/agent')
  const { startTime, endTime } = getTimestampRange('15m')
  const [dataplane, setDataplane] = useState<any[]>([])
  const [serviceCount, setServiceCount] = useState<number>(0)
  const [capabilities, setCapabilities] = useState<CapabilityItem[]>([])

  const getStatus = () => {
    getDataplaneDatasourcesApi({
      startTime,
      endTime,
    }).then((res: any) => {
      const { serviceCount: sc, datasources } = res || {}
      setServiceCount(sc || 0)
      setDataplane(datasources || [])
    })

    getCapabilitiesApi().then((res: any) => {
      const sortedCapabilities = (res?.totalSupported || []).sort(
        (a: CapabilityItem, b: CapabilityItem) => a.type.localeCompare(b.type),
      )
      setCapabilities(sortedCapabilities)
    })
  }

  useEffect(() => {
    if (startTime && endTime) {
      getStatus()
    }
  }, [])

  const handleCapabilityClick = (type: string) => {
    // 根据 type 跳转到对应的设置页面
    // TODO: 根据不同的 type 跳转到不同的配置页面
    console.log('Navigating to settings for type:', type)
    window.open('/#/dataplane?tab=dataplane&filter=' + type, '_blank')
  }

  return (
    <Tooltip
      title={
        <div>
          <div className="text-nowrap mb-3">{t('header.datasourceStatus.title')}</div>

          {/* 数据能力状态 */}
          {capabilities.length > 0 && (
            <div className="mb-3">
              <div className="text-xs font-semibold mb-2">{t('dataCapability.title')}</div>
              <div className="space-y-1">
                {capabilities.map((capability) => (
                  <div
                    key={capability.type}
                    className="flex items-center justify-between gap-2 text-xs hover:bg-black/10 rounded px-2 py-1 transition-colors"
                  >
                    <div className="flex items-center gap-2">
                      {capability.enabled ? (
                        <IoMdCheckmarkCircle className="text-green-500 flex-shrink-0" size={14} />
                      ) : (
                        <IoMdRadioButtonOff className="text-gray-400 flex-shrink-0" size={14} />
                      )}
                      <span>{capability.type}</span>
                    </div>
                    <div
                      className="flex items-center gap-1 text-gray-400 hover:text-blue-400 cursor-pointer"
                      onClick={() => handleCapabilityClick(capability.type)}
                    >
                      <IoSettingsOutline className="flex-shrink-0" size={14} />
                      <span className="text-[10px]">{t('dataCapability.config')}</span>
                    </div>
                  </div>
                ))}
              </div>
            </div>
          )}

          {/* 数据源状态 */}
          {/* <div className="mb-3 border-t border-gray-600 pt-3">
            {dataplane?.reduce((sum, item) => sum + item.count, 0) === 0 ? (
              <div
                className="flex items-center gap-2 text-xs cursor-pointer"
                onClick={() => {
                  window.open('/#/dataplane?tab=dataplane', '_blank')
                }}
              >
                <IoMdWarning className="text-yellow-500" />
                <span className="text-red-500 underline">
                  {t('header.datasourceStatus.noDatasource')}
                </span>
              </div>
            ) : (
              <>
                <div
                  className="text-xs cursor-pointer flex items-center gap-2 mb-2"
                  onClick={() => {
                    window.open('/#/dataplane?tab=dataplane', '_blank')
                  }}
                >
                  <IoSettingsOutline />
                  <span className="underline">
                    {t('header.datasourceStatus.currentDatasources')}
                  </span>
                </div>
                <div className="flex items-center gap-3">
                  {dataplane.map((item: any) => (
                    <div key={item.type} className="flex items-center gap-2">
                      <img src={dataplaneIconMap[item.type]} className="w-5 h-5" alt={item.type} />{' '}
                      {item.count}
                    </div>
                  ))}
                </div>
              </>
            )}
          </div> */}

          {/* 服务状态 */}
          <div className="border-t border-gray-600 pt-3">
            <div className="text-xs font-semibold mb-2">
              {t('dataCapability.serviceMonitorStatus')}
            </div>
            <div className="flex items-center justify-between gap-2 text-xs hover:bg-black/10 rounded px-2 py-1 transition-colors">
              <div className="flex items-center gap-2">
                {serviceCount > 0 ? (
                  <IoMdCheckmarkCircle className="text-green-500 flex-shrink-0" size={14} />
                ) : (
                  <IoMdRadioButtonOff className="text-gray-400 flex-shrink-0" size={14} />
                )}
                <span>{t('dataCapability.services')}</span>
                {serviceCount > 0 && <span className="text-gray-400">({serviceCount})</span>}
              </div>
              <div
                className="flex items-center gap-1 text-gray-400 hover:text-blue-400 cursor-pointer"
                onClick={() => {
                  window.open('/#/dataplane?tab=serviceMatcher', '_blank')
                }}
              >
                <IoSettingsOutline className="flex-shrink-0" size={14} />
                <span className="text-[10px]">{t('dataCapability.config')}</span>
              </div>
            </div>
          </div>
        </div>
      }
    >
      <div className={styles.loaderContainer}>
        <div className="flex items-center">
          {dataplane?.reduce((sum, item) => sum + item.count, 0) === 0 ? (
            <IoWarning className="text-yellow-500  mr-1" />
          ) : (
            <div className={styles.loader}></div>
          )}
          <span className="text-xs text-[var(--ant-color-text-tertiary)]">
            {t('header.badges.datasource')}
          </span>
        </div>
        <div className="flex items-center">
          {serviceCount === 0 ? (
            <IoWarning className="text-yellow-500  mr-1" />
          ) : (
            <div className={styles.loader}></div>
          )}
          <span className="text-xs text-[var(--ant-color-text-tertiary)]">
            {t('header.badges.services')}
          </span>
        </div>
      </div>
    </Tooltip>
  )
}

export default DataStatus
