/**
 * Copyright 2024 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

// 时间转换工具类

import { format } from 'date-fns'
import dayjs from 'dayjs'
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

export function convertUTCToBeijing(utcTimeString) {
  // 将 UTC 时间字符串转换为 Date 对象
  const utcDate = new Date(utcTimeString)

  // 获取 UTC 时间
  const beijingDate = new Date(utcDate.toLocaleString('en-US', { timeZone: 'Asia/Shanghai' }))

  // 格式化为 YYYY-MM-DD HH:mm:ss
  const year = beijingDate.getFullYear()
  const month = String(beijingDate.getMonth() + 1).padStart(2, '0')
  const day = String(beijingDate.getDate()).padStart(2, '0')
  const hours = String(beijingDate.getHours()).padStart(2, '0')
  const minutes = String(beijingDate.getMinutes()).padStart(2, '0')
  const seconds = String(beijingDate.getSeconds()).padStart(2, '0')

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
