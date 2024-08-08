// 时间范围 （a,b]	步长	点数
// 0～15min	       30s	  1～30
// 15～30 min	   1min	  15～30
// 30min ~ 2.5h	   5min	  6~30
// 2.5h ~ 5h	   10min  15~30
// 5h~15h	       30min  10 ~ 30
// 15h~30h	       1h	  15~30
// 30h-max         (endTime - startTime)/ 30


// 时间范围 （a,b]	步长	点数
// 0～7.5min	   30s	  1～15
// 7.5min～15 min   1min	 8 ～15
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
    const timeDiff = endTime - startTime;

    const SECOND = 1000000; // 微秒
    const MINUTE = 60 * SECOND;
    const HOUR = 60 * MINUTE;

    let step = SECOND; // 默认步长为1秒
    // 最长30点数版本
    // if (timeDiff <= 15 * MINUTE) {
    //     step = 30 * SECOND;
    // } else if (timeDiff <= 30 * MINUTE) {
    //     step = 60 * SECOND;
    // } else if (timeDiff <= 2.5 * HOUR) {
    //     step = 5 * 60 * SECOND;
    // } else if (timeDiff <= 5 * HOUR) {
    //     step = 10 * 60 * SECOND;
    // } else if (timeDiff <= 15 * HOUR) {
    //     step = 30 * 60 * SECOND;
    // } else if (timeDiff <= 30 * HOUR) {
    //     step = 60 * 60 * SECOND;
    // } else {
    //     // 确保步长为整秒数
    //     step = Math.ceil(timeDiff / 30 / SECOND) * SECOND;
    // }
    // 最长30点数版本
    if (timeDiff <= 7.5 * MINUTE) {
        step = 30 * SECOND;
    } else if (timeDiff <= 15 * MINUTE) {
        step = 60 * SECOND;
    } else if (timeDiff <= 1.25 * HOUR) {
        step = 5 * 60 * SECOND;
    } else if (timeDiff <= 2.5 * HOUR) {
        step = 10 * 60 * SECOND;
    } else if (timeDiff <= 7.5 * HOUR) {
        step = 30 * 60 * SECOND;
    } else if (timeDiff <= 15 * HOUR) {
        step = 60 * 60 * SECOND;
    } else {
        // 确保步长为整秒数
        step = Math.ceil(timeDiff / 15 / SECOND) * SECOND;
    }

    return step;
}
