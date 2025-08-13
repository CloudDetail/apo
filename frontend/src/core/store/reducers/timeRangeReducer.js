/**
 * Copyright 2024 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

// 默认近15分钟

import { createSelector } from 'reselect'

//
export const initialTimeRangeState = {
  startTime: null,
  endTime: null,
  rangeType: null,
  rangeTypeKey: null,
  //末次刷新时间
  refreshTimestamp: null,
  refreshInInterval: null,
}
export function GetInitalTimeRange() {
  return {
    startTime: new Date().getTime() * 1000 - 15 * 60 * 1000000,
    endTime: new Date().getTime() * 1000,
    rangeType: null,
  }
}
export function GetInitalTimeRangeState() {
  return {
    rangeTypeKey: '15m',
  }
}
export function getTimestampRange(rangeTypeKey, convertToSecond = false) {
  const now = new Date()
  const nowMicroseconds = now.getTime() * 1000 // 当前时间的时间戳（微秒）

  let startTime, endTime
  endTime = nowMicroseconds
  if (!rangeTypeKey) {
    startTime = new Date()
  } else if (typeof timeRangeMap[rangeTypeKey].rangeType === 'number') {
    // 处理数字分钟数
    startTime = endTime - timeRangeMap[rangeTypeKey].rangeType * 60 * 1000000
  } else {
    switch (rangeTypeKey) {
      case 'today':
        const startOfDay = new Date(now.getFullYear(), now.getMonth(), now.getDate()).getTime()
        startTime = startOfDay * 1000
        break
      case 'yesterday':
        const startOfYesterday = new Date(
          now.getFullYear(),
          now.getMonth(),
          now.getDate() - 1,
        ).getTime()
        const endOfYesterday = startOfYesterday + 24 * 60 * 60 * 1000 - 1 // 减去1毫秒
        startTime = startOfYesterday * 1000
        endTime = endOfYesterday * 1000
        break
      case 'thisWeek':
        const day = now.getDay() // 周日=0...周六=6
        const offsetToMon = (day + 6) % 7 // 距离周一的天数
        const mondayMs = new Date(
          now.getFullYear(),
          now.getMonth(),
          now.getDate() - offsetToMon,
          0, 0, 0, 0
        ).getTime()
        startTime = mondayMs * 1000
        break
      case 'last7Days': { // 最近 7 天（包括今天）
        const startOfLast7Days = new Date(
          now.getFullYear(),
          now.getMonth(),
          now.getDate() - 6, // 往前推 6 天，加上今天共 7 天
          0, 0, 0, 0
        ).getTime()
        startTime = startOfLast7Days * 1000
        break
      }
      default:
        throw new Error('Unknown range type')
    }
  }
  if (convertToSecond) {
    return { startTime: toNearestSecond(startTime), endTime: toNearestSecond(endTime) }
  }

  return { startTime, endTime }
}
const timeRangeReducer = (state = initialTimeRangeState, action) => {
  switch (action.type) {
    case 'SET_TIMERANGE':
      if (action.payload.rangeTypeKey) {
        const { startTime, endTime } = getTimestampRange(action.payload.rangeTypeKey)
        return {
          ...state,
          startTime: startTime,
          endTime: endTime,
          rangeTypeKey: action.payload.rangeTypeKey,
          refreshTimestamp: Date.now() * 1000,
        }
      } else {
        return {
          ...state,
          startTime: action.payload.startTime,
          endTime: action.payload.endTime,
          rangeTypeKey: null,
          refreshTimestamp: null,
        }
      }
    case 'INIT_TIMERANGE': {
      const { startTime, endTime } = getTimestampRange('15m')
      return {
        ...state,
        startTime: startTime,
        endTime: endTime,
        rangeTypeKey: '15m',
        refreshTimestamp: Date.now() * 1000,
      }
    }
    case 'REFRESH_TIMERANGE': {
      const { startTime, endTime } = getTimestampRange(state.rangeTypeKey)
      return {
        ...state,
        startTime: startTime,
        endTime: endTime,
        refreshTimestamp: Date.now() * 1000,
      }
    }

    default:
      return state
  }
}

export default timeRangeReducer

// 基础 selector
const selectTimeRangeState = (state) => state.timeRange

// 返回微妙级的
export const selectProcessedTimeRange = createSelector([selectTimeRangeState], (timeRangeState) => {
  const { startTime, endTime, rangeType } = timeRangeState
  // 在这里对 startTime 和 endTime 进行处理，例如返回一个格式化后的对象
  return rangeType
    ? getTimestampRange(rangeType)
    : {
      startTime: startTime,
      endTime: endTime,
    }
})
//返回整秒的微妙时间戳，用于步长传参
export const selectSecondsTimeRange = createSelector([selectTimeRangeState], (timeRangeState) => {
  const { startTime, endTime, rangeType } = timeRangeState
  // 在这里对 startTime 和 endTime 进行处理，例如返回一个格式化后的对象
  return rangeType
    ? getTimestampRange(rangeType, true)
    : {
      startTime: toNearestSecond(startTime),
      endTime: toNearestSecond(endTime),
    }
})

// 转微秒级时间戳为整秒的微秒时间戳
export function toNearestSecond(microseconds) {
  const seconds = Math.floor(microseconds / 1e6)
  return seconds * 1e6
}

export const timeRangeList = [
  {
    name: 'Last 5 minutes',
    rangeType: 5,
    from: 'now-5m',
    to: 'now',
  },
  {
    name: 'Last 15 minutes',
    rangeType: 15,
    from: 'now-15m',
    to: 'now',
  },
  {
    name: 'Last 30 minutes',
    rangeType: 30,
    from: 'now-30m',
    to: 'now',
  },
  {
    name: 'Last 1 hours',
    rangeType: 60,
    from: 'now-1h',
    to: 'now',
  },
  {
    name: 'Last 3 hours',
    rangeType: 180,
    from: 'now-3h',
    to: 'now',
  },
  {
    name: 'Last 6 hours',
    rangeType: 360,
    from: 'now-6h',
    to: 'now',
  },
  {
    name: 'Last 12 hours',
    rangeType: 720,
    from: 'now-12h',
    to: 'now',
  },
  {
    name: 'Today',
    rangeType: 'today',
    from: 'now/d',
    to: 'now/d',
  },
  {
    name: 'Yesterday',
    rangeType: 'yesterday',
    from: 'now-1d/d',
    to: 'now-1d/d',
  },
  {
    name: 'This week',
    rangeType: 'this week',
    from: 'now/w',
    to: 'now/w',
  },
]

export const timeRangeMap = {
  '5m': {
    name: 'Last 5 minutes',
    rangeType: 5,
    from: 'now-5m',
    to: 'now',
  },
  '15m': {
    name: 'Last 15 minutes',
    rangeType: 15,
    from: 'now-15m',
    to: 'now',
  },
  '30m': {
    name: 'Last 30 minutes',
    rangeType: 30,
    from: 'now-30m',
    to: 'now',
  },
  '1h': {
    name: 'Last 1 hours',
    rangeType: 60,
    from: 'now-1h',
    to: 'now',
  },
  '3h': {
    name: 'Last 3 hours',
    rangeType: 180,
    from: 'now-3h',
    to: 'now',
  },
  '6h': {
    name: 'Last 6 hours',
    rangeType: 360,
    from: 'now-6h',
    to: 'now',
  },
  '12h': {
    name: 'Last 12 hours',
    rangeType: 720,
    from: 'now-12h',
    to: 'now',
  },
  today: {
    name: 'Today',
    rangeType: 'today',
    from: 'now/d',
    to: 'now/d',
  },
  yesterday: {
    name: 'Yesterday',
    rangeType: 'yesterday',
    from: 'now-1d/d',
    to: 'now-1d/d',
  },
  thisWeek: {
    name: 'This week',
    rangeType: 'this week',
    from: 'now/w',
    to: 'now/w',
  },
  last7Days: {
    name: 'Last 7 Days',
    rangeType: 'last7Days',
  }
}
