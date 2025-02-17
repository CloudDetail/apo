import {
  Form,
  Segmented,
  Image,
  Radio,
  Flex,
  Input,
  Typography,
  InputNumber,
  Switch,
  Select,
  Divider,
  Alert,
  Space,
} from 'antd'
import { useEffect } from 'react'
import { Trans, useTranslation } from 'react-i18next'
import { traceItems } from 'src/core/views/IntegrationCenter/constant'
import ApmTypeRadio from './ApmTypeRadio'
import { t } from 'i18next'
import { MdOutlineSpaceBar } from 'react-icons/md'
import { AiOutlineInfoCircle } from 'react-icons/ai'
const addressEg =
  'Eg. Jaeger: apo-jaeger-collector-svc:16686/jaeger，SkyWalking: skywalking-ui-svc:12800'
const apmTypeOptions = [
  {
    label: t('core/dataIntegration:datasourceApo'),
    value: 'self-collector',
  },
  {
    label: 'sidecar',
    value: 'sidecar',
  },
  {
    label: 'collector',
    value: 'collector',
  },
]
const traceApiMap = {
  skywalking: {
    key: 'skywalking',
    formItems: [
      {
        name: 'address',
        label: t('core/dataIntegration:address'),
        rules: [{ required: true }],
        placeholder: addressEg,
      },
      {
        name: 'user',
        label: t('core/dataIntegration:user'),
        secret: true,
      },
      {
        name: 'password',
        label: t('core/dataIntegration:password'),
        secret: true,
      },
    ],
  },
  jaeger: {
    key: 'jaeger',
    formItems: [
      {
        name: 'address',
        label: t('core/dataIntegration:address'),
        rules: [{ required: true }],
        placeholder: addressEg,
      },
    ],
  },
  nbs3: {
    key: 'nbs3',
    formItems: [
      {
        name: 'address',
        label: t('core/dataIntegration:address'),
        rules: [{ required: true }],
        placeholder: addressEg,
      },
      {
        name: 'user',
        label: t('core/dataIntegration:user'),
        secret: true,
      },
      {
        name: 'password',
        label: t('core/dataIntegration:password'),
        secret: true,
      },
    ],
  },
  arms: {
    key: 'arms',
    formItems: [
      {
        name: 'address',
        label: t('core/dataIntegration:address'),
        rules: [{ required: true }],
        placeholder: addressEg,
      },
      {
        name: 'accessKey',
        label: t('core/dataIntegration:accessKey'),
        rules: [{ required: true }],
        secret: true,
      },
      {
        name: 'accessSecret',
        label: t('core/dataIntegration:accessSecret'),
        rules: [{ required: true }],
        secret: true,
      },
    ],
  },
  huawei: {
    key: 'huawei',
    formItems: [
      {
        name: 'accessKey',
        label: t('core/dataIntegration:accessKey'),
        secret: true,
      },
      {
        name: 'accessSecret',
        label: t('core/dataIntegration:accessSecret'),
        secret: true,
      },
    ],
  },
  elastic: {
    key: 'elastic',
    formItems: [
      {
        name: 'address',
        label: t('core/dataIntegration:address'),
        rules: [{ required: true }],
        placeholder: addressEg,
      },
      {
        name: 'user',
        label: t('core/dataIntegration:user'),
        secret: true,
      },
      {
        name: 'password',
        label: t('core/dataIntegration:password'),
        secret: true,
      },
    ],
  },
  pinpoint: {
    key: 'pinpoint',
    formItems: [
      {
        name: 'address',
        label: t('core/dataIntegration:address'),
        rules: [{ required: true }],
        placeholder: addressEg,
      },
    ],
  },
}

