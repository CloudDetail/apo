// 默认近15分钟
// rangeValue: number/ string/ null 空的情况下直接用指定的starttime endtime

import { createSelector } from 'reselect'

//
export const initialTimeRangeState = {
  startTime: new Date().getTime() * 1000 - 15 * 60 * 1000000,
  endTime: new Date().getTime() * 1000,
  rangeType: null,
}
export function GetInitalTimeRange() {
  return {
    startTime: new Date().getTime() * 1000 - 15 * 60 * 1000000,
    endTime: new Date().getTime() * 1000,
    rangeType: null,
  }
}

export function getTimestampRange(rangeType, convertToSecond = false) {
  const now = new Date()
  const nowMicroseconds = now.getTime() * 1000 // 当前时间的时间戳（微秒）

  let startTime, endTime
  endTime = nowMicroseconds
  if (!rangeType) {
    startTime = new Date()
  } else if (typeof rangeType === 'number') {
    // 处理数字分钟数
    console.log(rangeType)
    startTime = endTime - rangeType * 60 * 1000000
    console.log(startTime)
  } else {
    switch (rangeType) {
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
      case 'this week':
        const startOfWeek = new Date(now.setDate(now.getDate() - now.getDay())).setHours(0, 0, 0, 0)
        startTime = startOfWeek * 1000
        break
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
      return {
        ...state,
        startTime: action.payload.startTime,
        endTime: action.payload.endTime,
        rangeType: action.payload.rangeType,
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
