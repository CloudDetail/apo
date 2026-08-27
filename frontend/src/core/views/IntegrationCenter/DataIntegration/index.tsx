/**
 * Copyright 2025 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import { Tabs } from 'antd'
import { useTranslation } from 'react-i18next'
import { useSearchParams } from 'react-router-dom'
import BotManagement from './BotManagement'
import ProbeManagement from './ProbeManagement'
import styles from './dataIntegrationPage.module.scss'

const MANAGEMENT_TAB = 'management'
const BOTS_TAB = 'bots'

export default function DataIntegrationPage() {
  const { t } = useTranslation('core/dataIntegration')
  const [searchParams, setSearchParams] = useSearchParams()
  const activeTab = searchParams.get('tab') === BOTS_TAB ? BOTS_TAB : MANAGEMENT_TAB

  const handleTabChange = (tab: string) => {
    const nextParams = new URLSearchParams(searchParams)
    if (tab === BOTS_TAB) {
      nextParams.set('tab', BOTS_TAB)
    } else {
      nextParams.delete('tab')
    }
    setSearchParams(nextParams)
  }

  return (
    <div className={styles.page}>
      <Tabs
        className={styles.tabs}
        activeKey={activeTab}
        onChange={handleTabChange}
        items={[
          {
            key: MANAGEMENT_TAB,
            label: t('probeManagementTabs.management'),
            children: activeTab === MANAGEMENT_TAB ? <ProbeManagement /> : null,
          },
          {
            key: BOTS_TAB,
            label: t('probeManagementTabs.bots'),
            children: activeTab === BOTS_TAB ? <BotManagement /> : null,
          },
        ]}
      />
    </div>
  )
}
