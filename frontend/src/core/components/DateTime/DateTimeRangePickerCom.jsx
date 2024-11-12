//此组件用于页面内的时间选择 不具备存储功能 目前用于log和trace页面

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
import { endOfDay, format, startOfDay } from 'date-fns'
import React, { useEffect, useState } from 'react'
import { DateRange } from 'react-date-range'
import { useSearchParams } from 'react-router-dom'
import {
  GetInitalTimeRange,
  getTimestampRange,
  initialTimeRangeState,
  timeRangeList,
  timeRangeMap,
} from 'src/core/store/reducers/timeRangeReducer'
import { convertTime, DateToISO, ISOToTimestamp, TimestampToISO, ValidDate } from 'src/core/utils/time'
import './index.css'
import 'react-date-range/dist/styles.css' // main style file
import 'react-date-range/dist/theme/default.css' // theme css file
import { useSelector } from 'react-redux'
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
  const { type } = props
  const [searchParams, setSearchParams] = useSearchParams()
  const [dropdownVisible, setDropdownVisible] = useState(false)
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
    startDate: startOfDay(new Date()),
    endDate: endOfDay(new Date()),
    key: 'selection',
  })
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
    if (needChangeUrl) {
      setSearchParams(params, { replace: true })
    }
  }
  const confirmTimeRange = (event) => {
    if (startTimeInvalid || endTimeInvalid) {
      event.preventDefault()
      event.stopPropagation()
      return
    }
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
      startDate: startOfDay(state.startDate),
      endDate: endOfDay(state.endDate),
      key: state.key,
    })
    setInputStartTime(format(state.startDate, 'yyyy-MM-dd HH:mm:ss'))
    setInputEndTime(format(endOfDay(state.endDate), 'yyyy-MM-dd HH:mm:ss'))
  }

  const handleTimeRange = (key) => {
    const { startTime, endTime } = getTimestampRange(key)
    const fromString = convertTime(startTime, 'yyyy-mm-dd hh:mm:ss')
    const toString = convertTime(endTime, 'yyyy-mm-dd hh:mm:ss')
    setInputStartTime(fromString)
    setInputEndTime(toString)
    setDropdownVisible(false)
    updataUrlTimeRange(fromString, toString)
  }
  const validStartTime = () => {
    let feedback = '请输入或选择正确的日期'
    let result = true
    if (!inputStartTime || inputStartTime?.length === 0 || !ValidDate(inputStartTime)) {
      result = false
    } else if (new Date(inputStartTime) > new Date(inputEndTime)) {
      feedback = '开始时间不能晚于结束时间'
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
    let feedback = '请输入或选择正确的日期'
    let result = true
    if (!inputEndTime || inputEndTime?.length === 0 || !ValidDate(inputEndTime)) {
      result = false
    } else if (new Date(inputStartTime) > new Date(inputEndTime)) {
      feedback = '结束时间不能早于开始时间'
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
      initFromString = TimestampToISO(storeTimeRange.startTime)
      initToString = TimestampToISO(storeTimeRange.endTime)
    } else {
      const initTimeRange = GetInitalTimeRange()
      initFromString = convertTime(initTimeRange.startTime, 'yyyy-mm-dd hh:mm:ss')
      initToString = convertTime(initTimeRange.endTime, 'yyyy-mm-dd hh:mm:ss')
    }

    setInputStartTime(initFromString, 'yyyy-mm-dd hh:mm:ss')
    setInputEndTime(initToString, 'yyyy-mm-dd hh:mm:ss')
    updataUrlTimeRange(initFromString, initToString)
  }
  useEffect(() => {
    const from = searchParams.get(type + '-from')
    const to = searchParams.get(type + '-to')
    // console.log(type,'url',from,to)

    if ((!from || !to) && (!inputStartTime || !inputEndTime)) {
      initTimeRange()
      return
    }
    //iso -> timestamp -> string
    const fromString = convertTime(ISOToTimestamp(from), 'yyyy-mm-dd hh:mm:ss')
    const toString = convertTime(ISOToTimestamp(to), 'yyyy-mm-dd hh:mm:ss')
    if (fromString && toString && ValidDate(fromString) && ValidDate(toString)) {
      if (inputStartTime !== fromString || inputEndTime !== toString) {
        // console.log('url发现参数和store不符，更新精确时间')
        setInputStartTime(fromString)
        setInputEndTime(toString)
        updataUrlTimeRange(fromString, toString)
      }
    } else {
      initTimeRange()
    }
  }, [searchParams])

  useEffect(() => {
    const startTimeValid = validStartTime()
    const endTimeValid = validEndTime()
    if (startTimeValid && endTimeValid) {
      setDateRange({
        startDate: startOfDay(new Date(inputStartTime).getTime()),
        endDate: endOfDay(new Date(inputEndTime).getTime()),
        key: 'selection',
      })
    }
  }, [inputStartTime, inputEndTime])
  return (
    <CDropdown
      dark
      autoClose="outside"
      visible={dropdownVisible}
      onShow={() => setDropdownVisible(true)}
      onHide={() => setDropdownVisible(false)}
    >
      <CDropdownToggle color="dark" className="" size="sm" onClick={() => setDropdownVisible(true)}>
        <CIcon icon={cilClock} className="mr-2" />
        <span className="text-sm">
          {inputStartTime} to {inputEndTime}
          {/* {displayStartTime + ' to ' + displayEndTime} */}
        </span>
      </CDropdownToggle>
      <CDropdownMenu className="m-0 p-0">
        <div className="w-[600px] flex">
          <div className="w-3/5 border-r border-r-slate-700  px-3 py-2">
            <CForm noValidate>
              <div>绝对时间范围</div>

              <CFormLabel className="text-sm mt-2 block">开始{startTimeInvalid && 1}</CFormLabel>

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
                结束
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
              <CButton color="primary" size="sm" className="mt-3" onClick={confirmTimeRange}>
                应用时间范围
              </CButton>
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
            <div className="p-2">快速范围</div>
            <CDropdownMenu className="w-2/5 border-0 text-base">
              {Object.keys(timeRangeMap).map((key) => {
                return (
                  <CDropdownItem key={key} onClick={() => handleTimeRange(key)}>
                    {timeRangeMap[key].name}
                  </CDropdownItem>
                )
              })}
            </CDropdownMenu>
          </div>
        </div>
      </CDropdownMenu>
    </CDropdown>
  )
}
