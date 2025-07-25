/**
 * Copyright 2025 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import { Form, Input, Segmented } from 'antd'
import { useTranslation } from 'react-i18next'
import DbTypeRadio from './DbTypeRadio'
import { AiOutlineInfoCircle } from 'react-icons/ai'

const LogsFormItem = () => {
  const { t } = useTranslation('core/dataIntegration')
  const form = Form.useFormInstance()
  // const clickhouseCluster = Form.useWatch(['log', 'logAPI', 'chConfig', 'replication'], form)
  const logCollectModeValue = Form.useWatch(['log', 'selfCollectConfig', 'mode'], form)
  const logCollectModeOptions = [
    {
      label: t('logCollect.full'),
      value: 'full',
    },
    {
      label: t('logCollect.sample'),
      value: 'sample',
    },
  ]
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
      <Form.Item name={['log', 'name']} label={t('logsName')} className="mb-0" hidden>
        <Input readOnly disabled />
      </Form.Item>
      <Form.Item
        name={['log', 'selfCollectConfig', 'mode']}
        label={t('logCollectMode')}
        className="mb-0"
        rules={[{ required: true }]}
        initialValue={'full'}
      >
        <Segmented options={logCollectModeOptions} />
      </Form.Item>
      {logCollectModeValue === 'sample' && (
        <span className="text-xs text-[var(--ant-color-text-secondary)] flex mt-1">
          <AiOutlineInfoCircle size={16} className="mr-1 " color="#1668dc" />
          {t('sampleHint')}
        </span>
      )}
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
