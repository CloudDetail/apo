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
import {
  GetInitalTimeRange,
  getTimestampRange,
  timeRangeMap,
} from 'src/core/store/reducers/timeRangeReducer'
import { convertTime, timeUtils } from 'src/core/utils/time'
import 'react-date-range/dist/styles.css' // main style file
import 'react-date-range/dist/theme/default.css' // theme css file
import { useSelector } from 'react-redux'
import { useTranslation } from 'react-i18next'
import { Button, Menu } from 'antd'

export default function DateTimeRangePickerCom(props) {
  const { t } = useTranslation('core/dateTime')
  const { defaultTimerange = null, readOnly = false, changeTimeRange = null } = props
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
    startDate: timeUtils.startOfDay(new Date()),
    endDate: timeUtils.endOfDay(new Date()),
    key: 'selection',
  })
  const [rangeTypeKey, setRangeTypeKey] = useState(Object.keys(timeRangeMap)[0])

  const confirmTimeRange = (event) => {
    if (startTimeInvalid || endTimeInvalid) {
      event.preventDefault()
      event.stopPropagation()
      return
    }
    const startTime = new Date(inputStartTime).getTime() * 1000
    const endTime = new Date(inputEndTime).getTime() * 1000
    changeTimeRange?.({
      startTime,
      endTime,
      rangeType: null,
    })
    setDropdownVisible(false)
  }

  const changeInputStartTime = (event) => {
    setInputStartTime(event.target.value)
  }
  const changeInputEndTime = (event) => {
    setInputEndTime(event.target.value)
  }
  const handleDateRange = (ranges) => {
    setRangeTypeKey(null)
    let state = ranges.selection
    setDateRange({
      startDate: timeUtils.startOfDay(state.startDate),
      endDate: timeUtils.endOfDay(state.endDate),
      key: state.key,
    })
    setInputStartTime(timeUtils.format(state.startDate, 'yyyy-MM-dd HH:mm:ss'))
    if (areDatesEqual(state.endDate, new Date())) {
      setInputEndTime(timeUtils.format(new Date(), 'yyyy-MM-dd HH:mm:ss'))
    } else {
      setInputEndTime(timeUtils.format(timeUtils.endOfDay(state.endDate), 'yyyy-MM-dd HH:mm:ss'))
    }
  }
  const areDatesEqual = (date1, date2) => {
    return (
      date1.getFullYear() === date2.getFullYear() &&
      date1.getMonth() === date2.getMonth() &&
      date1.getDate() === date2.getDate()
    )
  }
  const handleTimeRange = (key) => {
    const { startTime, endTime } = getTimestampRange(key)
    const fromString = convertTime(startTime, 'yyyy-mm-dd hh:mm:ss')
    const toString = convertTime(endTime, 'yyyy-mm-dd hh:mm:ss')
    setInputStartTime(fromString)
    setInputEndTime(toString)
    setDropdownVisible(false)
    const newStartTime = new Date(fromString).getTime() * 1000
    const newEndTime = new Date(toString).getTime() * 1000
    changeTimeRange?.({
      startTime: newStartTime,
      endTime: newEndTime,
      rangeType: key,
    })
    setRangeTypeKey(key)
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
    } else if (new Date(inputEndTime) > new Date()) {
      feedback = t('dateTimeRangePicker.endTimeLongerThanCurrentTimeFeedback')
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

    const newStartTime = new Date(initFromString).getTime() * 1000
    const newEndTime = new Date(initToString).getTime() * 1000
    changeTimeRange?.({
      startTime: newStartTime,
      endTime: newEndTime,
      rangeType: null,
    })
  }

  useEffect(() => {
    if (defaultTimerange) {
      if (defaultTimerange.rangeType) {
        setRangeTypeKey(defaultTimerange.rangeType)
        const { startTime, endTime } = getTimestampRange(defaultTimerange.rangeType)
        const fromString = convertTime(startTime, 'yyyy-mm-dd hh:mm:ss')
        const toString = convertTime(endTime, 'yyyy-mm-dd hh:mm:ss')
        setInputStartTime(fromString)
        setInputEndTime(toString)
      } else {
        const fromString = convertTime(defaultTimerange.startTime, 'yyyy-mm-dd hh:mm:ss')
        const toString = convertTime(defaultTimerange.endTime, 'yyyy-mm-dd hh:mm:ss')
        setInputStartTime(fromString)
        setInputEndTime(toString)
      }
    } else {
      initTimeRange()
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
      <CDropdownToggle className="" size="sm" onClick={() => !readOnly && setDropdownVisible(true)}>
        <CIcon icon={cilClock} className="mr-2" />
        <span className="text-sm">
          {rangeTypeKey ? timeRangeMap[rangeTypeKey].name : inputStartTime + ' to ' + inputEndTime}
          {/* {inputStartTime} to {inputEndTime} */}
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
                  max={timeUtils.format(new Date(), "yyyy-MM-dd'T'HH:mm")}
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
                  max={timeUtils.format(new Date(), "yyyy-MM-dd'T'HH:mm")}
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
              showSelectionPreview={true}
              ranges={[dateRange]}
              onChange={handleDateRange}
              editableDateInputs={true}
              hh
              showDateDisplay={false}
              maxDate={new Date()}
            />
          </div>
          <div className="w-2/5">
            <div className="p-2">{t('dateTimeRangePicker.quickRangeTitle')}</div>
            <Menu selectedKeys={[rangeTypeKey]} onClick={({ key }) => handleTimeRange(key)}>
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
