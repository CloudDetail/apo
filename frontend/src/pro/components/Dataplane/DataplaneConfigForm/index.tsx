import { Button, Checkbox, Collapse, Form, Input } from 'antd'
import { createDataplaneApi, updateDataplaneApi } from 'src/pro/api/dataplane'
import { useCallback, useEffect, useMemo, useState } from 'react'
import { DataplaneItem, DataplaneTypeItem } from 'src/pro/types/datasource'
import { notify } from 'src/core/utils/notify'
import LogsFormItem from 'src/core/components/Integration/LogsFormItem'
import MetricsFormItem from 'src/core/components/Integration/MetricsFormItem'
import TraceFormItem from 'src/core/components/Integration/TraceFormItem'
import { useTranslation } from 'react-i18next'
type ConfigType = 'trace' | 'metric' | 'log'

const DataplaneConfigForm = ({
  dataplaneParams,
  onSave,
  readonly = false,
  showApoConfig = false,
}: {
  dataplaneParams: DataplaneTypeItem | DataplaneItem
  onSave?: (apoConfig: any) => void
  readonly: boolean
  showApoConfig: boolean
}) => {
  const { t } = useTranslation('core/dataIntegration')
  const { t: tAgent } = useTranslation('pro/agentDatasource')

  const { paramSpec, typeName, capabilitySpec } = dataplaneParams
  const capability = 'capability' in dataplaneParams ? dataplaneParams.capability : []
  const params = 'params' in dataplaneParams ? dataplaneParams.params : '{}'
  const name = 'name' in dataplaneParams ? dataplaneParams.name : ''
  const id = 'id' in dataplaneParams ? dataplaneParams.id : undefined
  const [form] = Form.useForm()
  const [apoConfig, setApoConfig] = useState<ConfigType[]>([])

  const saveDataplane = (values: any) => {
    values.params = JSON.stringify(values.params)
    delete values.trace
    delete values.log
    delete values.metric
    const api = id ? updateDataplaneApi : createDataplaneApi
    api(values).then((res) => {
      console.log(res)
      notify({
        message: tAgent('messages.connectSuccess'),
        type: 'success',
      })
    })
  }
  const saveForm = () => {
    form.validateFields().then((values) => {
      const info: Record<string, any> = {}
      console.log(values)
      dataSourceConfigs.forEach(({ key }) => {
        if (!apoConfig?.includes(key)) {
          info[key] = values[key]
          // delete values[key]
        }
      })

      onSave?.(info)
      saveDataplane(values)
    })
  }
  useEffect(() => {
    form.resetFields()
    form.setFieldsValue({
      id: id,
      typ: typeName,
      capability: capability || [],
      params: JSON.parse(params || '{}'),
      name: name || '',
    })
  }, [typeName, dataplaneParams])
  const dataSourceConfigs = useMemo(
    () => [
      {
        key: 'trace' as const,
        label: t('traceIntegration'),
        component: TraceFormItem,
      },
      {
        key: 'metric' as const,
        label: t('metricsIntegration'),
        component: MetricsFormItem,
      },
      {
        key: 'log' as const,
        label: t('logsIntegration'),
        component: LogsFormItem,
      },
    ],
    [t],
  )
  // 处理配置切换
  const handleConfigToggle = useCallback((type: ConfigType) => {
    setApoConfig((prev) =>
      prev.includes(type) ? prev.filter((item) => item !== type) : [...prev, type],
    )
  }, [])
  // Collapse 额外内容组件
  const CollapseExtra = useCallback(
    ({ type }: { type: ConfigType }) => (
      <Checkbox
        checked={apoConfig.includes(type)}
        onClick={(e) => {
          e.stopPropagation()
          handleConfigToggle(type)
        }}
      >
        {tAgent('form.config')}
      </Checkbox>
    ),
    [apoConfig, handleConfigToggle],
  )
  // 表单初始值
  const initialValues = useMemo(
    () => ({
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
      apoCollector: {
        collectorAddr: 'apo-nginx-proxy-svc',
      },
    }),
    [],
  )
  return (
    <div className="flex flex-col h-full">
      {dataplaneParams && (
        <Form
          layout="vertical"
          className="mx-4 flex-1 h-0 overflow-y-auto"
          initialValues={{
            typ: typeName,
            ...initialValues,
          }}
          form={form}
          disabled={readonly}
        >
          <Form.Item name="id" hidden>
            <Input />
          </Form.Item>
          <Form.Item name="typ" hidden rules={[{ required: true }]}>
            <Input defaultValue={typeName} />
          </Form.Item>
          <Form.Item name="name" label={tAgent('form.dataSourceName')} rules={[{ required: true }]}>
            <Input />
          </Form.Item>
          {paramSpec?.children?.map((item) => (
            <Form.Item
              key={item.name}
              name={['params', item.name]}
              label={item.name}
              rules={[{ required: true }]}
            >
              <Input />
            </Form.Item>
          ))}

          <Form.Item name="capability" label="capability" rules={[{ required: true }]}>
            <Checkbox.Group
              options={capabilitySpec?.map((item) => ({ label: item, value: item }))}
            />
          </Form.Item>

          {showApoConfig &&
            dataSourceConfigs.map(({ key, label, component: Component }) => (
              <Collapse
                key={key}
                className="mb-2"
                items={[
                  {
                    forceRender: true,
                    key,
                    label,
                    children: <Component />,
                    extra: <CollapseExtra type={key} />,
                  },
                ]}
              />
            ))}
        </Form>
      )}
      {!readonly && (
        <div className="flex justify-end  flex-0">
          <Button type="primary" onClick={saveForm}>
            {tAgent('form.connect')}
          </Button>
        </div>
      )}
    </div>
  )
}

export default DataplaneConfigForm
