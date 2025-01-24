/**
 * Copyright 2025 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import opentelemetry from 'src/core/assets/dataCollector/opentelemetry.svg'
import skywalking from 'src/core/assets/dataCollector/skywalking.svg'
import aliyun from 'src/core/assets/dataCollector/aliyun.svg'
import tingyun from 'src/core/assets/dataCollector/tingyun.svg'
import huawei from 'src/core/assets/dataCollector/huawei.svg'
import other from 'src/core/assets/dataCollector/other.svg'
import prometheus from 'src/core/assets/dataCollector/prometheus.svg'
import victoriametrics from 'src/core/assets/dataCollector/victoriametrics.svg'
import zabbix from 'src/core/assets/dataCollector/zabbix.svg'
import tcop from 'src/core/assets/dataCollector/tcop.svg'
import tianyiyun from 'src/core/assets/dataCollector/tianyiyun.svg'
import loki from 'src/core/assets/dataCollector/loki.svg'
import clickhouse from 'src/core/assets/dataCollector/clickhouse.svg'
import kafka from 'src/core/assets/dataCollector/kafka.svg'
import elastic from 'src/core/assets/dataCollector/elastic.svg'
import apo from 'src/core/assets/images/logo.svg'
import cloudEvent from 'src/core/assets/dataCollector/cloudEvent.svg'
import setting from 'src/core/assets/dataCollector/setting.svg'
import {
  DatasourceItemData,
  DatasourceKey,
  DatasourceListData,
  IntegrationType,
  LogsKey,
  MetricsKey,
  TraceKey,
} from './types'
import { AlertKey } from 'src/core/types/alertIntegration'
export const defaultItem: DatasourceItemData = {
  key: 'apo',
  src: apo,
  name: '全新APO安装',
  disableChecked: true,
}
export const datasourceSrc: Record<DatasourceKey, string> = {
  opentelemetry: opentelemetry,
  skywalking: skywalking,
  aliyun: aliyun,
  tingyun: tingyun,
  huawei: huawei,
  other: other,
  prometheus: prometheus,
  victoriametrics: victoriametrics,
  zabbix: zabbix,
  tcop: tcop,
  tianyiyun: tianyiyun,
  loki: loki,
  clickhouse: clickhouse,
  kafka: kafka,
  elastic: elastic,
  apo: apo,
  json: setting,
}
export const traceItems: DatasourceItemData<TraceKey>[] = [
  {
    src: opentelemetry,
    key: 'opentelemetry',
    name: 'Opentelemetry',
    description: '>= 1.11.x ',
  },
  {
    src: skywalking,
    key: 'skywalking',
    name: 'SkyWalking ',
    description: '>= 8.x',
  },
  {
    src: aliyun,
    key: 'aliyun',
    name: '阿里云 ARMS',
  },
  {
    key: 'tingyun',
    src: tingyun,
    name: '听云',
  },
  {
    key: 'huawei',
    src: huawei,
    name: '华为AOM',
    // description: '>= 1.x',
  },
  {
    key: 'other',
    src: other,
    name: '',
    description: '其他支持采用 OTLP 格式输出数据的探针',
  },
]
export const metricsItems: DatasourceItemData<MetricsKey>[] = [
  {
    key: 'prometheus',
    src: prometheus,
    name: 'Prometheus',
  },
  {
    key: 'victoriametrics',
    src: victoriametrics,
    name: 'VictoriaMetrics',
  },
  {
    key: 'zabbix',
    src: zabbix,
    name: 'Zabbix',
  },
  {
    key: 'aliyun',
    src: aliyun,
    name: '阿里云云监控',
  },
  {
    key: 'huawei',
    src: huawei,
    name: '华为云云监控',
  },
  {
    key: 'tcop',
    src: tcop,
    name: '腾讯云可观测平台',
  },
  {
    key: 'tianyiyun',
    src: tianyiyun,
    name: '天翼云云监控',
  },
]
export const logsItems: DatasourceItemData<LogsKey>[] = [
  {
    key: 'elastic',
    src: elastic,
    name: 'ELK',
  },
  {
    key: 'loki',
    src: loki,
    name: 'Loki',
  },
  {
    key: 'clickhouse',
    src: clickhouse,
    name: 'ClickHouse',
  },
  {
    key: 'kafka',
    src: kafka,
    name: 'Kafka',
  },
]
export const alertItems: DatasourceItemData<AlertKey>[] = [
  // defaultItem,
  {
    key: 'json',
    src: setting,
    name: '标准告警源',
  },
  {
    key: 'zabbix',
    src: zabbix,
    name: 'Zabbix',
  },
  {
    key: 'prometheus',
    src: prometheus,
    name: 'Prometheus',
  },
]

export const datasourceMap: Record<IntegrationType, DatasourceListData> = {
  trace: {
    type: 'trace',
    title: '链路追踪接入',
    items: traceItems,
  },
  metrics: {
    type: 'metrics',
    title: '指标接入',
    description: '建议选择APO自采，其他接入方式可能导致数据缺失',
    items: metricsItems,
  },
  logs: {
    type: 'logs',
    title: '日志接入',
    description: '建议选择APO自采，其他接入方式可能导致数据缺失',
    items: logsItems,
  },
  alert: {
    type: 'alert',
    title: '告警事件接入',
    // description: '建议选择APO自采，其他接入方式可能导致数据缺失',
    items: alertItems,
  },
}
