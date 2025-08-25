import React, { useEffect, useState } from 'react'
import { Tooltip } from 'antd'
import { useTranslation } from 'react-i18next'
import { IoSettingsOutline, IoWarning } from 'react-icons/io5'
import { IoMdWarning } from 'react-icons/io'
import { LuServer } from 'react-icons/lu'
import { dataplaneIconMap } from 'src/pro/constants'
import { useSelector } from 'react-redux'
import { selectProcessedTimeRange } from 'src/core/store/reducers/timeRangeReducer'
import { getDataplaneDatasourcesApi } from 'src/pro/api/dataplane'
import styles from './DataStatus.module.scss'

const DataStatus: React.FC = () => {
  const { t } = useTranslation('pro/agent')
  const { startTime, endTime } = useSelector(selectProcessedTimeRange)
  const [dataplane, setDataplane] = useState<any[]>([])
  const [serviceCount, setServiceCount] = useState<number>(0)

  const getStatus = () => {
    getDataplaneDatasourcesApi({
      startTime,
      endTime,
    }).then((res: any) => {
      const { serviceCount: sc, datasources } = res || {}
      setServiceCount(sc || 0)
      setDataplane(datasources || [])
    })
  }

  useEffect(() => {
    getStatus()
  }, [startTime, endTime])

  return (
    <Tooltip
      title={
        <div>
          <div className="text-nowrap">{t('header.datasourceStatus.title')}</div>
          <div>
            {dataplane?.reduce((sum, item) => sum + item.count, 0) === 0 ? (
              <div
                className="flex items-center gap-2 text-xs my-2 cursor-pointer"
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
                  className="text-xs  cursor-pointer flex items-center gap-2 mb-2"
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
                      <img src={dataplaneIconMap[item.type]} className="w-5 h-5" /> {item.count}
                    </div>
                  ))}
                </div>
              </>
            )}
          </div>
          {serviceCount === 0 ? (
            <div
              className="flex items-center gap-2 text-xs mb-2 cursor-pointer"
              onClick={() => {
                window.open('/#/dataplane?tab=serviceMatcher', '_blank')
              }}
            >
              <IoMdWarning className="text-yellow-500" />
              <span className="text-red-500 underline">{t('header.serviceStatus.noService')}</span>
            </div>
          ) : (
            <div className="">
              <div
                className="text-xs  cursor-pointer mr-2 flex items-center gap-2"
                onClick={() => {
                  window.open('/#/dataplane?tab=serviceMatcher', '_blank')
                }}
              >
                <IoSettingsOutline />
                <span className="underline">{t('header.serviceStatus.currentServices')}</span>
              </div>
              {serviceCount > 0 && (
                <div className="flex items-center gap-2">
                  <LuServer className="text-emerald-400" size={16} />
                  {serviceCount}
                </div>
              )}
            </div>
          )}
        </div>
      }
    >
      <div className={styles.loaderContainer} style={{ width: '300px' }}>
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
