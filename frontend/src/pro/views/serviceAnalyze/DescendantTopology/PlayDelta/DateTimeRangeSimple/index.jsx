//此组件用于业务告警分析的快速时间选择，跨度一般不会很大

import { cilCalendar, cilClock } from '@coreui/icons'
import CIcon from '@coreui/icons-react'
import {
  CButton,
  CDropdown,
  CDropdownItem,
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
  initialTimeRangeState,
  selectProcessedTimeRange,
  selectSecondsTimeRange,
  timeRangeList,
  timeRangeMap,
} from 'src/core/store/reducers/timeRangeReducer'
import {
  convertTime,
  DateToISO,
  ISOToTimestamp,
  TimestampToISO,
  ValidDate,
  timeUtils,
} from 'src/core/utils/time'
// import './index.css'
import 'react-date-range/dist/styles.css' // main style file
import 'react-date-range/dist/theme/default.css' // theme css file
import { useSelector } from 'react-redux'
import { Button, DatePicker, Dropdown, Popover } from 'antd'
import dayjs from 'dayjs'
import { useServiceAnalyzeContext } from 'src/pro/contexts/ServiceAnalyzeContext'
import { useUpdateEffect } from 'react-use'
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
export default function DateTimeRangeSimple(props) {
  const { type, showTimerange = true, triggerNode = null } = props
  const [searchParams, setSearchParams] = useSearchParams()
  // 左侧输入框是否合规以及违反规则的反馈文本
  const [startTimeInvalid, setStartTimeInvalid] = useState(false)
  const [endTimeInvalid, setEndTimeInvalid] = useState(false)
  const [startTimeFeedback, setStartTimeFeedback] = useState()
  const [endTimeFeedback, setEndTimeFeedback] = useState()
  const [inputStartTime, setInputStartTime] = useState()
  const [inputEndTime, setInputEndTime] = useState()
  const storeTimeRange = useSelector((state) => state.timeRange)
  const { deltaStartTime, deltaEndTime, changeAlertDescendantDelta } = useServiceAnalyzeContext()
  const { startTime, endTime } = useSelector(selectSecondsTimeRange)
  const [dateRange, setDateRange] = useState([new Date(), new Date()])
  // yyyy-mm-dd hh:mm:ss 时间输入
  const updataUrlTimeRange = (fromString, toString) => {
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
    // if (needChangeUrl) {
    //   let url = new URL(window.location.href)

    //   // 设置新的查询参数
    //   url.search = params.toString()

    //   // 如果有 hash 路由，也设置到 URL 的 hash
    //   if (url.hash) {
    //     url.hash = url.hash.split('?')[0] + '?' + params.toString()
    //   }
    //   console.log(url.toString())
    //   // 使用 URL 对象转换为字符串，传入 replaceState
    //   window.history.replaceState(null, '', url.toString())
    // }
  }
  const confirmTimeRange = (event) => {
    if (startTimeInvalid || endTimeInvalid) {
      event.preventDefault()
      event.stopPropagation()
      return
    }
    const fromISO = DateToISO(inputStartTime)
    const toISO = DateToISO(inputEndTime)
    updataUrlTimeRange(fromISO, toISO)
  }

  const initTimeRange = () => {
    let initFromDate, initToDate
    if (deltaStartTime && deltaEndTime) {
      initFromDate = new Date(deltaStartTime / 1000)
      initToDate = new Date(deltaEndTime / 1000)
    }
    console.log(initFromDate, initToDate)

    setDateRange([initFromDate, initToDate])
    if (initFromDate && initToDate) {
      const fromString = timeUtils.format(initFromDate)
      const toString = timeUtils.format(initToDate)
      updataUrlTimeRange(fromString, toString)
    }
  }
  useEffect(() => {
    const from = searchParams.get(type + '-from')
    const to = searchParams.get(type + '-to')
    // console.log(type,'url',from,to)

    if ((!from || !to) && (!dateRange[0] || !dateRange[1])) {
      initTimeRange()
      return
    }
    //iso -> timestamp -> string
    const fromString = convertTime(ISOToTimestamp(from), 'yyyy-mm-dd hh:mm:ss')
    const toString = convertTime(ISOToTimestamp(to), 'yyyy-mm-dd hh:mm:ss')
    if (fromString && toString && ValidDate(fromString) && ValidDate(toString)) {
      if (inputStartTime !== fromString || inputEndTime !== toString) {
        setDateRange([new Date(fromString), new Date(toString)])
        updataUrlTimeRange(fromString, toString)
      }
    } else {
      initTimeRange()
    }
  }, [searchParams])
  useUpdateEffect(() => {
    initTimeRange()
    updataUrlTimeRange()
  }, [deltaStartTime, deltaEndTime])
  const onChange = (date, dateString) => {
    console.log('onChange')
    // Convert antd DatePicker's value to a Date object
    const startDate = date[0] ? date[0].toDate() : null
    const endDate = date[1] ? date[1].toDate() : null

    setDateRange([startDate, endDate])
    if (startDate && endDate) {
      changeAlertDescendantDelta({
        deltaStartTime: startDate.getTime() * 1000,
        deltaEndTime: endDate.getTime() * 1000,
      })
    }
  }
  const disabledDate = (current) => {
    // 禁用范围之外的日期
    if (!current) return false

    const currentDate = current.toDate()
    const startTimeDate = new Date(startTime / 1000)
    const endTimeDate = new Date(endTime / 1000)

    return currentDate < startTimeDate || currentDate > endTimeDate
  }

  // Convert Date object to dayjs object (antd DatePicker requires dayjs)
  const dateRangeDayjs = dateRange.map(date => date ? dayjs(date) : null)

  return (
    <DatePicker.RangePicker
      showTime
      value={dateRangeDayjs}
      onChange={onChange}
      disabledDate={disabledDate}
      allowClear={false}
    />
  )
}
