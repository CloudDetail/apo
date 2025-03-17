/**
 * Copyright 2025 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

function debounce(func, delay, immediate) {
  let timer;
  return function () {
    if(timer) clearTimeout(timer)

    if(immediate) {
      let firstRun = !timer

      timer = setTimeout(() => {
        timer = null
      }, delay)

      if(firstRun) {
        func.apply(this, arguments)
      }
    } else {
      timer = setTimeout(() => {
        func.apply(this, arguments)
      }, delay)
    }
  }
}

export default debounce