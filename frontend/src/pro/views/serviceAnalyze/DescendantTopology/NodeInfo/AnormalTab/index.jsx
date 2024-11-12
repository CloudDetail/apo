import { Tabs, Tooltip } from 'antd'
import React from 'react'
import AnormalList from '../AnormalList'
import { useAlertAnalyzeContext } from 'src/pro/contexts/AlertAnalyzeContext'
import { convertTime } from 'src/core/utils/time'
import { useSelector } from 'react-redux'
import { selectSecondsTimeRange } from 'src/core/store/reducers/timeRangeReducer'

const AnormalTab = () => {
  const { deltaStartTime, deltaEndTime } = useAlertAnalyzeContext()
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
          <span>当前播放时间段事件</span>
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
          <span>分析总时间段事件</span>
        </Tooltip>
      ),
      children: <AnormalList />,
    },
  ]
  return <Tabs defaultActiveKey="1" items={items} />
}
export default AnormalTab
