/**
 * Copyright 2025 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import { Tabs } from 'antd'
import { useSearchParams } from 'react-router-dom'
import ClusterTable from './ClusterTable'
import ProbeManagement from './ProbeManagement'
import styles from './probeManagement.module.scss'

export default function DataIntegrationPage() {
  const [searchParams, setSearchParams] = useSearchParams()
  const activeTab = searchParams.get('tab') === 'management' ? 'management' : 'access'

  const changeTab = (key: string) => {
    const nextParams = new URLSearchParams(searchParams)
    nextParams.set('tab', key)
    setSearchParams(nextParams, { replace: true })
  }

  return (
    <div className={styles.page}>
      <Tabs
        className={styles.tabs}
        activeKey={activeTab}
        onChange={changeTab}
        items={[
          { key: 'access', label: '探针接入' },
          { key: 'management', label: '探针管理' },
        ]}
      />
      {activeTab === 'access' ? <ClusterTable /> : <ProbeManagement />}
    </div>
  )
}
