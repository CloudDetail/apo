/**
 * Copyright 2025 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import { Card } from 'antd'
import { AlertKey } from 'src/core/types/alertIntegration'
import ZabbixInfo from './ZabbixInfo'
import styles from './index.module.scss'
import PrometheusInfo from './PrometheusInfo'
import JsonInfo from './JsonInfo'
function SourceTypeInfo({ sourceType }) {
  return (
    <Card className={styles.sourceTypeInfo}>
      {sourceType === 'zabbix' ? (
        <ZabbixInfo />
      ) : sourceType === 'prometheus' ? (
        <PrometheusInfo />
      ) : (
        <JsonInfo />
      )}
    </Card>
  )
}
export default SourceTypeInfo
