/**
 * Copyright 2024 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

//此组件用于页面内的时间选择 不具备存储功能 目前用于log和trace页面

import { cilCalendar, cilClock } from '@coreui/icons'
import CIcon from '@coreui/icons-react'
import {
  CDropdown,
  CDropdownMenu,
  CDropdownToggle,
  CForm,
  CFormFeedback,
  CFormInput,
  CFormLabel,
  CInputGroup,
  CInputGroupText,
} from '@coreui/react'
import React, { useEffect, useState } from 'react'
import { DateRange } from 'react-date-range'
import { useSearchParams } from 'react-router-dom'
import {
  GetInitalTimeRange,
  getTimestampRange,
  timeRangeMap,
} from 'src/core/store/reducers/timeRangeReducer'
import { convertTime, DateToISO, ISOToTimestamp, timeUtils } from 'src/core/utils/time'
import './index.css'
import 'react-date-range/dist/styles.css' // main style file
import 'react-date-range/dist/theme/default.css' // theme css file
import { useSelector } from 'react-redux'
import { useTranslation } from 'react-i18next'
import { Button, Menu } from 'antd'
import { useLogsTraceFilterContext } from 'src/oss/contexts/LogsTraceFilterContext'
// logs: logsFrom, logsTo
// traces: traceFrom, traceTo
const TypeUrlParamMap = {
  logs: {
    from: 'logs-from',
    to: 'logs-to',
  },
  traces: {
    from: 'trace-from',
    to: 'trace-to',
  },
}
export default function DateTimeRangePickerCom(props) {
  const { t } = useTranslation('core/dateTime')
  const { type, defaultTimerange = null } = props
  const [searchParams, setSearchParams] = useSearchParams()
  const [dropdownVisible, setDropdownVisible] = useState(false)
  const { setStartTime, setEndTime } = useLogsTraceFilterContext((ctx) => ctx)
  // 快速范围
  // 左侧输入框是否合规以及违反规则的反馈文本
  const [startTimeInvalid, setStartTimeInvalid] = useState(false)
  const [endTimeInvalid, setEndTimeInvalid] = useState(false)
  const [startTimeFeedback, setStartTimeFeedback] = useState()
  const [endTimeFeedback, setEndTimeFeedback] = useState()
  const [inputStartTime, setInputStartTime] = useState()
  const [inputEndTime, setInputEndTime] = useState()
  const storeTimeRange = useSelector((state) => state.timeRange)
  const [dateRange, setDateRange] = useState({
    startDate: timeUtils.startOfDay(new Date()),
    endDate: timeUtils.endOfDay(new Date()),
    key: 'selection',
  })
  // yyyy-mm-dd hh:mm:ss 时间输入
  const updataUrlTimeRange = (fromString, toString) => {
    if (type) {
      const params = new URLSearchParams(searchParams)
      const from = searchParams.get(type + '-from')
      const to = searchParams.get(type + '-to')
      let needChangeUrl = false
      const fromStringToISO = DateToISO(fromString)
      const toStringToISO = DateToISO(toString)
      if (fromStringToISO !== from) {
        params.set(type + '-from', fromStringToISO)
        needChangeUrl = true
      }
      if (toStringToISO !== to) {
        params.set(type + '-to', toStringToISO)
        needChangeUrl = true
      }
      if (needChangeUrl) {
        setSearchParams(params, { replace: true })
      }
    }
  }
  const confirmTimeRange = (event) => {
    if (startTimeInvalid || endTimeInvalid) {
      event.preventDefault()
      event.stopPropagation()
      return
    }
    setStartTime?.(new Date(inputStartTime).getTime() * 1000)
    setEndTime?.(new Date(inputEndTime).getTime() * 1000)
    setDropdownVisible(false)
    const fromISO = DateToISO(inputStartTime)
    const toISO = DateToISO(inputEndTime)
    updataUrlTimeRange(fromISO, toISO)
  }

  const changeInputStartTime = (event) => {
    setInputStartTime(event.target.value)
  }
  const changeInputEndTime = (event) => {
    setInputEndTime(event.target.value)
  }
  const handleDateRange = (ranges) => {
    let state = ranges.selection
    setDateRange({
      startDate: timeUtils.startOfDay(state.startDate),
      endDate: timeUtils.endOfDay(state.endDate),
      key: state.key,
    })
    setInputStartTime(timeUtils.format(state.startDate, 'yyyy-MM-dd HH:mm:ss'))
    setInputEndTime(timeUtils.format(timeUtils.endOfDay(state.endDate), 'yyyy-MM-dd HH:mm:ss'))
  }

  const handleTimeRange = (key) => {
    const { startTime, endTime } = getTimestampRange(key)
    const fromString = convertTime(startTime, 'yyyy-mm-dd hh:mm:ss')
    const toString = convertTime(endTime, 'yyyy-mm-dd hh:mm:ss')
    setInputStartTime(fromString)
    setInputEndTime(toString)
    setDropdownVisible(false)
    updataUrlTimeRange(fromString, toString)
    setStartTime?.(new Date(fromString).getTime() * 1000)
    setEndTime?.(new Date(toString).getTime() * 1000)
  }
  const validStartTime = () => {
    let feedback = t('dateTimeRangePicker.selectCorrectTimeRangeFeedback')
    let result = true
    if (!inputStartTime || inputStartTime?.length === 0 || !timeUtils.isValid(inputStartTime)) {
      result = false
    } else if (new Date(inputStartTime) > new Date(inputEndTime)) {
      feedback = t('dateTimeRangePicker.startTimeLongerThanEndTimeFeedback')
      result = false
    }
    if (result) {
      feedback = ''
    }
    setStartTimeFeedback(feedback)
    setStartTimeInvalid(!result)
    return result
  }
  const validEndTime = () => {
    let feedback = t('dateTimeRangePicker.selectCorrectTimeRangeFeedback')
    let result = true
    if (!inputEndTime || inputEndTime?.length === 0 || !timeUtils.isValid(inputEndTime)) {
      result = false
    } else if (new Date(inputStartTime) > new Date(inputEndTime)) {
      feedback = t('dateTimeRangePicker.endTimeLessThanStartTimeFeedback')
      result = false
    }
    if (result) {
      feedback = ''
    }
    setEndTimeFeedback(feedback)
    setEndTimeInvalid(!result)
    return result
  }
  const initTimeRange = () => {
    let initFromString, initToString
    if (storeTimeRange.startTime && storeTimeRange.endTime) {
      initFromString = timeUtils.microTimestampToIso(storeTimeRange.startTime)
      initToString = timeUtils.microTimestampToIso(storeTimeRange.endTime)
    } else {
      const initTimeRange = GetInitalTimeRange()
      initFromString = convertTime(initTimeRange.startTime, 'yyyy-mm-dd hh:mm:ss')
      initToString = convertTime(initTimeRange.endTime, 'yyyy-mm-dd hh:mm:ss')
    }

    setInputStartTime(initFromString, 'yyyy-mm-dd hh:mm:ss')
    setInputEndTime(initToString, 'yyyy-mm-dd hh:mm:ss')
    updataUrlTimeRange(initFromString, initToString)
    setStartTime(new Date(initFromString).getTime() * 1000)
    setEndTime(new Date(initToString).getTime() * 1000)
  }
  useEffect(() => {
    const from = searchParams.get(type + '-from')
    const to = searchParams.get(type + '-to')
    // console.log(type,'url',from,to)
    if (defaultTimerange) {
      return
    }
    if ((!from || !to) && (!inputStartTime || !inputEndTime)) {
      initTimeRange()
      return
    }
    //iso -> timestamp -> string
    const fromString = convertTime(ISOToTimestamp(from), 'yyyy-mm-dd hh:mm:ss')
    const toString = convertTime(ISOToTimestamp(to), 'yyyy-mm-dd hh:mm:ss')
    if (fromString && toString && timeUtils.isValid(fromString) && timeUtils.isValid(toString)) {
      if (inputStartTime !== fromString || inputEndTime !== toString) {
        // console.log('url发现参数和store不符，更新精确时间')
        setInputStartTime(fromString)
        setInputEndTime(toString)
        updataUrlTimeRange(fromString, toString)
        setStartTime(new Date(fromString).getTime() * 1000)
        setEndTime(new Date(toString).getTime() * 1000)
      }
    } else {
      initTimeRange()
    }
  }, [searchParams])
  useEffect(() => {
    if (defaultTimerange) {
      const fromString = convertTime(defaultTimerange[0], 'yyyy-mm-dd hh:mm:ss')
      const toString = convertTime(defaultTimerange[1], 'yyyy-mm-dd hh:mm:ss')
      setInputStartTime(fromString)
      setInputEndTime(toString)
      // setStartTime(new Date(fromString).getTime() * 1000)
      // setEndTime(new Date(toString).getTime() * 1000)
    }
  }, [defaultTimerange])
  useEffect(() => {
    const startTimeValid = validStartTime()
    const endTimeValid = validEndTime()
    if (startTimeValid && endTimeValid) {
      setDateRange({
        startDate: timeUtils.startOfDay(new Date(inputStartTime).getTime()),
        endDate: timeUtils.endOfDay(new Date(inputEndTime).getTime()),
        key: 'selection',
      })
    }
  }, [inputStartTime, inputEndTime])
  return (
    <CDropdown
      autoClose="outside"
      visible={dropdownVisible}
      onShow={() => setDropdownVisible(true)}
      onHide={() => setDropdownVisible(false)}
    >
      <CDropdownToggle className="" size="sm" onClick={() => setDropdownVisible(true)}>
        <CIcon icon={cilClock} className="mr-2" />
        <span className="text-sm">
          {inputStartTime} to {inputEndTime}
          {/* {displayStartTime + ' to ' + displayEndTime} */}
        </span>
      </CDropdownToggle>
      <CDropdownMenu className="m-0 p-0">
        <div className="w-[600px] flex">
          <div className="w-3/5 border-r border-r-slate-300  px-3 py-2">
            <CForm noValidate>
              <div>{t('dateTimeRangePicker.absoluteTimeRangeTitle')}</div>

              <CFormLabel className="text-sm mt-2 block">
                {t('dateTimeRangePicker.startTitle')}
              </CFormLabel>

              <CInputGroup className="mb-2">
                <CFormInput
                  value={inputStartTime}
                  type="text"
                  onChange={changeInputStartTime}
                  required
                  invalid={startTimeInvalid}
                />
                <CInputGroupText>
                  <CIcon icon={cilCalendar} />
                </CInputGroupText>
                <CFormFeedback invalid>{startTimeFeedback}</CFormFeedback>
              </CInputGroup>

              <CFormLabel htmlFor="basic-url" className="text-sm block">
                {t('dateTimeRangePicker.endTitle')}
              </CFormLabel>
              <CInputGroup>
                <CFormInput
                  value={inputEndTime}
                  onChange={changeInputEndTime}
                  required
                  invalid={endTimeInvalid}
                />
                <CInputGroupText id="basic-addon2">
                  <CIcon icon={cilCalendar} />
                </CInputGroupText>
                <CFormFeedback invalid>{endTimeFeedback}</CFormFeedback>
              </CInputGroup>
              <Button type="primary" className="mt-3" onClick={confirmTimeRange}>
                {t('dateTimeRangePicker.applyTimeRangeButtonText')}
              </Button>
            </CForm>
            <DateRange
              moveRangeOnFirstSelection={false}
              ranges={[dateRange]}
              onChange={handleDateRange}
              editableDateInputs={true}
              hh
              showDateDisplay={false}
            />
          </div>
          <div className="w-2/5">
            <div className="p-2">{t('dateTimeRangePicker.quickRangeTitle')}</div>
            <Menu
              selectedKeys={[storeTimeRange.rangeTypeKey]}
              onClick={({ key }) => handleTimeRange(key)}
            >
              {Object.keys(timeRangeMap).map((key) => (
                <Menu.Item key={key}>{timeRangeMap[key].name}</Menu.Item>
              ))}
            </Menu>
          </div>
        </div>
      </CDropdownMenu>
    </CDropdown>
  )
}
