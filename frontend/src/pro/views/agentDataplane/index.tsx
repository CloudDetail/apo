import { Tabs } from 'antd'
import DataplanePage from './DataplanePage'
import { BasicCard } from 'src/core/components/Card/BasicCard'
import ServiceMatcherPage from './ServiceMatcherPage'
import { getTimestampRange } from 'src/core/store/reducers/timeRangeReducer'
import { useState, useEffect } from 'react'
import DateTimeRangePickerCom from './components/DateTimeRangePickerCom'
import { useTranslation } from 'react-i18next'
import { useSearchParams } from 'react-router-dom'
import styles from './index.module.scss'
const AgentDataplanePage = () => {
  const { t } = useTranslation('pro/dataplane')
  const [searchParams, setSearchParams] = useSearchParams()

  // 从URL参数获取tab key，默认为'dataplane'
  const getInitialTabKey = () => {
    const tabKey = searchParams.get('tab')
    return tabKey === 'serviceMatcher' ? 'serviceMatcher' : 'dataplane'
  }

  const [key, setKey] = useState(getInitialTabKey)

  // 当URL参数变化时更新tab key
  useEffect(() => {
    const newTabKey = getInitialTabKey()
    if (newTabKey !== key) {
      setKey(newTabKey)
    }
  }, [searchParams])

  const defaultTimerange = {
    startTime: getTimestampRange('last7Days').startTime,
    endTime: getTimestampRange('last7Days').endTime,
    rangeType: 'last7Days',
  }
  const [timeRange, setTimeRange] = useState(defaultTimerange)

  const refreshTimeRange = () => {
    if (timeRange.rangeType) {
      setTimeRange({
        startTime: getTimestampRange(timeRange.rangeType).startTime,
        endTime: getTimestampRange(timeRange.rangeType).endTime,
        rangeType: timeRange.rangeType,
      })
    }
  }

  const tabItems = [
    {
      key: 'dataplane',
      label: t('dataSourceAccess'),
      children: <DataplanePage />,
    },
    {
      key: 'serviceMatcher',
      label: t('monitoringServiceList'),
      children: <ServiceMatcherPage timeRange={timeRange} refreshTimeRange={refreshTimeRange} />,
    },
  ]

  const changeTimeRange = (timeRange) => {
    setTimeRange(timeRange)
  }

  // 处理tab切换，更新URL参数
  const handleTabChange = (newKey) => {
    setKey(newKey)
    const newParams = new URLSearchParams(searchParams)
    if (newKey === 'dataplane') {
      newParams.delete('tab')
    } else {
      newParams.set('tab', newKey)
    }
    setSearchParams(newParams, { replace: true })
  }

  return (
    <>
      <BasicCard>
        <Tabs
          activeKey={key}
          onChange={handleTabChange}
          items={tabItems}
          className={styles.dataplane}
          tabBarExtraContent={
            key === 'serviceMatcher' && (
              <DateTimeRangePickerCom
                changeTimeRange={changeTimeRange}
                defaultTimerange={timeRange}
              />
            )
          }
        />
      </BasicCard>
    </>
  )
}

export default AgentDataplanePage
