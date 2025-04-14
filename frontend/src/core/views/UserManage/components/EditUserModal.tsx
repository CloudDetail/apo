import React from 'react';
import { Form, Input, Select, Flex, Divider } from 'antd';
import { useTranslation } from 'react-i18next';
import FormModal from 'src/core/components/Modal/FormModal';
import { User } from 'src/core/types/user';

interface EditUserModalProps {
  visible: boolean;
  user: User | null;
  roleItems: Array<{ label: string; value: string | number }>;
  onCancel: () => void;
  onFinish: (values: {
    role: string | number;
    corporation?: string;
    email?: string;
    phone?: string;
  }) => Promise<void>;
  onResetPassword: (values: {
    newPassword: string;
    confirmPassword: string;
  }) => Promise<void>;
}

export const EditUserModal: React.FC<EditUserModalProps> = ({
  visible,
  user,
  roleItems,
  onCancel,
  onFinish,
  onResetPassword,
}) => {
  const { t } = useTranslation('core/userManage');

  return (
    <FormModal
      title={t('editModal.title')}
      open={visible}
      onCancel={onCancel}
      footer={null}
    >
      <FormModal.Section
        onFinish={onFinish}
        initialValues={user}
      >
        <Flex gap={16} className="mb-6">
          <Form.Item
            name="username"
            labelCol={{ span: 8 }}
            label={t('index.userName')}
            rules={[{ required: true, message: t('editModal.userNameRequired') }]}
            style={{ marginBottom: 0, flex: 1 }}
          >
            <Input disabled />
          </Form.Item>
          <Form.Item
            name="roleId"
            labelCol={{ span: 8 }}
            label={t('index.role')}
            initialValue={user?.roleList[0]?.roleId}
            rules={[{ required: true, message: t('editModal.selectRole') }]}
            style={{ marginBottom: 0, flex: 1 }}
          >
            <Select options={roleItems} />
          </Form.Item>
        </Flex>
        <Form.Item
          name="corporation"
          label={t('index.corporation')}
        >
          <Input />
        </Form.Item>
        <Form.Item
          name="email"
          label={t('index.email')}
          rules={[{ type: 'email', message: t('editModal.emailInvalid') }]}
        >
          <Input placeholder={t('editModal.emailPlaceholder')} />
        </Form.Item>
        <Form.Item
          name="phone"
          label={t('index.phone')}
          // Regular expression of Chinese mainland mobile phone number
          rules={[{ pattern: /^1[3-9]\d{9}$/, message: t('editModal.phoneInvalid') }]}
        >
          <Input placeholder={t('editModal.phonePlaceholder')} />
        </Form.Item>
      </FormModal.Section>
      <Divider />
      <FormModal.Section
        onFinish={onResetPassword}
      >
        <Form.Item
          label={t('editModal.newPassword')}
          name="newPassword"
          rules={[
            {
              required: true,
              message: t('editModal.newPasswordPlaceholder'),
            },
            {
              pattern: /^(?=.*[a-z])(?=.*[A-Z])(?=.*[!@#$%^&*()\-_+=<>?/{}[\]|:;.,~]).{9,}$/,
              message: t('editModal.newPasswordPattern'),
            },
          ]}
        >
          <Input.Password placeholder={t('editModal.newPasswordPlaceholder')} />
        </Form.Item>
        <Form.Item
          label={t('editModal.confirmPassword')}
          name="confirmPassword"
          rules={[
            { required: true, message: t('editModal.confirmPasswordRequired') },
            ({ getFieldValue }) => ({
              validator(_, value) {
                if (!value || getFieldValue('newPassword') === value) {
                  return Promise.resolve()
                }
                return Promise.reject(new Error(t('editModal.confirmPasswordMismatch')))
              },
            }),
          ]}
        >
          <Input.Password placeholder={t('editModal.confirmPasswordPlaceholder')} />
        </Form.Item>
      </FormModal.Section>
    </FormModal>
  );
};