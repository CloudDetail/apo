import React, { ReactNode } from 'react';
import { Form, FormInstance, Space, Button, Flex } from 'antd';
import CommonModal from './CommonModal';
import { useTranslation } from 'react-i18next';

interface FormModalProps {
  title: string;
  open: boolean;
  onCancel: () => void;
  width?: number | string;
  footer?: ReactNode | null;
  okText?: string;
  cancelText?: string;
  confirmLoading?: boolean;
  children: ReactNode;
}

interface FormSectionProps {
  form?: FormInstance;
  onFinish?: (values: any) => void;
  onCancel?: () => void;
  initialValues?: Record<string, any>;
  children: ReactNode;
  formProps?: Record<string, any>;
  okText?: string;
  cancelText?: string;
  confirmLoading?: boolean;
}

// 表单区域组件
const FormSection: React.FC<FormSectionProps> = ({
  form: externalForm,
  onFinish,
  initialValues,
  children,
  formProps,
  okText,
  cancelText,
  confirmLoading
}) => {
  const [form] = Form.useForm(externalForm);
  const { t } = useTranslation('common');

  return (
    <Form
      form={form}
      layout="horizontal"
      autoComplete="off"
      onFinish={onFinish}
      initialValues={initialValues}
      {...formProps}
    >
      {children}
      <Flex justify="end" gap="small">
        <Button onClick={() => form.resetFields()}>
          {cancelText || t('reset')}
        </Button>
        <Button type="primary" htmlType="submit" loading={confirmLoading}>
          {okText || t('confirm')}
        </Button>
      </Flex>
    </Form>
  );
};

// 为FormModal添加静态属性
const FormModal: React.FC<FormModalProps> & { Section: typeof FormSection } = ({
  title,
  open,
  onCancel,
  width,
  children
}) => {
  return (
    <CommonModal
      title={title}
      open={open}
      onCancel={onCancel}
      width={width}
      footer={null}
    >
      {children}
    </CommonModal>
  );
};

FormModal.Section = FormSection;

export default FormModal;