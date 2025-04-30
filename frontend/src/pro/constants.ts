import { t } from "i18next"

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
]

export const anormalTypes = ['app', 'container', 'infra', 'network', 'error', 'appInstance']
