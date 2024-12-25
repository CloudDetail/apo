/**
 * Copyright 2024 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

// src/core/utils/toast.js

let addToastFunction = null

export const setAddToastFunction = (fn) => {
  addToastFunction = fn
}

export const showToast = ({ title, message = '', color = 'default' }) => {
  if (addToastFunction) {
    addToastFunction({ title, message, color })
  } else {
    console.warn('Toast function is not set')
  }
}
