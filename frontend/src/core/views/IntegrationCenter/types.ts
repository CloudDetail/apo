/**
 * Copyright 2025 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import { ReactNode } from 'react'
import { AlertKey } from 'src/core/types/alertIntegration'

export type IntegrationType = 'trace' | 'metrics' | 'logs' | 'alert'

export type ApoKey = 'apo'
export type TraceKey = 'opentelemetry' | 'skywalking' | 'aliyun' | 'tingyun' | 'huawei' | 'other'
export type MetricsKey =
  | 'prometheus'
  | 'victoriametrics'
  | 'zabbix'
  | 'aliyun'
  | 'huawei'
  | 'tcop'
  | 'tianyiyun'
export type LogsKey = 'elastic' | 'loki' | 'clickhouse' | 'kafka'
export type DatasourceKey = ApoKey | TraceKey | MetricsKey | LogsKey | AlertKey

export interface DatasourceItemData<T = DatasourceKey> {
  src: string
  key: T
  name: string | ReactNode
  description?: string
  disableChecked?: boolean
}
export interface DatasourceListData {
  type: IntegrationType
  title: string | ReactNode
  description?: string
  items: DatasourceItemData[]
}
export type TraceConfigType = 'adapter' | 'forwarder' | 'both'
export type MetricsConfigType = 'bridge' | 'proxy' | 'both'
export type LogsConfigType = 'forwarder' | 'connector' | 'both'

export interface Schemas {
  [key: string]: string[]
}

export interface TargetTag {
  id: string
  tagName: string
}
