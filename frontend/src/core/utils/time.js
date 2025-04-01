/**
 * Copyright 2024 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

// 时间转换工具类
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
 * 将微秒级时间转换为指定格式
 * @param {number} timeInMicros - 输入的微秒级时间戳
 * @param {string} format - 转换的格式，支持 'us', 'ms', 's', 'm', 'h', 'yyyy-mm-dd hh:mm:ss','yyyy-mm-dd hh:mm:ss.SSS'
 * @param {number} [precision=0] - 精度，小数点后的位数，仅适用于 'ms', 's', 'm', 'h'
 * @returns {null|string|number} 转换后的时间
 */
export function convertTime(timeInMicros, format, precision = 0) {
  if (timeInMicros === null) {
    return null
  }
  const date = new Date(timeInMicros / 1000) // 转换为毫秒级时间戳

  switch (format) {
    case 'us':
      return parseFloat(timeInMicros.toFixed(precision)) // 返回微秒
    case 'ms':
      return parseFloat((timeInMicros / 1000).toFixed(precision)) // 转换为毫秒
    case 's':
      return parseFloat((timeInMicros / 1000000).toFixed(precision)) // 转换为秒
    case 'm':
      return parseFloat((timeInMicros / (1000000 * 60)).toFixed(precision)) // 转换为分钟
    case 'h':
      return parseFloat((timeInMicros / (1000000 * 60 * 60)).toFixed(precision)) // 转换为小时
    case 'HH:MM':
      const HH = String(date.getHours()).padStart(2, '0')
      const MM = String(date.getMinutes()).padStart(2, '0')
      return `${HH}:${MM}`
    case 'yyyy-mm-dd hh:mm:ss':
      const yyyy = date.getFullYear()
      const mm = String(date.getMonth() + 1).padStart(2, '0') // 月份从0开始
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

export function splitTimeRange(startTimestamp, endTimestamp) {
  // 将输入的微秒级时间戳转换为毫秒级时间戳
  let startTime = new Date(startTimestamp / 1000)
  let endTime = new Date(endTimestamp / 1000)

  // 如果输入时间段不符合整分钟，调整为整分钟
  if (startTime.getSeconds() !== 0 || startTime.getMilliseconds() !== 0) {
    startTime.setSeconds(0, 0)
  }

  if (endTime.getSeconds() !== 0 || endTime.getMilliseconds() !== 0) {
    endTime.setSeconds(0, 0)
  }

  // 计算总时间差，单位为分钟
  let totalMinutes = Math.ceil((endTime - startTime) / 60000)

  // 确定分段数
  let segmentCount = Math.min(5, totalMinutes)

  // 计算每段的分钟数，至少为1分钟
  let segmentMinutes = Math.ceil(totalMinutes / segmentCount)

  let segments = []
  for (let i = 0; i < segmentCount; i++) {
    let segmentStart = new Date(startTime.getTime() + i * segmentMinutes * 60000)
    let segmentEnd = new Date(segmentStart.getTime() + segmentMinutes * 60000)

    // 确保最后一个时间段结束时间不超过endTime
    if (segmentEnd > endTime) {
      segmentEnd = endTime
    }

    segments.push({
      start: segmentStart.getTime() * 1000,
      end: segmentEnd.getTime() * 1000,
    })

    // 终止条件，防止时间段溢出
    if (segmentEnd >= endTime) {
      break
    }
  }

  return segments
}

/**
 * 微秒时间戳转ISO
 * @param {number} timestamp - 输入的微秒级时间戳
 * @returns {string} 转换后的时间
 */
export function TimestampToISO(timestamp) {
  const milliseconds = Math.floor(timestamp / 1000) // 转换为毫秒
  const date = new Date(milliseconds) // 使用毫秒创建日期对象
  const microseconds = timestamp % 1000 // 取微秒部分
  const isoString = date.toISOString().replace('Z', '') // 移除末尾的 'Z'
  const microPart = String(microseconds).padStart(3, '0') // 保证微秒部分是3位
  return `${isoString}${microPart}Z`
}

/**
 * ISO转微秒时间戳
 * @param {string} isoString - 输入的ISO
 * @returns {number|null} 转换后的微秒级时间戳
 */
export function ISOToTimestamp(isoString) {
  const iso8601Regex = /^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}\.\d{6}Z$/
  // 验证格式
  if (!iso8601Regex.test(isoString)) {
    return null
  }
  const datePart = isoString.slice(0, -4) + 'Z' // 去掉微秒部分后转换为日期
  const date = new Date(datePart)
  // 检查日期是否有效
  if (isNaN(date.getTime())) {
    return null
  }
  const milliTimestamp = date.getTime() // 获取毫秒级时间戳
  const microPart = isoString.slice(-4, -1) // 提取微秒部分
  const microseconds = parseInt(microPart, 10) // 转换为整数
  const microsecondTimestamp = milliTimestamp * 1000 + microseconds // 计算微秒级时间戳

  return microsecondTimestamp
}

/**
 * 判断字符串类型是否满足 yyyy-mm-dd hh:mm:ss 格式 满足返回true，反之false
 * @param {string} dateString - 输入的字符串类型时间
 * @returns {boolean} 是否合规
 */
export function ValidDate(dateString) {
  // 尝试解析字符串
  const date = new Date(dateString)
  // 检查是否为 Invalid Date
  if (isNaN(date.getTime())) {
    return false
  }
  return dateString === format(date.getTime(), 'yyyy-MM-dd HH:mm:ss')
}

/**
 * yyyy-MM-dd HH:mm:ss 转ISO
 * @param {string} dateString - 输入的字符串类型时间
 * @returns {string} 是否合规
 */
export function DateToISO(dateString) {
  // 尝试解析字符串
  const date = new Date(dateString)
  // 检查是否为 Invalid Date
  if (isNaN(date.getTime())) {
    return null
  }
  return TimestampToISO(date.getTime() * 1000)
}

/**
 * [时间单位转换-毫秒]
 * 默认 ms
 * @param   {number}  time [time value]
 * @return  {string|number}       [time value with unit]
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
 * 时间工具类封装
 */
export const timeUtils = {
  /**
   * 格式化日期
   * @param {Date|number|string} date - 日期对象、时间戳或日期字符串
   * @param {string} formatStr - 格式化字符串，例如 'yyyy-MM-dd HH:mm:ss'
   * @returns {string} 格式化后的日期字符串
   */
  format: (date, formatStr = 'yyyy-MM-dd HH:mm:ss') => {
    if (!date) return ''
    try {
      // 处理各种可能的输入类型
      let dateObj = date
      if (typeof date === 'string') {
        dateObj = new Date(date)
      } else if (typeof date === 'number') {
        dateObj = new Date(date)
      }

      if (!isValid(dateObj)) return ''
      return format(dateObj, formatStr)
    } catch (error) {
      console.error('日期格式化错误:', error)
      return ''
    }
  },

  /**
   * 解析ISO格式日期字符串
   * @param {string} isoStr - ISO格式日期字符串
   * @returns {Date} 日期对象
   */
  parseISO: (isoStr) => {
    try {
      return parseISO(isoStr)
    } catch (error) {
      console.error('ISO日期解析错误:', error)
      return new Date()
    }
  },

  /**
   * 获取相对时间描述
   * @param {Date|number|string} date - 日期对象、时间戳或日期字符串
   * @param {Date} baseDate - 基准日期，默认为当前时间
   * @returns {string} 相对时间描述，如"3天前"
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
      console.error('相对时间计算错误:', error)
      return ''
    }
  },

  /**
   * 日期加减操作
   * @param {Date|number|string} date - 日期对象、时间戳或日期字符串
   * @param {number} amount - 要加减的数量
   * @param {string} unit - 单位：'day', 'month', 'year'
   * @returns {Date} 新的日期对象
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
      console.error('日期加减操作错误:', error)
      return new Date()
    }
  },

  /**
   * 判断日期是否有效
   * @param {Date|number|string} date - 日期对象、时间戳或日期字符串
   * @returns {boolean} 是否为有效日期
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
   * 获取当前时间的时间戳（毫秒）
   * @returns {number} 当前时间戳
   */
  now: () => {
    return Date.now()
  },

  /**
   * 转换为本地时区时间
   * @param {string} utcTimeString - UTC时间字符串
   * @returns {string} 本地时区时间字符串
   */
  toLocalTime: (utcTimeString) => {
    return convertUTCToLocal(utcTimeString)
  },
  /**
   * 微秒级时间戳转换为指定格式
   * @param {number} microTimestamp - 微秒级时间戳
   * @param {string} formatStr - 格式化字符串，例如 'yyyy-MM-dd HH:mm:ss.SSS'
   * @returns {string} 格式化后的时间字符串
   */
  formatMicroTimestamp: (microTimestamp, formatStr = 'yyyy-MM-dd HH:mm:ss.SSS') => {
    if (!microTimestamp) return ''
    try {
      // 将微秒转换为毫秒
      const milliseconds = Math.floor(microTimestamp / 1000)
      const date = new Date(milliseconds)
      
      // 处理微秒部分
      const microsecondPart = String(microTimestamp % 1000).padStart(3, '0')
      
      // 先用date-fns格式化到毫秒
      let formatted = format(date, formatStr)
      
      // 如果格式中包含微秒占位符，则替换
      if (formatStr.includes('.SSS')) {
        // 已经包含毫秒，我们可以在实际应用中扩展为微秒如果需要
        return formatted
      } else if (formatStr.includes('SSS')) {
        // 直接包含毫秒占位符
        return formatted
      } else {
        // 不需要微秒信息
        return formatted
      }
    } catch (error) {
      console.error('微秒时间戳格式化错误:', error)
      return ''
    }
  },
  
  /**
   * 将ISO格式转换为微秒级时间戳
   * @param {string} isoString - ISO格式时间字符串
   * @returns {number|null} 微秒级时间戳或null（如果转换失败）
   */
  isoToMicroTimestamp: (isoString) => {
    return ISOToTimestamp(isoString)
  },
  
  /**
   * 将微秒级时间戳转换为ISO格式
   * @param {number} microTimestamp - 微秒级时间戳
   * @returns {string} ISO格式时间字符串
   */
  microTimestampToIso: (microTimestamp) => {
    return TimestampToISO(microTimestamp)
  },
  
  /**
   * 获取当前时间的微秒级时间戳
   * 注意：JavaScript无法直接获取微秒精度，这里是模拟的
   * @returns {number} 当前时间的微秒级时间戳（毫秒*1000+随机微秒）
   */
  nowMicro: () => {
    const now = Date.now()
    // 模拟微秒部分（实际上JS无法获取真实微秒）
    const microPart = Math.floor(Math.random() * 1000)
    return now * 1000 + microPart
  },
  
  /**
   * 转换为指定时区的时间
   * @param {Date|number|string} date - 日期对象、时间戳或日期字符串
   * @param {string} timezone - 目标时区，例如 'Asia/Shanghai'，默认为本地时区
   * @param {string} formatStr - 输出格式
   * @returns {string} 指定时区的时间字符串
   */
  toTimezone: (date, timezone = Intl.DateTimeFormat().resolvedOptions().timeZone, formatStr = 'YYYY-MM-DD HH:mm:ss') => {
    if (!date) return ''
    try {
      return dayjs(date).tz(timezone).format(formatStr)
    } catch (error) {
      console.error('时区转换错误:', error)
      return ''
    }
  },
  
  /**
   * 将UTC时间转换为本地时间
   * @param {string|Date} utcTime - UTC时间
   * @param {string} formatStr - 输出格式
   * @returns {string} 本地时间字符串
   */
  utcToLocal: (utcTime, formatStr = 'YYYY-MM-DD HH:mm:ss') => {
    if (!utcTime) return ''
    try {
      return dayjs.utc(utcTime).local().format(formatStr)
    } catch (error) {
      console.error('UTC转本地时间错误:', error)
      return convertUTCToLocal(utcTime) // 使用原有函数作为备选
    }
  },
  
  /**
   * 将本地时间转换为UTC时间
   * @param {string|Date} localTime - 本地时间
   * @param {string} formatStr - 输出格式
   * @returns {string} UTC时间字符串
   */
  localToUtc: (localTime, formatStr = 'YYYY-MM-DD HH:mm:ss') => {
    if (!localTime) return ''
    try {
      return dayjs(localTime).utc().format(formatStr)
    } catch (error) {
      console.error('本地转UTC时间错误:', error)
      return ''
    }
  },
  
  /**
   * 微秒级时间转换为指定单位和格式
   * @param {number} microTimestamp - 微秒级时间戳
   * @param {string} unit - 目标单位，支持 'us', 'ms', 's', 'm', 'h'
   * @param {number} precision - 精度，小数点后位数
   * @returns {number|null} 转换后的时间值
   */
  convertMicroTime: (microTimestamp, unit = 'ms', precision = 2) => {
    return Number(convertTime(microTimestamp, unit, precision))
  },

  /**
   * 获取一天的开始时间
   * @param {Date|number|string} date - 日期对象、时间戳或日期字符串
   * @returns {Date} 当天开始时间的日期对象
   */
  startOfDay: (date) => {
    if (!date) return startOfDay(new Date())
    try {
      return startOfDay(new Date(date))
    } catch (error) {
      console.error('获取日开始时间错误:', error)
      return startOfDay(new Date())
    }
  },
  
  /**
   * 获取一天的结束时间
   * @param {Date|number|string} date - 日期对象、时间戳或日期字符串
   * @returns {Date} 当天结束时间的日期对象
   */
  endOfDay: (date) => {
    if (!date) return endOfDay(new Date())
    try {
      return endOfDay(new Date(date))
    } catch (error) {
      console.error('获取日结束时间错误:', error)
      return endOfDay(new Date())
    }
  },
  
  /**
   * 获取一周的开始时间
   * @param {Date|number|string} date - 日期对象、时间戳或日期字符串
   * @param {Object} options - 配置选项，同date-fns的startOfWeek
   * @returns {Date} 当周开始时间的日期对象
   */
  startOfWeek: (date, options) => {
    if (!date) return startOfWeek(new Date(), options)
    try {
      return startOfWeek(new Date(date), options)
    } catch (error) {
      console.error('获取周开始时间错误:', error)
      return startOfWeek(new Date(), options)
    }
  },
  
  /**
   * 获取一周的结束时间
   * @param {Date|number|string} date - 日期对象、时间戳或日期字符串
   * @param {Object} options - 配置选项，同date-fns的endOfWeek
   * @returns {Date} 当周结束时间的日期对象
   */
  endOfWeek: (date, options) => {
    if (!date) return endOfWeek(new Date(), options)
    try {
      return endOfWeek(new Date(date), options)
    } catch (error) {
      console.error('获取周结束时间错误:', error)
      return endOfWeek(new Date(), options)
    }
  },
  
  /**
   * 获取一个月的开始时间
   * @param {Date|number|string} date - 日期对象、时间戳或日期字符串
   * @returns {Date} 当月开始时间的日期对象
   */
  startOfMonth: (date) => {
    if (!date) return startOfMonth(new Date())
    try {
      return startOfMonth(new Date(date))
    } catch (error) {
      console.error('获取月开始时间错误:', error)
      return startOfMonth(new Date())
    }
  },
  
  /**
   * 获取一个月的结束时间
   * @param {Date|number|string} date - 日期对象、时间戳或日期字符串
   * @returns {Date} 当月结束时间的日期对象
   */
  endOfMonth: (date) => {
    if (!date) return endOfMonth(new Date())
    try {
      return endOfMonth(new Date(date))
    } catch (error) {
      console.error('获取月结束时间错误:', error)
      return endOfMonth(new Date())
    }
  },
  
  /**
  * 获取两个日期之间的时间差
  * @param {Date|number|string} dateLeft - 较早的日期
  * @param {Date|number|string} dateRight - 较晚的日期
  * @param {string} unit - 时间单位，可选值：'years', 'months', 'days', 'hours', 'minutes', 'seconds'
  * @returns {number} 时间差
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
          console.warn(`未知的时间单位: ${unit}，使用默认单位 'days'`);
          return differenceInDays(right, left);
        }
    } catch (error) {
      console.error(`计算${unit}差异错误:`, error);
      return 0;
    }
  },
}