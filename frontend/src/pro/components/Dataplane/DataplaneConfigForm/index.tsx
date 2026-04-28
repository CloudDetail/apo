import { Alert, Button, Form, Input } from 'antd'
import {
  checkDataplaneDatasourceApi,
  createDataplaneApi,
  updateDataplaneApi,
} from 'src/pro/api/dataplane'
import { useEffect, useMemo, useState } from 'react'
import { DataplaneItem, DataplaneTypeItem } from 'src/pro/types/datasource'
import { notify } from 'src/core/utils/notify'
import { useTranslation } from 'react-i18next'
import i18n from 'src/i18n'
import SetupGuide from '../DataplaneGuide'

const ELASTIC_LOG_ONLY_CAPABILITIES = ['Log']
const ELASTIC_DEFAULT_LOG_INDEX_PATTERN = 'logs-*'
const ELASTIC_PRIMARY_PARAM_NAMES = ['address', 'apiKey']

const isDataplaneItem = (
  dataplaneParams: DataplaneTypeItem | DataplaneItem,
): dataplaneParams is DataplaneItem => 'capability' in dataplaneParams

const parseParams = (params: string) => {
  try {
    return JSON.parse(params || '{}')
  } catch {
    return {}
  }
}

const normalizeElasticFormValues = (values: any) => {
  const rawParams = values?.params || {}
  const params: Record<string, string> = {}

  ELASTIC_PRIMARY_PARAM_NAMES.forEach((name) => {
    const value = rawParams[name]
    if (typeof value === 'string' && value.trim()) {
      params[name] = value.trim()
    }
  })
  params.authType = 'apiKey'
  params.logIndexPattern = ELASTIC_DEFAULT_LOG_INDEX_PATTERN

  return {
    ...values,
    capability: ELASTIC_LOG_ONLY_CAPABILITIES,
    params,
  }
}

