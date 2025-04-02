/**
 * Copyright 2024 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

// Time conversion utility class
import {
  format,
  formatDistance,
  parseISO,
  addDays,
  addMonths,
  addYears,
  isValid,
  startOfDay,
  endOfDay,
  startOfWeek,
  endOfWeek,
  startOfMonth,
  endOfMonth,
  differenceInDays,
  differenceInHours,
  differenceInYears,
  differenceInMonths,
  differenceInMinutes,
  differenceInSeconds
} from 'date-fns'
import dayjs from 'dayjs'
import timezone from 'dayjs/plugin/timezone'
import utc from 'dayjs/plugin/utc'

dayjs.extend(utc)
dayjs.extend(timezone)

/**
 * Convert microsecond timestamp to specified format
 * @param {number} timeInMicros - Input microsecond timestamp
 * @param {string} format - Target format, supports: 'us', 'ms', 's', 'm', 'h', 'yyyy-mm-dd hh:mm:ss','yyyy-mm-dd hh:mm:ss.SSS'
 * @param {number} [precision=0] - Decimal precision, applies to 'ms', 's', 'm', 'h'
 * @returns {null|string|number} Converted time
 */
export function convertTime(timeInMicros, format, precision = 0) {
  if (timeInMicros === null) {
    return null
  }
  const date = new Date(timeInMicros / 1000) // Convert to millisecond timestamp

  switch (format) {
    case 'us':
      return parseFloat(timeInMicros.toFixed(precision)) // Return microseconds
    case 'ms':
      return parseFloat((timeInMicros / 1000).toFixed(precision)) // Convert to milliseconds
    case 's':
      return parseFloat((timeInMicros / 1000000).toFixed(precision)) // Convert to seconds
    case 'm':
      return parseFloat((timeInMicros / (1000000 * 60)).toFixed(precision)) // Convert to minutes
    case 'h':
      return parseFloat((timeInMicros / (1000000 * 60 * 60)).toFixed(precision)) // Convert to hours
    case 'HH:MM':
      const HH = String(date.getHours()).padStart(2, '0')
      const MM = String(date.getMinutes()).padStart(2, '0')
      return `${HH}:${MM}`
    case 'yyyy-mm-dd hh:mm:ss':
      const yyyy = date.getFullYear()
      const mm = String(date.getMonth() + 1).padStart(2, '0') // Months are 0-based
      const dd = String(date.getDate()).padStart(2, '0')
      const hh = String(date.getHours()).padStart(2, '0')
      const min = String(date.getMinutes()).padStart(2, '0')
      const ss = String(date.getSeconds()).padStart(2, '0')
      return `${yyyy}-${mm}-${dd} ${hh}:${min}:${ss}`
    case 'yyyy-mm-dd hh:mm:ss.SSS':
      const year = date.getFullYear()
      const month = (date.getMonth() + 1).toString().padStart(2, '0')
      const day = date.getDate().toString().padStart(2, '0')
      const hour = date.getHours().toString().padStart(2, '0')
      const minute = date.getMinutes().toString().padStart(2, '0')
      const second = date.getSeconds().toString().padStart(2, '0')
      const milliseconds = date.getMilliseconds().toString().padStart(3, '0')
      return `${year}-${month}-${day} ${hour}:${minute}:${second}.${milliseconds}`
    default:
      throw new Error(`Unsupported format: ${format}`)
  }
}

/**
 * Split time range into intervals
 * @param {number} startTimestamp - Start timestamp in microseconds
 * @param {number} endTimestamp - End timestamp in microseconds
 * @returns {Array} Time segments array
 */
export function splitTimeRange(startTimestamp, endTimestamp) {
  // Convert input microsecond timestamps to milliseconds
  let startTime = new Date(startTimestamp / 1000)
  let endTime = new Date(endTimestamp / 1000)

  // Adjust to full minutes if input is not aligned
  if (startTime.getSeconds() !== 0 || startTime.getMilliseconds() !== 0) {
    startTime.setSeconds(0, 0)
  }

  if (endTime.getSeconds() !== 0 || endTime.getMilliseconds() !== 0) {
    endTime.setSeconds(0, 0)
  }

  // Calculate total time difference in minutes
  let totalMinutes = Math.ceil((endTime - startTime) / 60000)

  // Determine number of segments (max 5)
  let segmentCount = Math.min(5, totalMinutes)

  // Calculate minutes per segment (minimum 1 minute)
  let segmentMinutes = Math.ceil(totalMinutes / segmentCount)

  let segments = []
  for (let i = 0; i < segmentCount; i++) {
    let segmentStart = new Date(startTime.getTime() + i * segmentMinutes * 60000)
    let segmentEnd = new Date(segmentStart.getTime() + segmentMinutes * 60000)

    // Ensure last segment doesn't exceed endTime
    if (segmentEnd > endTime) {
      segmentEnd = endTime
    }

    segments.push({
      start: segmentStart.getTime() * 1000,
      end: segmentEnd.getTime() * 1000,
    })

    // Break loop if reached end time
    if (segmentEnd >= endTime) {
      break
    }
  }

  return segments
}

