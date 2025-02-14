import { Form, Input, Segmented } from 'antd'
import { useTranslation } from 'react-i18next'
import DsTypeRadio from './DsTypeRadio'

const MetricsFormItem = () => {
  const { t } = useTranslation('core/dataIntegration')
  return (
    <>
      <Form.Item
        name={['metric', 'dsType']}
        // label={t('dsType')}
        className="mb-0"
        valuePropName="value"
      >
        <DsTypeRadio />
      </Form.Item>
      <Form.Item
        name={['metric', 'name']}
        label={t('metricsName')}
        className="mb-0"
        valuePropName="value"
        hidden
      >
        <Input readOnly disabled />
      </Form.Item>
      {/* <Form.Item
        name={['metric', 'mode']}
        label={t('integrationMode')}
        className="mb-0"
        valuePropName="value"
      >
        <Segmented options={['pql']} />
      </Form.Item>
      <Form.Item
        name={['metric', 'name']}
        label={t('metricsName')}
        className="mb-0"
        valuePropName="value"
      >
        <Input readOnly disabled />
      </Form.Item>
      <Form.Item
        name={['metric', 'metricAPI', 'vmConfig', 'serverURL']}
        label={t('vmConfig.serverURL')}
        className="mb-0"
        valuePropName="value"
      >
        <Input readOnly disabled />
      </Form.Item>
      <Form.Item
        name={['metric', 'metricAPI', 'vmConfig', 'username']}
        label={t('vmConfig.username')}
        className="mb-0"
        valuePropName="value"
      >
        <Input readOnly disabled />
      </Form.Item>
      <Form.Item
        name={['metric', 'metricAPI', 'vmConfig', 'password']}
        label={t('vmConfig.password')}
        className="mb-0"
        valuePropName="value"
      >
        <Input readOnly disabled />
      </Form.Item> */}
    </>
  )
}
export default MetricsFormItem
