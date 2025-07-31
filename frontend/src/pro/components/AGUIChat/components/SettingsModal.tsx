import React from 'react'
import { Modal, Form, Input, Button } from 'antd'
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
  const [form] = Form.useForm()

  const handleSave = async () => {
    try {
      const values = await form.validateFields()
      onConfigChange(values)
      onSave()
    } catch (error) {
      console.error('表单验证失败:', error)
    }
  }

  return (
    <Modal
      title="AG-UI 设置"
      open={open}
      onCancel={onClose}
      footer={[
        <Button key="cancel" onClick={onClose}>
          取消
        </Button>,
        <Button key="save" type="primary" onClick={handleSave}>
          保存
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
          label="Base URL"
          name="baseUrl"
          rules={[
            { required: true, message: '请输入Base URL' },
            { type: 'url', message: '请输入有效的URL' },
          ]}
        >
          <Input placeholder="请输入Base URL" />
        </Form.Item>

        <Form.Item
          label="API Key"
          name="apiKey"
          rules={[{ required: true, message: '请输入API Key' }]}
        >
          <Input.Password placeholder="请输入API Key" />
        </Form.Item>
      </Form>
    </Modal>
  )
}

export default SettingsModal