const TraceFormItem = () => {
  const { t } = useTranslation('core/dataIntegration')

  const form = Form.useFormInstance()
  const apmTypeValue = Form.useWatch(['trace', 'apmType'], form)
  const modeValue = Form.useWatch(['trace', 'mode'], form)
  const traceAPI = Form.useWatch('traceAPI', form)
  const instrumentAll = Form.useWatch(['trace', 'selfCollectConfig', 'instrumentAll'], form)
  useEffect(() => {
    if (apmTypeValue === 'opentelemetry') {
      form.setFieldValue(['trace', 'mode'], 'self-collector')
    } else if (apmTypeValue && apmTypeValue !== 'opentelemetry' && modeValue === 'self-collector') {
      form.setFieldValue(['trace', 'mode'], 'sidecar')
    }
    if (modeValue === 'sidecar' && traceAPI && traceAPI[apmTypeValue]) {
      Object.entries(traceAPI[apmTypeValue]).map(([key, value]) => {
        form.setFieldValue(['trace', 'traceApi', apmTypeValue, key], value)
      })
    }
    form.setFieldValue(['trace', 'traceApi', 'timeout'], traceAPI?.timeout)
  }, [apmTypeValue, modeValue])
  return (
    <div>
      <Form.Item
        name={['trace', 'apmType']}
        label={t('apmType')}
        className="mb-0"
        valuePropName="value"
        rules={[{ required: true }]}
      >
        <ApmTypeRadio />
      </Form.Item>
      <Form.Item name={['trace', 'mode']} label={t('mode')} rules={[{ required: true }]}>
        <Segmented
          options={
            apmTypeValue === 'other'
              ? [apmTypeOptions[2]]
              : apmTypeValue === 'opentelemetry'
                ? [apmTypeOptions[0]]
                : apmTypeOptions.slice(1, 3)
          }
          defaultValue=""
        />
      </Form.Item>
      {modeValue && modeValue !== 'self-collector' && (
        <span className="text-xs text-gray-400 flex">
          <AiOutlineInfoCircle size={16} className="mr-1 " color="#1668dc" />
          {t(modeValue)}
        </span>
      )}
      {modeValue === 'sidecar' && traceApiMap[apmTypeValue] && (
        <>
          <Divider></Divider>
          <Typography.Title level={5}>Sidecar APM Config</Typography.Title>
          <Alert type="info" showIcon message={t('sidecarGlobal')} className="mb-1"></Alert>
          <div className="px-3">
            <Form.Item
              label={
                <Space>
                  {t('timeout')}
                  <span className=" text-gray-400 text-xs">{t('timeoutHint')}</span>
                </Space>
              }
              name={['trace', 'traceApi', 'timeout']}
              initialValue={15}
            >
              <InputNumber addonAfter={t('second')} />
            </Form.Item>
            {traceApiMap[apmTypeValue]?.formItems.map((item) => (
              <>
                <Form.Item
                  label={item.label}
                  name={['trace', 'traceApi', traceApiMap[apmTypeValue].key, item.name]}
                  rules={item.rules}
                >
                  {item.secret ? (
                    <Input.Password visibilityToggle={false} />
                  ) : (
                    <Input placeholder={item.placeholder} />
                  )}
                </Form.Item>
              </>
            ))}
          </div>
        </>
      )}

      {modeValue === 'self-collector' && (
        <>
          <Divider></Divider>
          <Typography.Title level={5}>APO自采配置</Typography.Title>
          <div className="px-3">
            <Form.Item
              label={t('instrumentAll')}
              name={['trace', 'selfCollectConfig', 'instrumentAll']}
            >
              <Switch checkedChildren={t('yes')} unCheckedChildren={t('no')} />
            </Form.Item>
            {!instrumentAll ? (
              <Form.Item
                label={t('instrumentNS')}
                name={['trace', 'selfCollectConfig', 'instrumentNS']}
              >
                <Select
                  tokenSeparators={[',']}
                  mode="tags"
                  style={{ width: '100%' }}
                  placeholder={
                    <Trans
                      t={t}
                      i18nKey="namespacePlaceholder"
                      components={{
                        // icon: <SiComma />,
                        span: <span className="flex items-center" />,
                      }}
                    />
                  }
                  options={[]}
                  open={false}
                  suffixIcon={null}
                />
              </Form.Item>
            ) : (
              <Form.Item
                label={
                  <Space>
                    {t('instrumentDisabledNS')}
                    <span className="text-gray-400 text-xs">{t('balckList')}</span>
                  </Space>
                }
                name={['trace', 'selfCollectConfig', 'instrumentDisabledNS']}
              >
                <Select
                  mode="tags"
                  style={{ width: '100%' }}
                  placeholder={
                    <Trans
                      t={t}
                      i18nKey="namespacePlaceholder"
                      components={{
                        icon: <MdOutlineSpaceBar />,
                        span: <span className="flex items-center" />,
                      }}
                    />
                  }
                  options={[]}
                  open={false}
                  suffixIcon={null}
                />
              </Form.Item>
            )}
          </div>
        </>
      )}
    </div>
  )
}
export default TraceFormItem
