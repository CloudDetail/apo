import { Button } from 'antd'
import React from 'react'
import {
  TbPlayerPauseFilled,
  TbPlayerPlayFilled,
  TbPlayerTrackNextFilled,
  TbPlayerTrackPrevFilled,
} from 'react-icons/tb'
import { useSelector } from 'react-redux'
import { useServiceAnalyzeContext } from 'src/pro/contexts/ServiceAnalyzeContext'
import { selectSecondsTimeRange } from 'src/core/store/reducers/timeRangeReducer'
import DateTimeRangeSimple from './DateTimeRangeSimple'
import { t } from 'i18next'

const PlayDelta = () => {
  const {
    isPlaying,
    handlePlay,
    deltaStartTime,
    deltaEndTime,
    setDeltaTimerange,
    changeAlertDescendantDelta,
  } = useServiceAnalyzeContext()
  const { startTime, endTime } = useSelector(selectSecondsTimeRange)
  const clickLeftTimerange = () => {
    const step = deltaEndTime - deltaStartTime

    let newRange = {
      deltaStartTime: null,
      deltaEndTime: null,
    }
    if (deltaStartTime - step >= startTime) {
      newRange.deltaStartTime = deltaStartTime - step
      newRange.deltaEndTime = deltaEndTime - step
    } else {
      newRange.deltaStartTime = startTime
      newRange.deltaEndTime = startTime + step
    }
    changeAlertDescendantDelta(newRange)
  }
  const clickRightTimerange = () => {
    const step = deltaEndTime - deltaStartTime
    let newRange = {
      deltaStartTime: null,
      deltaEndTime: null,
    }
    if (deltaEndTime + step <= endTime) {
      newRange.deltaStartTime = deltaStartTime + step
      newRange.deltaEndTime = deltaEndTime + step
    } else {
      newRange.deltaStartTime = endTime - step
      newRange.deltaEndTime = endTime
    }
    changeAlertDescendantDelta(newRange)
  }
  return (
    <div style={{ display: 'flex', flexDirection: 'column', alignItems: 'center' }}>
      <span className="text-xs flex items-center">
        {t('pro/serviceAnalyze:currentPlayTime')}
        <DateTimeRangeSimple type="delta" showTimerange={false} />
        {/* {convertTime(deltaStartTime, 'yyyy-mm-dd hh:mm:ss')} 至{' '}
        {convertTime(deltaEndTime, 'yyyy-mm-dd hh:mm:ss')} */}
      </span>
      <div className="flex">
        <Button
          type="link"
          onClick={clickLeftTimerange}
          disabled={isPlaying || deltaStartTime === startTime}
        >
          <TbPlayerTrackPrevFilled />
        </Button>
        <Button type="link" onClick={() => handlePlay()} disabled={isPlaying}>
          {isPlaying ? <TbPlayerPauseFilled /> : <TbPlayerPlayFilled />}
        </Button>
        <Button
          type="link"
          onClick={clickRightTimerange}
          disabled={isPlaying || deltaEndTime === endTime}
        >
          <TbPlayerTrackNextFilled />
        </Button>
      </div>
    </div>
  )
}
export default PlayDelta
