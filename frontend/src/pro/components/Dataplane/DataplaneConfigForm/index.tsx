import { Alert, Button, Form, Input } from 'antd'
import {
  checkDataplaneDatasourceApi,
  createDataplaneApi,
  updateDataplaneApi,
} from 'src/pro/api/dataplane'
import { useEffect, useMemo, useState } from 'react'
import { DataplaneItem, DataplaneTypeItem } from 'src/pro/types/datasource'
import { notify } from 'src/core/utils/notify'
import LogsFormItem from 'src/core/components/Integration/LogsFormItem'
import MetricsFormItem from 'src/core/components/Integration/MetricsFormItem'
import TraceFormItem from 'src/core/components/Integration/TraceFormItem'
import { useTranslation } from 'react-i18next'
import i18n from 'src/i18n'

const DataplaneConfigForm = ({
  dataplaneParams,
  onSave,
  readonly = false,
}: {
  dataplaneParams: DataplaneTypeItem | DataplaneItem
  onSave?: (apoConfig: any) => void
  readonly: boolean
}) => {
  const { t } = useTranslation('core/dataIntegration')
  const { t: tAgent } = useTranslation('pro/agentDatasource')

  const { paramSpec, typeName, capabilitySpec, desc } = dataplaneParams
  const capability = 'capabilitySpec' in dataplaneParams ? dataplaneParams.capabilitySpec : []
  const params = 'params' in dataplaneParams ? dataplaneParams.params : '{}'
  const name = 'name' in dataplaneParams ? dataplaneParams.name : ''
  const id = 'id' in dataplaneParams ? dataplaneParams.id : undefined
  const [form] = Form.useForm()
  const [checkLoading, setCheckLoading] = useState(false)
  const [checkResult, setCheckResult] = useState<any>(null)
  const saveDataplane = (values: any) => {
    values.params = JSON.stringify(values.params)
    delete values.trace
    delete values.log
    delete values.metric
    const api = id ? updateDataplaneApi : createDataplaneApi
    api(values)
      .then((res) => {
        notify({
          message: tAgent('messages.connectSuccess'),
          type: 'success',
        })
      })
      .finally(() => {
        onSave?.(values)
      })
  }
  const saveForm = () => {
    form.validateFields().then((values) => {
      const info: Record<string, any> = {}
      dataSourceConfigs.forEach(({ key }) => {
        info[key] = values[key]
      })

      saveDataplane(values)
    })
  }
  const testForm = () => {
    setCheckLoading(true)
    form
      .validateFields()
      .then((values) => {
        checkDataplaneDatasourceApi({
          dataSource: values.typ,
          attributes: JSON.stringify(values.params),
          enabledCapabilities: capabilitySpec,
        })
          .then((res) => {
            setCheckResult(res)
          })
          .finally(() => {
            setCheckLoading(false)
          })
      })
      .catch(() => {
        setCheckLoading(false)
      })
  }
  useEffect(() => {
    form.resetFields()
    form.setFieldsValue({
      id: id,
      typ: typeName,
      capability: capability || capabilitySpec,
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

  // 表单初始值
  const initialValues = useMemo(
    () => ({
      capability: capabilitySpec || [],
    }),
    [capabilitySpec],
  )
  return (
    <div className="flex flex-col h-full">
      {checkResult && checkResult.message && (
        <Alert
          closable
          onClose={() => setCheckResult(null)}
          showIcon
          message={tAgent('messages.testFailed')}
          description={
            <>
              <span className="text-[var(--ant-color-text-secondary)]">{checkResult.message}</span>
            </>
          }
          type="error"
        />
      )}
      {checkResult && checkResult.success && (
        <Alert
          closable
          onClose={() => setCheckResult(null)}
          showIcon
          message={tAgent('messages.testSuccess')}
          type="success"
        />
      )}

      {/* <Collapse
        defaultActiveKey={['1']}
        className="my-2"
        items={[
          {
            key: '1',
            label: typeName + ' ' + tAgent('desc'),
            children: <MarkdownRender content={desc} theme="light" />,
          },
        ]}
      /> */}
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
          scrollToFirstError={{ behavior: 'instant', block: 'end', focus: true }}
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
          <Form.Item name="note" label={tAgent('form.note')}>
            <Input.TextArea rows={2} />
          </Form.Item>
          {paramSpec?.children?.map((item) => (
            <Form.Item
              key={item.name}
              name={['params', item.name]}
              label={
                <div>
                  {/* {item.name}
                  <span className="ml-2 text-xs text-[var(--ant-color-text-secondary)]"> */}
                  {i18n.language === 'zh' ? item.desc : item.desc_en}
                  {/* </span> */}
                </div>
              }
              rules={[{ required: !item?.optional }]}
            >
              <Input />
            </Form.Item>
          ))}

          <Form.Item name="capability" label="capability" hidden>
            {/* <Checkbox.Group
                  options={capabilitySpec?.map((item) => ({ label: item, value: item }))}
                /> */}
            {/* <Capability capability={capabilitySpec || []} /> */}
          </Form.Item>
        </Form>
      )}
      {!readonly && (
        <div className="flex justify-end  flex-0 items-end">
          <Button
            onClick={testForm}
            className="mr-2"
            color="primary"
            variant="outlined"
            loading={checkLoading}
          >
            {tAgent('form.test')}
          </Button>
          <Button type="primary" onClick={saveForm}>
            {tAgent('form.connect')}
          </Button>
        </div>
      )}
    </div>
  )
}

export default DataplaneConfigForm