/**
 * Convert microsecond timestamp to ISO format
 * @param {number} timestamp - Input microsecond timestamp
 * @returns {string} ISO formatted string
 */
export function TimestampToISO(timestamp) {
  const milliseconds = Math.floor(timestamp / 1000) // Convert to milliseconds
  const date = new Date(milliseconds)
  const microseconds = timestamp % 1000 // Extract microsecond part
  const isoString = date.toISOString().replace('Z', '') // Remove trailing 'Z'
  const microPart = String(microseconds).padStart(3, '0') // 3-digit microsecond
  return `${isoString}${microPart}Z`
}

/**
 * Convert ISO string to microsecond timestamp
 * @param {string} isoString - ISO formatted string
 * @returns {number|null} Microsecond timestamp
 */
export function ISOToTimestamp(isoString) {
  const iso8601Regex = /^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}\.\d{6}Z$/
  // Validate format
  if (!iso8601Regex.test(isoString)) {
    return null
  }
  const datePart = isoString.slice(0, -4) + 'Z' // Truncate microseconds
  const date = new Date(datePart)
  // Check whether is a Invalid Date
  if (isNaN(date.getTime())) {
    return null
  }
  const milliTimestamp = date.getTime()
  const microPart = isoString.slice(-4, -1) // Extract microseconds
  const microseconds = parseInt(microPart, 10)
  return milliTimestamp * 1000 + microseconds
}

/**
 * Validate date string format (yyyy-mm-dd hh:mm:ss)
 * @param {string} dateString - Input date string
 * @returns {boolean} Validation result
 */
export function ValidDate(dateString) {
  // Try to parse string
  const date = new Date(dateString)
  // Check whether is a Invalid Date
  if (isNaN(date.getTime())) {
    return false
  }
  return dateString === format(date.getTime(), 'yyyy-MM-dd HH:mm:ss')
}

/**
 * Convert date string to ISO format
 * @param {string} dateString - Input date string (yyyy-MM-dd HH:mm:ss)
 * @returns {string} ISO formatted string
 */
export function DateToISO(dateString) {
  // Try to parse string
  const date = new Date(dateString)
  // Check whether is a Invalid Date
  if (isNaN(date.getTime())) {
    return null
  }
  return TimestampToISO(date.getTime() * 1000)
}

/**
 * Format time duration with unit conversion
 * @param {number} time - Time value in milliseconds
 * @param {number} [reserve=2] - Decimal places
 * @returns {string|number} Formatted value with unit
 */
export const formatTime = (time, reserve = 2) => {
  let flag = false
  if (time === undefined) {
    return 0
  } else {
    flag = time < 0
    time = Math.abs(time)
    if (time > 3600000) {
      time = parseFloat((time / 3600000).toFixed(reserve)) + 'h'
    } else if (time >= 60000) {
      time = parseFloat((time / 60000).toFixed(reserve)) + 'min'
    } else if (time >= 1000) {
      time = parseFloat((time / 1000).toFixed(reserve)) + 's'
    } else if (time >= 10) {
      time = parseFloat(time.toFixed(reserve)) + 'ms'
    } else if (time > 0) {
      time = parseFloat(time.toFixed(reserve)) + 'ms'
    } else {
      time = parseFloat(time.toFixed(reserve))
    }
    return flag ? '-' + time : time
  }
}

