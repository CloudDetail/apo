/**
 * Copyright 2025 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */
import type { NotificationPlacement } from 'antd/es/notification/interface'
import { ReactNode } from 'react'

let notifyApi: any = null

export const setNotifyApi = (api: any) => {
  notifyApi = api
}

export interface NotifyOptions {
  type: 'success' | 'error' | 'info' | 'warning'
  message: string | ReactNode
  description?: string
  duration?: number
  placement?: NotificationPlacement
  onClick?: () => void
}

export const notify = (options: NotifyOptions) => {
  if (!notifyApi) {
    console.warn('error')
    return
  }

  const { type, ...rest } = options

  notifyApi[type]?.({
    duration: 3,
    placement: 'topRight',
    ...rest,
  })
}
