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
  CPopover,
} from '@coreui/react'
import { endOfDay, format, startOfDay } from 'date-fns'
import React, { useState, useEffect } from 'react'
import { DateRange } from 'react-date-range'

import 'react-date-range/dist/styles.css' // main style file
import 'react-date-range/dist/theme/default.css' // theme css file
import './index.css'
import { useDispatch, useSelector } from 'react-redux'
import {
  GetInitalTimeRange,
  getTimestampRange,
  timeRangeList,
} from 'src/store/reducers/timeRangeReducer'
import { convertTime, ISOToTimestamp, TimestampToISO } from 'src/utils/time'
import { useLocation, useSearchParams } from 'react-router-dom'
import { useUpdateEffect } from 'react-use'

const DateTimeRangePicker = React.memo((props) => {
  const location = useLocation()
  const [dropdownVisible, setDropdownVisible] = useState(false)
  const [searchParams, setSearchParams] = useSearchParams()
  const storeTimeRange = useSelector((state) => state.timeRange)

  const [state, setState] = useState({
    startDate: startOfDay(new Date()),
    endDate: endOfDay(new Date()),
    key: 'selection',
  })
  // 展示用的starttime（取消快速范围，快速范围只用于获取精确数据）
  const [startTime, setStartTime] = useState(format(state.startDate, 'yyyy-MM-dd HH:mm:ss'))
  const [endTime, setEndTime] = useState(format(state.endDate, 'yyyy-MM-dd HH:mm:ss'))
  const [startTimeInvalid, setStartTimeInvalid] = useState(false)
  const [endTimeInvalid, setEndTimeInvalid] = useState(false)
  const [startTimeFeedback, setStartTimeFeedback] = useState()
  const [endTimeFeedback, setEndTimeFeedback] = useState()

  const dispatch = useDispatch()

  const setStoreTimeRange = (value) => {
    dispatch({ type: 'SET_TIMERANGE', payload: value })
  }

  const initTimeRange = () => {
    setStartTime(convertTime(storeTimeRange.startTime, 'yyyy-mm-dd hh:mm:ss'))
    setEndTime(convertTime(storeTimeRange.endTime, 'yyyy-mm-dd hh:mm:ss'))
  }
  const updataUrlTimeRange = () => {
    const params = new URLSearchParams(searchParams)
    const from = searchParams.get('from')
    const to = searchParams.get('to')

    let needChangeUrl = false
    if (storeTimeRange.startTime !== ISOToTimestamp(from)) {
      params.set('from', TimestampToISO(storeTimeRange.startTime))
      needChangeUrl = true
    }
    if (storeTimeRange.endTime !== ISOToTimestamp(to)) {
      params.set('to', TimestampToISO(storeTimeRange.endTime))
      needChangeUrl = true
    }
    if (needChangeUrl) {
      let url = new URL(window.location.href)
      if (url.hash) {
        const newUrl = `${url.origin}/${url.hash.split('?')[0]}?${params.toString()}`
        window.history.replaceState(null, '', newUrl)
      }
    }
  }

  useEffect(() => {
    const from = searchParams.get('from')
    const to = searchParams.get('to')
    if (!from || !to) {
      updataUrlTimeRange()
      return
    }

    const fromTimestamp = ISOToTimestamp(from)
    const toTimestamp = ISOToTimestamp(to)
    if (fromTimestamp && toTimestamp) {
      if (storeTimeRange.startTime !== fromTimestamp || storeTimeRange.endTime !== toTimestamp) {
        // console.log('url发现参数和store不符，更新精确时间')
        setStoreTimeRange({
          rangeType: null,
          startTime: fromTimestamp,
          endTime: toTimestamp,
        })
      }
    } else {
      const initTimeRange = GetInitalTimeRange()
      setStoreTimeRange(initTimeRange)
    }
  }, [searchParams, location])
  // 打开下拉面板初始化该组件的数据
  useEffect(() => {
    if (dropdownVisible) {
      initTimeRange()
    }
  }, [dropdownVisible])
  // 存储数据变了初始化该组件的数据

  useUpdateEffect(() => {
    initTimeRange()
    updataUrlTimeRange()
  }, [storeTimeRange])
  // 选择的快速范围变了 修改日期选择器的时间

  // 输入框的日期时间变了 修改日期选择器的时间
  useEffect(() => {
    const startTimeValid = validStartTime()
    const endTimeValid = validEndTime()
    if (startTimeValid && endTimeValid) {
      setState({
        startDate: startOfDay(new Date(startTime).getTime()),
        endDate: endOfDay(new Date(endTime).getTime()),
        key: 'selection',
      })
    }
  }, [startTime, endTime])
  const handleSelect = (ranges) => {
    let state = ranges.selection
    setState({
      startDate: startOfDay(state.startDate),
      endDate: endOfDay(state.endDate),
      key: state.key,
    })
    setStartTime(format(state.startDate, 'yyyy-MM-dd HH:mm:ss'))
    setEndTime(format(endOfDay(state.endDate), 'yyyy-MM-dd HH:mm:ss'))
  }
  const changeStartTime = (event) => {
    setStartTime(event.target.value)
  }
  const changeEndTime = (event) => {
    setEndTime(event.target.value)
  }
  const selectTimeRange = (item) => {
    const { startTime, endTime } = getTimestampRange(item.rangeType)
    setStartTime(convertTime(startTime, 'yyyy-mm-dd hh:mm:ss'))
    setEndTime(convertTime(endTime, 'yyyy-mm-dd hh:mm:ss'))
    setDropdownVisible(false)
    setStoreTimeRange({
      rangeType: null,
      startTime: startTime,
      endTime: endTime,
    })
  }
  const confirmTimeRange = (event) => {
    if (startTimeInvalid || endTimeInvalid) {
      event.preventDefault()
      event.stopPropagation()
      return
    }
    setStoreTimeRange({
      rangeType: null,
      startTime: new Date(startTime).getTime() * 1000,
      endTime: new Date(endTime).getTime() * 1000,
    })
    setDropdownVisible(false)
  }
  function isValidDate(dateString) {
    // 尝试解析字符串
    const date = new Date(dateString)
    // 检查是否为 Invalid Date
    if (isNaN(date.getTime())) {
      return false
    }

    return dateString === format(date.getTime(), 'yyyy-MM-dd HH:mm:ss')
  }
  const validStartTime = () => {
    let feedback = '请输入或选择正确的日期'
    let result = true
    if (!startTime || startTime.length === 0 || !isValidDate(startTime)) {
      result = false
    } else if (new Date(startTime) > new Date(endTime)) {
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
    if (!endTime || endTime.length === 0 || !isValidDate(endTime)) {
      result = false
    } else if (new Date(startTime) > new Date(endTime)) {
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

  return (
    <>
      <CDropdown
        dark
        autoClose="outside"
        visible={dropdownVisible}
        onShow={() => setDropdownVisible(true)}
        onHide={() => setDropdownVisible(false)}
      >
        <CDropdownToggle
          color="dark"
          className=""
          size="sm"
          onClick={() => setDropdownVisible(true)}
        >
          <CIcon icon={cilClock} className="mr-2" />
          <span className="text-sm">
            {convertTime(storeTimeRange?.startTime, 'yyyy-mm-dd hh:mm:ss')} to{' '}
            {convertTime(storeTimeRange?.endTime, 'yyyy-mm-dd hh:mm:ss')}
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
                    value={startTime}
                    type="text"
                    onChange={changeStartTime}
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
                    value={endTime}
                    onChange={changeEndTime}
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
                ranges={[state]}
                onChange={handleSelect}
                editableDateInputs={true}
                hh
                showDateDisplay={false}
              />
            </div>
            <div className="w-2/5">
              <div className="p-2">快速范围</div>
              <CDropdownMenu className="w-2/5 border-0">
                {timeRangeList.map((item) => {
                  return (
                    <CDropdownItem key={item.rangeType} onClick={() => selectTimeRange(item)}>
                      {item.name}
                    </CDropdownItem>
                  )
                })}
              </CDropdownMenu>
            </div>
          </div>
        </CDropdownMenu>
      </CDropdown>
    </>
  )
})

export default DateTimeRangePicker