export function convertUTCToLocal(utcTimeString) {
  // Convert the UTC time string to a Date object
  const utcDate = new Date(utcTimeString)

  // Get the local UTC time
  const localDate = new Date(utcDate.toLocaleString('en-US', { timeZone: Intl.DateTimeFormat().resolvedOptions().timeZone }))

  // Format as YYYY-MM-DD HH:mm:ss
  const year = localDate.getFullYear()
  const month = String(localDate.getMonth() + 1).padStart(2, '0')
  const day = String(localDate.getDate()).padStart(2, '0')
  const hours = String(localDate.getHours()).padStart(2, '0')
  const minutes = String(localDate.getMinutes()).padStart(2, '0')
  const seconds = String(localDate.getSeconds()).padStart(2, '0')

  return `${year}-${month}-${day} ${hours}:${minutes}:${seconds}`
}

/**
 * 转化字节数到对应的量级(TB,GB,MB,KB)
 * @param {number} y 默认是B
 */
export const formatKMBT = (y, reserve = 2) => {
  const yy = Math.abs(y)
  if (yy >= Math.pow(1024, 4)) {
    return y < 0
      ? -1 * +(yy / Math.pow(1024, 4)).toFixed(reserve) + 'T'
      : (yy / Math.pow(1024, 4)).toFixed(reserve) + 'T'
  } else if (yy >= Math.pow(1024, 3)) {
    return y < 0
      ? -1 * +(yy / Math.pow(1024, 3)).toFixed(reserve) + 'G'
      : (yy / Math.pow(1024, 3)).toFixed(reserve) + 'G'
  } else if (yy >= Math.pow(1024, 2)) {
    return y < 0
      ? -1 * +(yy / Math.pow(1024, 2)).toFixed(reserve) + 'M'
      : (yy / Math.pow(1024, 2)).toFixed(reserve) + 'M'
  } else if (yy >= 1024) {
    return y < 0 ? -1 * +(yy / 1024).toFixed(reserve) + 'K' : (yy / 1024).toFixed(reserve) + 'K'
  } else if (yy < 1024 && yy >= 1) {
    return y < 0 ? -1 * +yy.toFixed(reserve) + 'B' : yy.toFixed(reserve) + 'B'
  } else if (yy < 1 && yy > 0) {
    return y < 0 ? -1 * +yy.toFixed(reserve) + 'B' : yy.toFixed(reserve) + 'B'
  } else if (yy === 0) {
    return 0
  } else {
    return yy
  }
}
/**
 * 转化字节数到对应的量级(TB,GB,MB)
 * @param {number} y 默认是MB
 */
export const formatMGT = (y) => {
  const yy = Math.abs(y)
  if (yy >= 1024 * 1024) {
    return y < 0
      ? -1 * +(yy / (1024 * 1024)).toFixed(2) + 'T'
      : (yy / (1024 * 1024)).toFixed(2) + 'T'
  } else if (yy >= 1024) {
    return y < 0 ? -1 * +(yy / 1024).toFixed(2) + 'G' : (yy / 1024).toFixed(2) + 'G'
  } else if (yy < 1024 && yy >= 1) {
    return y < 0 ? -1 * +yy.toFixed(2) + 'M' : yy.toFixed(2) + 'M'
  } else if (yy < 1 && yy > 0) {
    return y < 0 ? -1 * yy + 'M' : yy + 'M'
  } else if (yy === 0) {
    return 0
  } else {
    return yy + 'M'
  }
}

