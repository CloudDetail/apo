/**
 * Copyright 2025 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import React, { ReactNode } from 'react'
import { Form, FormInstance, Button, Flex } from 'antd'
import CommonModal from './CommonModal'
import { useTranslation } from 'react-i18next'

interface FormModalProps {
  title: string
  open: boolean
  onCancel: () => void
  width?: number | string
  footer?: ReactNode | null
  okText?: string
  cancelText?: string
  confirmLoading?: boolean
  children: ReactNode
}

interface FormSectionProps {
  form?: FormInstance
  onFinish?: (values: any) => void
  onCancel?: () => void
  initialValues?: Record<string, any>
  children: ReactNode
  formProps?: Record<string, any>
  okText?: string
  cancelText?: string
  confirmLoading?: boolean
}

const FormSection: React.FC<FormSectionProps> = ({
  form: externalForm,
  onFinish,
  initialValues,
  children,
  formProps,
  okText,
  cancelText,
  confirmLoading,
}) => {
  const [form] = Form.useForm(externalForm)
  const { t, i18n } = useTranslation('common')

  return (
    <Form
      form={form}
      layout="horizontal"
      autoComplete="off"
      onFinish={onFinish}
      initialValues={initialValues}
      labelCol={{ span: i18n.language === 'zh' ? 4 : 4 }}
      wrapperCol={{ span: 18 }}
      {...formProps}
    >
      {children}
      <Flex justify="end" gap="small">
        <Button onClick={() => form.resetFields()}>{cancelText || t('reset')}</Button>
        <Button type="primary" htmlType="submit" loading={confirmLoading}>
          {okText || t('confirm')}
        </Button>
      </Flex>
    </Form>
  )
}

// Add static attributes for FormModal
const FormModal: React.FC<FormModalProps> & { Section: typeof FormSection } = ({
  title,
  open,
  onCancel,
  width = '80%',
  children,
}) => {
  return (
    <CommonModal title={title} open={open} onCancel={onCancel} width={width} footer={null}>
      {children}
    </CommonModal>
  )
}

FormModal.Section = FormSection

export default FormModal
