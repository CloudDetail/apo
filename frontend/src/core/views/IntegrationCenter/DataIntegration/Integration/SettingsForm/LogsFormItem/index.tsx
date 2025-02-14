import { Form, Input, Segmented, Switch } from 'antd'
import { useTranslation } from 'react-i18next'
import DbTypeRadio from './DbTypeRadio'

const LogsFormItem = () => {
  const { t } = useTranslation('core/dataIntegration')
  const form = Form.useFormInstance()
  const clickhouseCluster = Form.useWatch(['log', 'logAPI', 'chConfig', 'replication'], form)
  return (
    <>
      <Form.Item
        name={['log', 'dbType']}
        // label={t('dbType')}
        className="mb-0"
        valuePropName="value"
      >
        <DbTypeRadio />
      </Form.Item>
      <Form.Item
        name={['log', 'name']}
        label={t('logsName')}
        className="mb-0"
        valuePropName="value"
        hidden
      >
        <Input readOnly disabled />
      </Form.Item>
      {/* <Form.Item
        name={['log', 'mode']}
        label={t('integrationMode')}
        className="mb-0"
        valuePropName="value"
      >
        <Segmented options={['sql']} />
      </Form.Item>
      <Form.Item
        name={['log', 'name']}
        label={t('logsName')}
        className="mb-0"
        valuePropName="value"
      >
        <Input readOnly disabled />
      </Form.Item>
      <Form.Item
        name={['log', 'logAPI', 'chConfig', 'address']}
        label={t('chConfig.address')}
        className="mb-0"
        valuePropName="value"
      >
        <Input readOnly disabled />
      </Form.Item>
      <Form.Item
        name={['log', 'logAPI', 'chConfig', 'username']}
        label={t('chConfig.username')}
        className="mb-0"
        valuePropName="value"
      >
        <Input readOnly disabled />
      </Form.Item>
      <Form.Item
        name={['log', 'logAPI', 'chConfig', 'password']}
        label={t('chConfig.password')}
        className="mb-0"
        valuePropName="value"
      >
        <Input readOnly disabled />
      </Form.Item>
      <Form.Item
        name={['log', 'logAPI', 'chConfig', 'database']}
        label={t('chConfig.database')}
        className="mb-0"
        valuePropName="value"
      >
        <Input readOnly disabled />
      </Form.Item>{' '}
      <Form.Item
        name={['log', 'logAPI', 'chConfig', 'replication']}
        label={t('chConfig.replication')}
        className="mb-0"
        valuePropName="value"
      >
        <Switch checkedChildren={t('yes')} unCheckedChildren={t('no')} disabled />
      </Form.Item>
      {clickhouseCluster && (
        <Form.Item
          name={['log', 'logAPI', 'chConfig', 'cluster']}
          label={t('chConfig.cluster')}
          className="mb-0"
          valuePropName="value"
        >
          <Input readOnly disabled />
        </Form.Item>
      )} */}
    </>
  )
}
export default LogsFormItem