export const formatCount = (y, reserve = 2) => {
  const WAN = 10000
  const YI = 100000000
  const ZHAO = 1000000000000
  const yy = Math.abs(y)
  if (yy < WAN) {
    return yy
  }
  if (yy < YI) {
    return `${(yy / WAN).toFixed(reserve)}${decodeURIComponent('%E4%B8%87')}` // 万
  }
  if (yy < ZHAO) {
    return `${(yy / YI).toFixed(reserve)}${decodeURIComponent('%E4%BA%BF')}` // 亿
  }
  return `${(yy / ZHAO).toFixed(reserve)}${decodeURIComponent('%E5%85%86')}` // 兆
}
//小数转化百分比
export const formatPercent = (number, reserve = 2) => {
  if (number !== 0 && !number) {
    return '-'
  }
  if (number === 0) {
    return 0
  } else {
    if (number >= 0.01) {
      return `${parseFloat(number + '').toFixed(reserve)}%`
    } else {
      return '<0.01%'
    }
  }
}
// export type IUnit =
//   | 'byteMin'
//   | 'byteMins'
//   | 'byte'
//   | 'KB'
//   | 'KB/S'
//   | 'ms'
//   | 'us'
//   | 'ns'
//   | 'count'
//   | '%'
//   | '100%'
//   | '核'
//   | '2'
//   | 'date'
//   | '个'
//   | '';
export const formatUnit = (data, unit, reserve = 2) => {
  if (data === -1 || data === 'NaN') {
    return 'N/A'
  }
  if (data === undefined || data === null) {
    return '--'
  }
  if (data === '0' || data === 0) {
    return 0
  }
  switch (unit) {
    case 'byteMin': //后台返回B
      return formatKMBT(data, reserve)
    case 'byteMins': //后台返回B
      return formatKMBT(data) ? formatKMBT(data).toString() + '/s' : 0
    case 'byte': //后台返回单位MB
      return formatMGT(data)
    case 'KB': //后台返回单位KB
      return formatKMBT(data * 1024)
    case 'KB/S':
      return formatKMBT(data * 1024).toString() + '/S'
    case 'ms': //后台返回ms
      return formatTime(data, reserve)
    case 'us': //后台返回us
      return formatTime(data / 1000, reserve)
    case 'ns': //后台返回ns
      return formatTime(data / 1000000, reserve)
    case 'date':
      return dayjs(data).format('YYYY-MM-DD HH:mm:ss')
    case 'count': //后台返回数量
      return formatCount(data, reserve)
    case '%': //百分数加单位
      return formatPercent(data, reserve)
    case '100%': //小数转换百分数
      return formatPercent(data * 100, reserve)
    case '核': {
      const v = parseFloat(data)
      if (v > 0 && v < 0.01) {
        if (v * 1000 < 0.01) {
          return '<0.01微核'
        } else {
          return (v * 1000).toFixed(2) + '微核'
        }
      }
      return `${v.toFixed(2)}核`
    }
    case '2':
      return `${parseFloat(data).toFixed(2)}`
    case '个':
      return `${data}个`
    default:
      return `${data}`
  }
}

/**
 * Time utility class encapsulation
 */
