// 时间转换工具类

import { format } from 'date-fns'

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

export function convertUTCToBeijing(utcString) {
  // 将 UTC 字符串转换为 Date 对象
  const utcDate = new Date(utcString)

  // 获取中国北京时间（UTC+8）的毫秒数
  const beijingTime = new Date(utcDate.getTime() + 8 * 60 * 60 * 1000)

  // 格式化输出为 yyyy-mm-dd hh:mm:ss
  const year = beijingTime.getFullYear()
  const month = String(beijingTime.getMonth() + 1).padStart(2, '0')
  const day = String(beijingTime.getDate()).padStart(2, '0')
  const hours = String(beijingTime.getHours()).padStart(2, '0')
  const minutes = String(beijingTime.getMinutes()).padStart(2, '0')
  const seconds = String(beijingTime.getSeconds()).padStart(2, '0')

  return `${year}-${month}-${day} ${hours}:${minutes}:${seconds}`
}
