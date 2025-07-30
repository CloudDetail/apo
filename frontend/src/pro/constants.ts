import { t } from 'i18next'
import apo from 'src/core/assets/images/logo.svg'
import datadog from 'src/pro/assets/datadog.png'
import newrelic from 'src/pro/assets/newrelic.png'
import dynatrace from 'src/pro/assets/dynatrace.png'
import nginx from 'src/pro/assets/nginx.png'
import arms from 'src/pro/assets/arms.png'
import prometheus from 'src/core/assets/dataCollector/prometheus.svg'
import loki from 'src/core/assets/dataCollector/loki.svg'
import tempo from 'src/pro/assets/tempo.png'
import openobserve from 'src/pro/assets/openobserve.png'
import signoz from 'src/pro/assets/signoz.png'
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
  'arms': arms,
  'Prometheus': prometheus,
  'Loki': loki,
  'Tempo': tempo,
  'OpenObserve': openobserve,
  'SigNoz': signoz,
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