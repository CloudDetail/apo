import { t } from 'i18next'
import apo from 'src/core/assets/images/logo.svg'
import datadog from 'src/pro/assets/datadog.png'
import newrelic from 'src/pro/assets/newrelic.png'
import dynatrace from 'src/pro/assets/dynatrace.png'
import nginx from 'src/pro/assets/nginx.png'
import aliyun from 'src/core/assets/dataCollector/aliyun.svg'
import prometheus from 'src/core/assets/dataCollector/prometheus.svg'
import loki from 'src/pro/assets/loki.svg'
import tempo from 'src/pro/assets/tempo.svg'
import openobserve from 'src/pro/assets/openobserve.png'
import signoz from 'src/pro/assets/signoz.svg'
import skywalking from 'src/core/assets/dataCollector/skywalking.svg'
import grafana from 'src/pro/assets/grafana.svg'
import jaeger from 'src/pro/assets/jaeger.svg'

export const anormalTypesOptions = [
  {
    value: 'unknown',
    key: 0,
    label: t('pro/constant:unknown'),
    hide: true,
  },
  {
    value: 'app',
    key: 1,
    label: t('pro/constant:app'),
  },
  {
    value: 'container',
    key: 2,
    label: t('pro/constant:container'),
  },
  {
    value: 'infra',
    key: 3,
    label: t('pro/constant:infra'),
  },
  {
    value: 'network',
    key: 4,
    label: t('pro/constant:network'),
  },
  {
    value: 'error',
    key: 5,
    label: t('pro/constant:error'),
  },
  {
    value: 'appInstance',
    key: 6,
    label: t('pro/constant:appInstance'),
  },
  {
    value: 'middleware',
    key: 7,
    label: t('pro/constant:middleware'),
  },
]

export const anormalTypes = [
  'app',
  'container',
  'infra',
  'network',
  'error',
  'appInstance',
  'middleware',
]
export const dataplaneIconMap = {
  'apo': apo,
  'datadog': datadog,
  'Apo': apo,
  'Datadog': datadog,
  'NewRelic': newrelic,
  'Dynatrace': dynatrace,
  'nginxLog': nginx,
  'arms': aliyun,
  'Prometheus': prometheus,
  'prometheus': prometheus,
  'loki': loki,
  'Tempo': tempo,
  'tempo': tempo,
  'OpenObserve': openobserve,
  'SigNoz': signoz,
  'signoz': signoz,
  'skywalking': skywalking,
  'jaeger': jaeger,
  'grafanaPrometheus': grafana,
}
export const datasourceOptions = [
  {
    value: 'APO',
    label: 'APO',
    icon: apo,
    dataTypes: ['RED Metrics', 'Service Topology', 'Basic Metrics', 'Logs', 'Traces', 'Alerts',],
  },
  {
    value: 'datadog',
    label: 'datadog',
    icon: datadog,
    dataTypes: ['RED Metrics', 'Service Topology', 'Basic Metrics', 'Logs', 'Traces', 'Alerts'],

  },
  {
    value: 'NewRelic',
    label: 'NewRelic',
    icon: newrelic,
    dataTypes: ['RED Metrics', 'Service Topology', 'Basic Metrics', 'Logs', 'Traces', 'Issues',],
  },
  {
    value: 'Dynatrace',
    label: 'Dynatrace',
    icon: dynatrace,
    dataTypes: ['RED Metrics', 'Service Topology', 'Basic Metrics', 'Logs', 'Traces', 'Problems'],

  }
]