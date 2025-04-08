import React, { ReactNode } from 'react';
import { Form, FormInstance, Space, Button } from 'antd';
import CommonModal from '../CommonModal';

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
  initialValues?: Record<string, any>;
  children: ReactNode;
  formProps?: Record<string, any>;
}

// 表单区域组件
const FormSection: React.FC<FormSectionProps> = ({
  form: externalForm,
  onFinish,
  initialValues,
  children,
  formProps
}) => {
  const [form] = Form.useForm(externalForm);

  return (
    <Form
      form={form}
      layout="vertical"
      onFinish={onFinish}
      initialValues={initialValues}
      {...formProps}
    >
      {children}
      <Space>
        <Button type="primary" htmlType="submit">
          保存
        </Button>
        <Button onClick={() => form.resetFields()}>
          重置
        </Button>
      </Space>
    </Form>
  );
};

// 为FormModal添加静态属性
const FormModal: React.FC<FormModalProps> & { Section: typeof FormSection } = ({
  title,
  open,
  onCancel,
  width,
  footer,
  okText,
  cancelText,
  confirmLoading = false,
  children
}) => {
  return (
    <CommonModal
      title={title}
      open={open}
      onCancel={onCancel}
      width={width}
      footer={footer}
      okText={okText}
      cancelText={cancelText}
      confirmLoading={confirmLoading}
    >
      {children}
    </CommonModal>
  );
};

FormModal.Section = FormSection;

export default FormModal;