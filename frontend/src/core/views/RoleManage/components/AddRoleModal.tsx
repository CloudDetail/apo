import React from 'react';
import { Form, Input } from 'antd';
import { useTranslation } from 'react-i18next';
import FormModal from 'src/core/components/Modal/FormModal';
import PermissionTree from 'src/core/components/PermissionTree';

interface AddRoleModalProps {
  visible: boolean;
  loading: boolean;
  onCancel: () => void;
  onFinish: (values: { roleName: string; description: string; permissionList: any }) => void;
}

export const AddRoleModal: React.FC<AddRoleModalProps> = ({
  visible,
  loading,
  onCancel,
  onFinish,
}) => {
  const { t } = useTranslation('core/roleManage');

  return (
    <FormModal
      title={t('addModal.title')}
      open={visible}
      onCancel={onCancel}
      confirmLoading={loading}
    >
      <FormModal.Section onFinish={onFinish}>
        <Form.Item
          name="roleName"
          label={t('index.roleName')}
          rules={[{ required: true, message: t('addModal.roleNameRequired') }]}
        >
          <Input />
        </Form.Item>
        <Form.Item name="description" label={t('index.description')}>
          <Input />
        </Form.Item>
        <Form.Item label={t('addModal.permissions')} name="permissionList">
          <PermissionTree
            subjectType="role"
            hasSaveButton={false}
            style={{ height: 'calc(100vh - 240px)', overflow: 'auto' }}
            actionsLocation="top"
            actionStyle={{ justifyContent: 'flex-start' }}
            styles={{ body: { padding: '0 8px' } }}
          />
        </Form.Item>
      </FormModal.Section>
    </FormModal>
  );
};