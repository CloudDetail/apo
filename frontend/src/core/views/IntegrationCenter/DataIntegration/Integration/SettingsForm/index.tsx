import { Button, Card, ConfigProvider, Divider, Form, Input, Segmented, Typography } from 'antd'
import { t } from 'i18next'
import { useTranslation } from 'react-i18next'
import TraceFormItem from './TraceFormItem'
import { useEffect } from 'react'
import {
  createDataIntegrationApi,
  getClusterIntegrationInfoApi,
  getIntegrationConfigApi,
  updateDataIntegrationApi,
} from 'src/core/api/integration'
import MetricsFormItem from './MetricsFormItem'
import LogsFormItem from './LogsFormItem'
import styles from './index.module.scss'
import { showToast } from 'src/core/utils/toast'
import APOCollectorFormItem from './APOCollectorFormItem'
import { useSearchParams } from 'react-router-dom'
import { portsDefault } from '../../../constant'
const clusterTypeOptions = ['k8s', 'vm']
const SettingsForm = () => {
  const { t } = useTranslation('core/dataIntegration')
  const { t: ct } = useTranslation('common')
  const [form] = Form.useForm()
  const [searchParams, setSearchParams] = useSearchParams()

  const saveIntegration = (params) => {
    delete params?.traceAPI
    let api = params.id ? updateDataIntegrationApi : createDataIntegrationApi
    api(params).then((res) => {
      showToast({
        color: 'success',
        title: t('saveSettingsSuccess'),
      })
      const newParams = new URLSearchParams(searchParams)
      newParams.set('clusterId', params.id || res?.id)
      newParams.set('activeKey', 'install')
      setSearchParams(newParams, { replace: true })
    })
  }
  const saveInfo = () => {
    form.validateFields().then((values) => {
      const params = { ...values }
      if (!values.apoCollector.ports || values.apoCollector.ports === undefined) {
        params.apoCollector.ports = portsDefault?.reduce((map, item) => {
          map[item.key] = item.value
          return map
        }, {})
      }
      saveIntegration(params)
    })
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

  const getClusterIntegrationInfo = async (clusterId: string) => {
    const res = await getClusterIntegrationInfoApi(clusterId)
    return { ...res }
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

        form.setFieldsValue(mergedData)
      } else {
        const integrationData = await getIntegrationInfo()
        form.setFieldsValue(integrationData)
      }
    }

    fetchData()
  }, [searchParams])
  return (
    <div className="flex flex-col h-full overflow-hidden">
      <div className="p-3 overflow-auto flex-1 h-full">
        <Form
          scrollToFirstError
          form={form}
          layout="vertical"
          validateMessages={{
            required: '${label}不可为空',
          }}
          initialValues={{
            clusterType: 'k8s',
            metric: {
              dsType: 'self-collector',
              mode: 'pql',
              name: 'APO-DEFAULT-VM',
            },
            log: {
              dbType: 'self-collector',
              mode: 'sql',
              name: 'APO-DEFAULT-CH',
            },
          }}
        >
          <Form.Item name="id" hidden>
            <Input />
          </Form.Item>
          <Form.Item name="name" label={t('clusterName')} rules={[{ required: true }]}>
            <Input />
          </Form.Item>
          <Form.Item name="clusterType" label={t('clusterType')} rules={[{ required: true }]}>
            <Segmented options={clusterTypeOptions} />
          </Form.Item>
          <APOCollectorFormItem />
          <Card type="inner" title={t('traceIntegration')} size="small">
            <TraceFormItem />
          </Card>
          <Card type="inner" title={t('metricsIntegration')} size="small" className="mt-2">
            <MetricsFormItem />
          </Card>
          <Card type="inner" title={t('logsIntegration')} size="small" className="mt-2">
            <LogsFormItem />
          </Card>
          <Form.Item name="traceAPI"></Form.Item>
        </Form>
      </div>
      <div className={styles.bottomDiv}>
        <Button className="mr-3">{ct('cancel')}</Button>
        <Button type="primary" onClick={saveInfo}>
          {t('saveSettings')}
        </Button>
      </div>
    </div>
  )
}
export default SettingsForm
