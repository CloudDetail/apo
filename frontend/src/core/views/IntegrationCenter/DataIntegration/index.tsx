/**
 * Copyright 2025 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import ProbeManagement from './ProbeManagement'
import styles from './dataIntegrationPage.module.scss'

export default function DataIntegrationPage() {
  return (
    <div className={styles.page}>
      <ProbeManagement />
    </div>
  )
}
