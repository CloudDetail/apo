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
import pinpoint from 'src/core/assets/dataCollector/pinpoint.png'
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
import { t } from 'i18next'
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
  prometheus: prometheus,
  victoriametric: victoriametrics,
  zabbix: zabbix,
  tcop: tcop,
  tianyiyun: tianyiyun,
  loki: loki,
  clickhouse: clickhouse,
  kafka: kafka,
  elastic: elastic,
  apo: apo,
  json: setting,
  pinpoint: pinpoint,
  other: other,
}
export const traceItems: DatasourceItemData<TraceKey>[] = [
  {
    src: opentelemetry,
    key: 'opentelemetry',
    name: 'Opentelemetry',
    description: '>= 1.11.x ',
    apmType: 'jaeger',
  },
  {
    src: skywalking,
    key: 'skywalking',
    name: 'SkyWalking',
    description: '>= 8.x',
    apmType: 'skywalking',
  },
  {
    src: aliyun,
    key: 'aliyun',
    name: t('core/dataIntegration:aliyunArms'),
    apmType: 'arms',
  },
  {
    key: 'tingyun',
    src: tingyun,
    name: t('core/dataIntegration:tingyun'),
    apmType: 'nbs3',
  },
  {
    key: 'huawei',
    src: huawei,
    name: t('core/dataIntegration:hauwei'),
    apmType: 'huawei',
    // description: '>= 1.x',
  },
  {
    key: 'elastic',
    src: elastic,
    name: 'ELK',
    apmType: 'elastic',
  },
  {
    key: 'pinpoint',
    src: pinpoint,
    name: 'PinPoint',
    apmType: 'pinpoint',
  },
  // {
  //   key: 'other',
  //   src: other,
  //   name: t('core/dataIntegration:other'),
  //   description: '其他支持采用 OTLP 格式输出数据的探针',
  //   apmType: 'other',
  // },
]
export const metricsItems: DatasourceItemData<MetricsKey>[] = [
  {
    key: 'prometheus',
    src: prometheus,
    name: 'Prometheus',
  },
  {
    key: 'victoriametric',
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
    name: t('core/alertsIntegration:json'),
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

export const datasourceMap: Partial<Record<IntegrationType, DatasourceListData>> = {
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
  // alert: {
  //   type: 'alert',
  //   title: '告警事件接入',
  //   // description: '建议选择APO自采，其他接入方式可能导致数据缺失',
  //   items: alertItems,
  // },
}

export const portsDefault = [
  {
    key: 'apoCollector',
    title: 'apo-collector',
    value: '30044',
    descriptions: t('core/dataIntegration:apoCollector.port1'),
  },
  {
    key: 'apoVector',
    title: 'apo-vector',
    value: '30310',
    descriptions: t('core/dataIntegration:apoCollector.port2'),
  },
  {
    key: 'apoOtelCollectorGatewayGrpc',
    title: 'apo-otel-collector-gateway',
    value: '30317',
    descriptions: t('core/dataIntegration:apoCollector.port3'),
  },
  {
    key: 'apoOtelCollectorGatewayK8s',
    title: 'apo-otel-collector-gateway',
    value: '30319',
    descriptions: t('core/dataIntegration:apoCollector.port4'),
  },
  {
    key: 'apoBackend',
    title: 'apo-backend',
    value: '31363',
    descriptions: t('core/dataIntegration:apoCollector.port5'),
  },
]
