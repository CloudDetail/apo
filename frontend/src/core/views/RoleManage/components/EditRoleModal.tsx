import React from 'react';
import { Form, Input } from 'antd';
import { useTranslation } from 'react-i18next';
import FormModal from 'src/core/components/Modal/FormModal';
import { Role } from 'src/core/types/role';

interface EditRoleModalProps {
  visible: boolean;
  loading: boolean;
  selectedRole: Role | null;
  onCancel: () => void;
  onFinish: (values: { roleName: string; description: string }) => void;
}

export const EditRoleModal: React.FC<EditRoleModalProps> = ({
  visible,
  loading,
  selectedRole,
  onCancel,
  onFinish,
}) => {
  const { t } = useTranslation('core/roleManage');

  return (
    <FormModal
      title={t('editModal.title')}
      open={visible}
      onCancel={onCancel}
      confirmLoading={loading}
    >
      <FormModal.Section
        onFinish={onFinish}
        initialValues={selectedRole ? { roleName: selectedRole.roleName, description: selectedRole.description } : {}}
      >
        <Form.Item
          name="roleName"
          label={t('index.roleName')}
          rules={[{ required: true, message: t('index.roleNameRequired') }]}
        >
          <Input disabled={selectedRole?.roleName === 'admin'} />
        </Form.Item>
        <Form.Item name="description" label={t('index.description')}>
          <Input />
        </Form.Item>
      </FormModal.Section>
    </FormModal>
  );
};