/**
 * Copyright 2024 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

// 时间范围 （a,b]	步长	点数
// 0～15min	       30s	  1～30
// 15～30 min	   1min	  15～30
// 30～1h          2min   15～30
// 1h~ 1.5h        3min   20~30
// 1.5h ~ 3h	   6min	  15~30
// 3h ~ 6h	       12min  15~30
// 6h~12h	       24min  15~ 30
// 12h~15h	       30min  24~30
// 15h~30h	       1h	  15~30
// 30h-max         (endTime - startTime)/ 30

// 时间范围 （a,b]	步长	点数
// 0～7.5min	   30s	  1～15
// 7.5min～15 min   1min	 8 ～15
//
// 15min ~1.25 h   5min	  6~15
// 1.25h ~ 2.5h	   10min  15
// 2.5h~7.5h       30min   15
// 7.5h~15h	       1h	  ~15
// 15h-max         (endTime - startTime)/ 15
/**
 * 定义步长
 * @param {number} startTime - 开始时间 微妙时间戳
 * @param {number} endTime - 结束时间 微妙时间戳
 * @returns {number} 返回步长（微妙）
 */
export function getStep(startTime, endTime) {
  const timeDiff = endTime - startTime

  const SECOND = 1000000 // 微秒
  const MINUTE = 60 * SECOND
  const HOUR = 60 * MINUTE

  let step = SECOND // 默认步长为1秒

  if (timeDiff <= 15 * MINUTE) {
    step = 30 * SECOND
  } else if (timeDiff <= 30 * MINUTE) {
    step = 1 * MINUTE
  } else if (timeDiff <= 1 * HOUR) {
    step = 2 * MINUTE
  } else if (timeDiff <= 1.5 * HOUR) {
    step = 3 * MINUTE
  } else if (timeDiff <= 3 * HOUR) {
    step = 6 * MINUTE
  } else if (timeDiff <= 6 * HOUR) {
    step = 12 * MINUTE
  } else if (timeDiff <= 12 * HOUR) {
    step = 24 * MINUTE
  } else if (timeDiff <= 15 * HOUR) {
    step = 30 * MINUTE
  } else if (timeDiff <= 30 * HOUR) {
    step = 1 * HOUR
  } else {
    // 如果时间范围超过30小时，则步长为总时间差除以30（确保步长为整秒数）
    step = Math.ceil(timeDiff / 30 / SECOND) * SECOND
  }

  // 最长15点数版本
  // if (timeDiff <= 7.5 * MINUTE) {
  //     step = 30 * SECOND;
  // } else if (timeDiff <= 15 * MINUTE) {
  //     step = 60 * SECOND;
  // } else if (timeDiff <= 1.25 * HOUR) {
  //     step = 5 * 60 * SECOND;
  // } else if (timeDiff <= 2.5 * HOUR) {
  //     step = 10 * 60 * SECOND;
  // } else if (timeDiff <= 7.5 * HOUR) {
  //     step = 30 * 60 * SECOND;
  // } else if (timeDiff <= 15 * HOUR) {
  //     step = 60 * 60 * SECOND;
  // } else {
  //     // 确保步长为整秒数
  //     step = Math.ceil(timeDiff / 15 / SECOND) * SECOND;
  // }

  return step
}