const DataplaneConfigForm = ({
  dataplaneParams,
  onSave,
  readonly = false,
}: {
  dataplaneParams: DataplaneTypeItem | DataplaneItem
  onSave?: (apoConfig: any) => void
  readonly: boolean
}) => {
  const { t: tAgent } = useTranslation('pro/agentDatasource')

  const { paramSpec, typeName, capabilitySpec } = dataplaneParams
  const isElastic = typeName?.toLowerCase() === 'elastic'
  const capability = isElastic ? ELASTIC_LOG_ONLY_CAPABILITIES : capabilitySpec || []
  const params = isDataplaneItem(dataplaneParams) ? dataplaneParams.params : '{}'
  const name = isDataplaneItem(dataplaneParams) ? dataplaneParams.name : ''
  const id = isDataplaneItem(dataplaneParams) ? dataplaneParams.id : undefined
  const [form] = Form.useForm()
  const [checkLoading, setCheckLoading] = useState(false)
  const [checkResult, setCheckResult] = useState<any>(null)
  const saveDataplane = (values: any) => {
    const payload = {
      ...values,
      params: JSON.stringify(values.params || {}),
    }
    delete payload.trace
    delete payload.log
    delete payload.metric
    const api = id ? updateDataplaneApi : createDataplaneApi
    api(payload)
      .then((res) => {
        notify({
          message: tAgent('messages.connectSuccess'),
          type: 'success',
        })
      })
      .finally(() => {
        onSave?.(payload)
      })
  }
  const saveForm = () => {
    form.validateFields().then((values) => {
      saveDataplane(isElastic ? normalizeElasticFormValues(values) : values)
    })
  }
  const testForm = () => {
    setCheckLoading(true)
    form
      .validateFields()
      .then((values) => {
        const normalizedValues = isElastic ? normalizeElasticFormValues(values) : values
        checkDataplaneDatasourceApi({
          dataSource: normalizedValues.typ,
          attributes: JSON.stringify(normalizedValues.params || {}),
          enabledCapabilities: normalizedValues.capability || capabilitySpec,
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
    const parsedParams = parseParams(params)
    if (isElastic && !parsedParams.authType) {
      parsedParams.authType = 'apiKey'
    }
    if (isElastic && !parsedParams.logIndexPattern) {
      parsedParams.logIndexPattern = ELASTIC_DEFAULT_LOG_INDEX_PATTERN
    }
    form.resetFields()
    form.setFieldsValue({
      id: id,
      typ: typeName,
      capability: capability || capabilitySpec,
      params: parsedParams,
      name: name || '',
    })
  }, [typeName, dataplaneParams])

  // 表单初始值
  const initialValues = useMemo(
    () => ({
      capability: capability || [],
    }),
    [capability],
  )

  const shouldRenderParam = (name: string) => {
    if (!isElastic) {
      return true
    }
    return ELASTIC_PRIMARY_PARAM_NAMES.includes(name)
  }

  const isParamRequired = (item: any) => {
    if (!isElastic) {
      return !item?.optional
    }
    return ELASTIC_PRIMARY_PARAM_NAMES.includes(item.name)
  }

  const renderParamInput = (name: string) => {
    if (name === 'apiKey') {
      return <Input.Password />
    }
    return <Input />
  }

  const getParamLabel = (item: any) => {
    if (isElastic && item.name === 'address') {
      return 'URL'
    }
    if (isElastic && item.name === 'apiKey') {
      return 'API Key'
    }
    return i18n.language === 'zh' ? item.desc : item.desc_en
  }

  return (
    <div className="flex h-full gap-2">
      <div className="flex-1 flex flex-col h-full overflow-y-hidden w-1/2  px-4">
        <div className="flex flex-col gap-2 w-full flex-1 ">
          {checkResult && !checkResult.success && checkResult.message && (
            <Alert
              closable
              onClose={() => setCheckResult(null)}
              showIcon
              message={tAgent('messages.testFailed')}
              description={
                <>
                  <span className="text-[var(--ant-color-text-secondary)]">
                    {checkResult.message}
                  </span>
                </>
              }
              type="error"
            />
          )}
          {checkResult && checkResult.success && checkResult.message && (
            <Alert
              closable
              onClose={() => setCheckResult(null)}
              showIcon
              message={tAgent('messages.testWarning')}
              description={
                <>
                  <span className="text-[var(--ant-color-text-secondary)]">
                    {checkResult.message}
                  </span>
                </>
              }
              type="warning"
            />
          )}
          {checkResult && checkResult.success && !checkResult.message && (
            <Alert
              closable
              onClose={() => setCheckResult(null)}
              showIcon
              message={tAgent('messages.testSuccess')}
              type="success"
            />
          )}
          {typeName?.toLowerCase() === 'signoz' && (
            <Alert showIcon message={tAgent('warnings.signozMissingFdMetric')} type="warning" />
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
              className="flex-1 h-0 overflow-y-auto"
              initialValues={{
                typ: typeName,
                ...initialValues,
              }}
              form={form}
              disabled={readonly}
              scrollToFirstError={{ behavior: 'instant', block: 'end', focus: true }}
              size="small"
            >
              <Form.Item name="id" hidden>
                <Input />
              </Form.Item>
              <Form.Item name="typ" hidden rules={[{ required: true }]}>
                <Input defaultValue={typeName} />
              </Form.Item>
              <Form.Item
                name="name"
                label={tAgent('form.dataSourceName')}
                rules={[{ required: true }]}
              >
                <Input />
              </Form.Item>
              {/* <Form.Item name="note" label={tAgent('form.note')}>
                <Input.TextArea rows={2} />
              </Form.Item> */}
              {paramSpec?.children?.filter((item) => shouldRenderParam(item.name)).map((item) => (
                <Form.Item
                  key={item.name}
                  name={['params', item.name]}
                  label={
                    <div>
                      {/* {item.name}
                  <span className="ml-2 text-xs text-[var(--ant-color-text-secondary)]"> */}
                      {getParamLabel(item)}
                      {/* </span> */}
                    </div>
                  }
                  rules={[
                    {
                      required: isParamRequired(item),
                      message: tAgent('messages.fieldRequired'),
                    },
                  ]}
                >
                  {renderParamInput(item.name)}
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
        </div>
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
      <SetupGuide typeName={typeName} />
    </div>
  )
}

export default DataplaneConfigForm