export const timeUtils = {
  /**
   * Format date
   * @param {Date|number|string} date - Date object, timestamp or date string
   * @param {string} formatStr - Format string, e.g. 'yyyy-MM-dd HH:mm:ss'
   * @returns {string} Formatted date string
   */
  format: (date, formatStr = 'yyyy-MM-dd HH:mm:ss') => {
    if (!date) return ''
    try {
      // Handle various input types
      let dateObj = date
      if (typeof date === 'string') {
        dateObj = new Date(date)
      } else if (typeof date === 'number') {
        dateObj = new Date(date)
      }

      if (!isValid(dateObj)) return ''
      return format(dateObj, formatStr)
    } catch (error) {
      console.error('Date formatting error:', error)
      return ''
    }
  },

  /**
   * Parse ISO format date string
   * @param {string} isoStr - ISO format date string
   * @returns {Date} Date object
   */
  parseISO: (isoStr) => {
    try {
      return parseISO(isoStr)
    } catch (error) {
      console.error('ISO date parsing error:', error)
      return new Date()
    }
  },

  /**
   * Get relative time description
   * @param {Date|number|string} date - Date object, timestamp or date string
   * @param {Date} baseDate - Base date, defaults to current time
   * @returns {string} Relative time description, e.g. "3 days ago"
   */
  fromNow: (date, baseDate = new Date()) => {
    if (!date) return ''
    try {
      let dateObj = date
      if (typeof date === 'string') {
        dateObj = new Date(date)
      } else if (typeof date === 'number') {
        dateObj = new Date(date)
      }

      if (!isValid(dateObj)) return ''
      return formatDistance(dateObj, baseDate, { addSuffix: true })
    } catch (error) {
      console.error('Relative time calculation error:', error)
      return ''
    }
  },

  /**
   * Date addition/subtraction operations
   * @param {Date|number|string} date - Date object, timestamp or date string
   * @param {number} amount - Quantity to add/subtract
   * @param {string} unit - Unit: 'day', 'month', 'year'
   * @returns {Date} New date object
   */
  add: (date, amount, unit = 'day') => {
    if (!date) return new Date()
    try {
      let dateObj = date
      if (typeof date === 'string') {
        dateObj = new Date(date)
      } else if (typeof date === 'number') {
        dateObj = new Date(date)
      }

      if (!isValid(dateObj)) return new Date()

      switch (unit) {
        case 'day':
          return addDays(dateObj, amount)
        case 'month':
          return addMonths(dateObj, amount)
        case 'year':
          return addYears(dateObj, amount)
        default:
          return dateObj instanceof Date ? dateObj : new Date(dateObj)
      }
    } catch (error) {
      console.error('Date arithmetic operation error:', error)
      return new Date()
    }
  },

  /**
   * Check if date is valid
   * @param {Date|number|string} date - Date object, timestamp or date string
   * @returns {boolean} Whether the date is valid
   */
  isValid: (date) => {
    if (!date) return false
    try {
      let dateObj = date
      if (typeof date === 'string') {
        dateObj = new Date(date)
      } else if (typeof date === 'number') {
        dateObj = new Date(date)
      }

      return isValid(dateObj)
    } catch (error) {
      return false
    }
  },

  /**
   * Get current timestamp (milliseconds)
   * @returns {number} Current timestamp
   */
  now: () => {
    return Date.now()
  },

  /**
   * Convert to local timezone time
   * @param {string} utcTimeString - UTC time string
   * @returns {string} Local timezone time string
   */
  toLocalTime: (utcTimeString) => {
    return convertUTCToLocal(utcTimeString)
  },

  /**
   * Convert microsecond timestamp to specified format
   * @param {number} microTimestamp - Microsecond timestamp
   * @param {string} formatStr - Format string, e.g. 'yyyy-MM-dd HH:mm:ss.SSS'
   * @returns {string} Formatted time string
   */
  formatMicroTimestamp: (microTimestamp, formatStr = 'yyyy-MM-dd HH:mm:ss.SSS') => {
    if (!microTimestamp) return ''
    try {
      // Convert microseconds to milliseconds
      const milliseconds = Math.floor(microTimestamp / 1000)
      const date = new Date(milliseconds)

      // Use date-fns to format with milliseconds
      let formatted = format(date, formatStr)

      // Return directly if format contains .SSS
      if (formatStr.includes('.SSS')) {
        return formatted
      }

      // Return formatted result without millisecond info
      return formatted
    } catch (error) {
      console.error('Microsecond timestamp formatting error:', error)
      return ''
    }
  },

  /**
   * Convert ISO format to microsecond timestamp
   * @param {string} isoString - ISO format time string
   * @returns {number|null} Microsecond timestamp or null if conversion fails
   */
  isoToMicroTimestamp: (isoString) => {
    return ISOToTimestamp(isoString)
  },

  /**
   * Convert microsecond timestamp to ISO format
   * @param {number} microTimestamp - Microsecond timestamp
   * @returns {string} ISO format time string
   */
  microTimestampToIso: (microTimestamp) => {
    return TimestampToISO(microTimestamp)
  },

  /**
   * Get current microsecond timestamp
   * Note: JavaScript cannot directly get microsecond precision, this is simulated
   * @returns {number} Current microsecond timestamp (milliseconds*1000 + random microseconds)
   */
  nowMicro: () => {
    const now = Date.now()
    // Simulate microsecond part (actual JS cannot get real microseconds)
    const microPart = Math.floor(Math.random() * 1000)
    return now * 1000 + microPart
  },

  /**
   * Convert to specified timezone time
   * @param {Date|number|string} date - Date object, timestamp or date string
   * @param {string} timezone - Target timezone, e.g. 'Asia/Shanghai', defaults to local timezone
   * @param {string} formatStr - Output format
   * @returns {string} Time string in specified timezone
   */
  toTimezone: (date, timezone = Intl.DateTimeFormat().resolvedOptions().timeZone, formatStr = 'YYYY-MM-DD HH:mm:ss') => {
    if (!date) return ''
    try {
      return dayjs(date).tz(timezone).format(formatStr)
    } catch (error) {
      console.error('Timezone conversion error:', error)
      return ''
    }
  },

  /**
   * Convert UTC time to local time
   * @param {string|Date} utcTime - UTC time
   * @param {string} formatStr - Output format
   * @returns {string} Local time string
   */
  utcToLocal: (utcTime, formatStr = 'YYYY-MM-DD HH:mm:ss') => {
    if (!utcTime) return ''
    try {
      return dayjs.utc(utcTime).local().format(formatStr)
    } catch (error) {
      console.error('UTC to local time conversion error:', error)
      return convertUTCToLocal(utcTime) // Fallback to original function
    }
  },

  /**
   * Convert local time to UTC time
   * @param {string|Date} localTime - Local time
   * @param {string} formatStr - Output format
   * @returns {string} UTC time string
   */
  localToUtc: (localTime, formatStr = 'YYYY-MM-DD HH:mm:ss') => {
    if (!localTime) return ''
    try {
      return dayjs(localTime).utc().format(formatStr)
    } catch (error) {
      console.error('Local to UTC conversion error:', error)
      return ''
    }
  },

  /**
   * Convert microsecond time to specified unit and format
   * @param {number} microTimestamp - Microsecond timestamp
   * @param {string} unit - Target unit: 'us', 'ms', 's', 'm', 'h'
   * @param {number} precision - Decimal precision
   * @returns {number|null} Converted time value
   */
  convertMicroTime: (microTimestamp, unit = 'ms', precision = 2) => {
    return Number(convertTime(microTimestamp, unit, precision))
  },

  /**
   * Get start time of a day
   * @param {Date|number|string} date - Date object, timestamp or date string
   * @returns {Date} Start time date object of the day
   */
  startOfDay: (date) => {
    if (!date) return startOfDay(new Date())
    try {
      return startOfDay(new Date(date))
    } catch (error) {
      console.error('Error getting day start time:', error)
      return startOfDay(new Date())
    }
  },

  /**
   * Get end time of a day
   * @param {Date|number|string} date - Date object, timestamp or date string
   * @returns {Date} End time date object of the day
   */
  endOfDay: (date) => {
    if (!date) return endOfDay(new Date())
    try {
      return endOfDay(new Date(date))
    } catch (error) {
      console.error('Error getting day end time:', error)
      return endOfDay(new Date())
    }
  },

  /**
   * Get start time of a week
   * @param {Date|number|string} date - Date object, timestamp or date string
   * @param {Object} options - Configuration options, same as date-fns startOfWeek
   * @returns {Date} Start time date object of the week
   */
  startOfWeek: (date, options) => {
    if (!date) return startOfWeek(new Date(), options)
    try {
      return startOfWeek(new Date(date), options)
    } catch (error) {
      console.error('Error getting week start time:', error)
      return startOfWeek(new Date(), options)
    }
  },

  /**
   * Get end time of a week
   * @param {Date|number|string} date - Date object, timestamp or date string
   * @param {Object} options - Configuration options, same as date-fns endOfWeek
   * @returns {Date} End time date object of the week
   */
  endOfWeek: (date, options) => {
    if (!date) return endOfWeek(new Date(), options)
    try {
      return endOfWeek(new Date(date), options)
    } catch (error) {
      console.error('Error getting week end time:', error)
      return endOfWeek(new Date(), options)
    }
  },

  /**
   * Get start time of a month
   * @param {Date|number|string} date - Date object, timestamp or date string
   * @returns {Date} Start time date object of the month
   */
  startOfMonth: (date) => {
    if (!date) return startOfMonth(new Date())
    try {
      return startOfMonth(new Date(date))
    } catch (error) {
      console.error('Error getting month start time:', error)
      return startOfMonth(new Date())
    }
  },

  /**
   * Get end time of a month
   * @param {Date|number|string} date - Date object, timestamp or date string
   * @returns {Date} End time date object of the month
   */
  endOfMonth: (date) => {
    if (!date) return endOfMonth(new Date())
    try {
      return endOfMonth(new Date(date))
    } catch (error) {
      console.error('Error getting month end time:', error)
      return endOfMonth(new Date())
    }
  },

  /**
   * Calculate time difference between two dates
   * @param {Date|number|string} dateLeft - Earlier date
   * @param {Date|number|string} dateRight - Later date
   * @param {string} unit - Time unit, options: 'years', 'months', 'days', 'hours', 'minutes', 'seconds'
   * @returns {number} Time difference
   */
  getDiff: (dateLeft, dateRight, unit = 'days') => {
    try {
      const left = new Date(dateLeft);
      const right = new Date(dateRight);

      switch(unit) {
        case 'years':
          return differenceInYears(right, left);
        case 'months':
          return differenceInMonths(right, left);
        case 'days':
          return differenceInDays(right, left);
        case 'hours':
          return differenceInHours(right, left);
        case 'minutes':
          return differenceInMinutes(right, left);
        case 'seconds':
          return differenceInSeconds(right, left);
        default:
          console.warn(`Unknown time unit: ${unit}, using default 'days'`);
          return differenceInDays(right, left);
        }
    } catch (error) {
      console.error(`Error calculating ${unit} difference:`, error);
      return 0;
    }
  },
}