/**
 * Copyright 2025 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import { Button, Card, ConfigProvider, Tabs, TabsProps } from 'antd'
import SettingsForm from './SettingsForm'
import { useTranslation } from 'react-i18next'
import { useEffect, useState } from 'react'
import { useLocation, useNavigate, useSearchParams } from 'react-router-dom'
import styles from './index.module.scss'
import InstallCmd from 'src/core/components/InstallCmd'
import { getClusterIntegrationInfoApi, getIntegrationConfigApi } from 'src/core/api/integration'
export default function IntegrationSettings() {
  const { t } = useTranslation('core/dataIntegration')
  const [searchParams, setSearchParams] = useSearchParams()
  const [activeKey, setActiveKey] = useState('config')
  const [formInitValues, setFormInitValues] = useState(null)
  const { pathname } = useLocation()
  const navigate = useNavigate()
  const isMinimal = pathname === '/probe-management/settings'
  const handleCancel = () => {
    if (!isMinimal) {
      navigate('/integration/data')
    } else {
      navigate('/probe-management')
    }
  }
  const items: TabsProps['items'] = [
    {
      key: 'config',
      label: t('config'),
      children: <SettingsForm formInitValues={formInitValues} />,
      style: {
        height: 'calc(100vh - 220px)',
      },
    },
    {
      key: 'install',
      label: t('installCmdTitle'),
      children: (
        <div className="h-full w-full flex flex-col">
          <div className="w-full flex-1 h-0 overflow-auto">
            <InstallCmd
              clusterId={searchParams.get('clusterId')}
              clusterType={searchParams.get('clusterType')}
              apoCollector={formInitValues?.apoCollector}
              isMinimal={isMinimal}
            />
          </div>

          <div className={`${styles.bottomDiv} w-full `}>
            <Button className="mr-3" onClick={handleCancel}>
              {t('goBack')}
            </Button>
          </div>
        </div>
      ),
      style: {
        height: 'calc(100vh - 220px)',
      },
    },
  ]
  const changeTab = (key: string) => {
    const newParams = new URLSearchParams(searchParams)
    newParams.set('activeKey', key)
    setSearchParams(newParams, { replace: true })
  }
  useEffect(() => {
    const activeKey = searchParams.get('activeKey')
    if (activeKey) {
      setActiveKey(activeKey)
    }
  }, [searchParams])

  const getClusterIntegrationInfo = async (clusterId: string) => {
    const res = await getClusterIntegrationInfoApi(clusterId)
    return { ...res }
  }
  const getIntegrationInfo = async () => {
    const res = await getIntegrationConfigApi()
    const { database, datasource, traceAPI } = res
    return {
      // metric: {
      //   ...datasource,
      //   metricAPI: {
      //     vmConfig: datasource.metricAPI.victoriametric,
      //   },
      // },
      // log: {
      //   ...database,
      //   logAPI: {
      //     chConfig: database.logAPI.clickhouse,
      //   },
      // },
      traceAPI,
    }
  }
  useEffect(() => {
    const fetchData = async () => {
      const clusterId = searchParams.get('clusterId')

      if (clusterId) {
        const [integrationData, clusterData] = await Promise.all([
          getIntegrationInfo(),
          getClusterIntegrationInfo(clusterId),
        ])

        const mergedData = { ...integrationData, ...clusterData }
        setFormInitValues(mergedData)
      } else {
        const integrationData = await getIntegrationInfo()
        setFormInitValues(integrationData)
      }
    }

    fetchData()
  }, [searchParams])
  return (
    <div style={{ height: 'calc(100vh - 100px)' }} className="flex">
      <div className="w-full">
        <ConfigProvider
          theme={{
            components: {
              Form: {
                itemMarginBottom: 6,
              },
            },
          }}
        >
          <Card
            title={t('dataIntegrationSettings')}
            className="h-full overflow-hidden"
            classNames={{ body: 'p-0 overflow-auto flex flex-col' }}
            styles={{ body: { height: 'calc(100% - 60px)' }, header: { minHeight: '45px' } }}
          >
            {searchParams.get('clusterId') ? (
              <div className="px-3 overflow-hidden">
                <Tabs
                  items={items}
                  className={styles.tabs}
                  onChange={(e) => changeTab(e)}
                  activeKey={activeKey}
                />
              </div>
            ) : (
              <SettingsForm formInitValues={formInitValues} />
            )}
          </Card>
        </ConfigProvider>
      </div>
      {/* <div className="w-1/2 pl-2">
        <IntegrationDoc />
      </div> */}
    </div>
  )
}
