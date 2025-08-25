import React from 'react'
import { Modal, Form, Input, Button } from 'antd'
import { useTranslation } from 'react-i18next'
import { ClientConfig } from '../types/index'

interface SettingsModalProps {
  open: boolean
  clientConfig: ClientConfig
  onClose: () => void
  onConfigChange: (config: ClientConfig) => void
  onSave: () => void
}

const SettingsModal: React.FC<SettingsModalProps> = ({
  open,
  clientConfig,
  onClose,
  onConfigChange,
  onSave,
}) => {
  const { t } = useTranslation('pro/agent')
  const [form] = Form.useForm()

  const handleSave = async () => {
    try {
      const values = await form.validateFields()
      onConfigChange(values)
      onSave()
    } catch (error) {
      console.error(t('settings.formValidationFailed'), error)
    }
  }

  return (
    <Modal
      title={t('settings.title')}
      open={open}
      onCancel={onClose}
      footer={[
        <Button key="cancel" onClick={onClose}>
          {t('settings.cancel')}
        </Button>,
        <Button key="save" type="primary" onClick={handleSave}>
          {t('settings.save')}
        </Button>,
      ]}
      width={500}
    >
      <Form
        form={form}
        layout="vertical"
        initialValues={clientConfig}
        onValuesChange={(_, allValues) => onConfigChange(allValues)}
      >
        <Form.Item
          label={t('settings.baseUrl')}
          name="baseUrl"
          rules={[
            { required: true, message: t('settings.baseUrlRequired') },
            { type: 'url', message: t('settings.baseUrlInvalid') },
          ]}
        >
          <Input placeholder={t('settings.baseUrlPlaceholder')} />
        </Form.Item>

        <Form.Item
          label={t('settings.apiKey')}
          name="apiKey"
          rules={[{ required: true, message: t('settings.apiKeyRequired') }]}
        >
          <Input.Password placeholder={t('settings.apiKeyPlaceholder')} />
        </Form.Item>
      </Form>
    </Modal>
  )
}

export default SettingsModal
