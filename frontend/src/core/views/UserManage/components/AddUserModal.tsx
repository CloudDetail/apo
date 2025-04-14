/**
 * Copyright 2025 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import React from 'react';
import { Form, Input, Select, Flex } from 'antd';
import { useTranslation } from 'react-i18next';
import FormModal from 'src/core/components/Modal/FormModal';

interface AddUserModalProps {
  visible: boolean;
  loading: boolean;
  roleItems: Array<{ label: string; value: string | number }>;
  onCancel: () => void;
  onFinish: (values: {
    username: string;
    password: string;
    confirmPassword: string;
    roleId: string | number;
    corporation?: string;
    email?: string;
    phone?: string;
  }) => Promise<void>;
}

export const AddUserModal: React.FC<AddUserModalProps> = ({
  visible,
  loading,
  roleItems,
  onCancel,
  onFinish,
}) => {
  const { t } = useTranslation('core/userManage');

  return (
    <FormModal
      title={t('index.addUser')}
      open={visible}
      onCancel={onCancel}
      confirmLoading={loading}
      footer={null}
    >
      <FormModal.Section onFinish={onFinish}>
        <Flex gap={8} className="mb-6">
          <Form.Item
            name="username"
            labelCol={{ span: 8 }}
            label={t('index.userName')}
            rules={[{ required: true, message: t('index.userNameRequired') }]}
            style={{ marginBottom: 0, flex: 1 }}
          >
            <Input autoComplete="new-user" />
          </Form.Item>
          <Form.Item
            name="roleId"
            labelCol={{ span: 8 }}
            label={t('index.role')}
            rules={[{ required: true, message: t('addModal.roleRequired') }]}
            style={{ marginBottom: 0, flex: 1 }}
          >
            <Select
              placeholder={t('addModal.selectRole')}
              options={roleItems}
            />
          </Form.Item>
        </Flex>
        <Form.Item
          name="password"
          label={t('addModal.password')}
          rules={[{ required: true, message: t('addModal.passwordRequired') }]}
        >
          <Input.Password autoComplete="new-password" />
        </Form.Item>
        <Form.Item
          label={t('addModal.confirmPassword')}
          name="confirmPassword"
          rules={[
            { required: true, message: t('addModal.confirmPasswordRequired') },
            ({ getFieldValue }) => ({
              validator(_, value) {
                if (!value || getFieldValue('password') === value) {
                  return Promise.resolve()
                }
                return Promise.reject(new Error(t('addModal.confirmPasswordMismatch')))
              },
            }),
          ]}
        >
          <Input.Password placeholder={t('addModal.confirmPasswordPlaceholder')} />
        </Form.Item>
        <Form.Item
          name="corporation"
          label={t('index.corporation')}
        >
          <Input />
        </Form.Item>
        <Form.Item
          name="email"
          label={t('index.email')}
          rules={[{ type: 'email', message: t('addModal.emailInvalid') }]}
        >
          <Input />
        </Form.Item>
        <Form.Item
          name="phone"
          label={t('index.phone')}
          // Regular expression of Chinese mainland mobile phone number
          rules={[{ pattern: /^1[3-9]\d{9}$/, message: t('addModal.phoneInvalid') }]}
        >
          <Input />
        </Form.Item>
      </FormModal.Section>
    </FormModal>
  );
};