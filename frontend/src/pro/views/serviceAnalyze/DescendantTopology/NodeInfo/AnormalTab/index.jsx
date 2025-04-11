import { Tabs, Tooltip } from 'antd'
import React from 'react'
import AnormalList from '../AnormalList'
import { useServiceAnalyzeContext } from 'src/pro/contexts/ServiceAnalyzeContext'
import { convertTime } from 'src/core/utils/time'
import { useSelector } from 'react-redux'
import { selectSecondsTimeRange } from 'src/core/store/reducers/timeRangeReducer'
import { useTranslation } from 'react-i18next'
const i18nKey = 'node'

const AnormalTab = () => {
  const { t } = useTranslation('pro/serviceAnalyze')
  const { deltaStartTime, deltaEndTime } = useServiceAnalyzeContext()
  const { startTime, endTime } = useSelector(selectSecondsTimeRange)
  const items = [
    {
      key: '1',
      label: (
        <Tooltip
          title={
            <span>
              {convertTime(deltaStartTime, 'yyyy-mm-dd hh:mm:ss')} 至{' '}
              {convertTime(deltaEndTime, 'yyyy-mm-dd hh:mm:ss')}
            </span>
          }
        >
          <span>{t(`${i18nKey}.currentTimeEvents`)}</span>
        </Tooltip>
      ),
      children: <AnormalList />,
    },
    {
      key: '2',
      label: (
        <Tooltip
          title={
            <span>
              {convertTime(startTime, 'yyyy-mm-dd hh:mm:ss')} 至{' '}
              {convertTime(endTime, 'yyyy-mm-dd hh:mm:ss')}
            </span>
          }
        >
          <span>{t(`${i18nKey}.totalTimeEvents`)}</span>
        </Tooltip>
      ),
      children: <AnormalList />,
    },
  ]
  return <Tabs defaultActiveKey="1" items={items} />
}
export default AnormalTab
