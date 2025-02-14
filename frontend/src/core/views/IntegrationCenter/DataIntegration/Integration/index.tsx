import { Card, ConfigProvider, Tabs, TabsProps } from 'antd'
import IntegrationDoc from './Document'
import SettingsForm from './SettingsForm'
import { useTranslation } from 'react-i18next'
import { useEffect, useState } from 'react'
import { useSearchParams } from 'react-router-dom'
import InstallCmd from './InstallCmd'
import styles from './index.module.scss'
export default function IntegrationSettings() {
  const { t } = useTranslation('core/dataIntegration')
  const [searchParams, setSearchParams] = useSearchParams()
  const [activeKey, setActiveKey] = useState('config')
  const items: TabsProps['items'] = [
    {
      key: 'config',
      label: t('config'),
      children: <SettingsForm />,
      style: {
        height: 'calc(100vh - 220px)',
      },
    },
    {
      key: 'install',
      label: t('installCmdTitle'),
      children: <InstallCmd clusterId={searchParams.get('activeKey')} />,
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
  return (
    <div style={{ height: 'calc(100vh - 100px)' }} className="flex">
      <div className="w-1/2">
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
            styles={{ body: { height: 'calc(100% - 60px)' } }}
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
              <SettingsForm />
            )}
          </Card>
        </ConfigProvider>
      </div>
      <div className="w-1/2 pl-2">
        <IntegrationDoc />
      </div>
    </div>
  )
}
