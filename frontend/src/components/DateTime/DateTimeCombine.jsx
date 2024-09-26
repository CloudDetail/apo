import React, { useEffect, useState } from 'react'
import TimeSinceRefresh from './TimeSinceRefresh'
import DateTimeRangePicker from './DateTimeRangePicker'
import { Button, Input, Popover, Segmented, Tooltip } from 'antd'
import { FiSend } from 'react-icons/fi'
import CopyButton from '../CopyButton'
import { useLocation, useSearchParams } from 'react-router-dom'
import { useSelector } from 'react-redux'
import { convertTime, TimestampToISO } from 'src/utils/time'
import { timeRangeMap } from 'src/store/reducers/timeRangeReducer'
import RefreshDateTime from './RefreshDateTime'

function ShareLink() {
  const location = useLocation()
  const [shareType, setShareType] = useState('绝对时间')
  const { rangeTypeKey, startTime, endTime } = useSelector((state) => state.timeRange)
  const [copyUrl, setCopyUrl] = useState(window.location.href)
  const [searchParams] = useSearchParams()
  useEffect(() => {
    let url = window.location.href
    if (shareType === '绝对时间' && rangeTypeKey) {
      const currentUrl = new URL(window.location.href)
      searchParams.delete('relativeTime')
      searchParams.set('from', TimestampToISO(startTime))
      searchParams.set('to', TimestampToISO(endTime))
      url = currentUrl.origin + '/#' + location.pathname + '?' + searchParams.toString()
    }
    setCopyUrl(url)
  }, [shareType, rangeTypeKey, location])
  return (
    <>
      <Popover
        content={
          <div className="w-[500px]">
            <Segmented
              options={[
                '绝对时间',
                {
                  label: '相对时间',
                  value: '相对时间',
                  disabled: !rangeTypeKey,
                },
              ]}
              onChange={setShareType}
            />
            <div className="my-2 text-gray-300 mx-1">
              {shareType === '绝对时间' ? (
                <>
                  {convertTime(startTime, 'yyyy-mm-dd hh:mm:ss')} to{' '}
                  {convertTime(endTime, 'yyyy-mm-dd hh:mm:ss')}
                </>
              ) : (
                timeRangeMap[rangeTypeKey].name
              )}
            </div>
            <Input value={copyUrl} addonAfter={<CopyButton value={copyUrl} />} />
          </div>
        }
      >
        <Button type="text" icon={<FiSend />}></Button>
      </Popover>
    </>
  )
}

export default function DateTimeCombine() {
  const { rangeTypeKey } = useSelector((state) => state.timeRange)
  return (
    <div className="flex items-center">
      {rangeTypeKey && <TimeSinceRefresh />} <DateTimeRangePicker />
      {rangeTypeKey && <RefreshDateTime />}
      <ShareLink />
    </div>
  )
}
