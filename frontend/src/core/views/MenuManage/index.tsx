/**
 * Copyright 2024 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import React, { useEffect, useMemo, useState } from 'react';
import { Menu, Layout, Card, Typography } from 'antd';
import { TeamOutlined, SafetyCertificateOutlined } from '@ant-design/icons';
import { updateRoleApi } from 'src/core/api/role';
import LoadingSpinner from 'src/core/components/Spinner';
import { useUserContext } from 'src/core/contexts/UserContext';
import { showToast } from 'src/core/utils/toast';
import { useTranslation } from 'react-i18next';
import { useMenuPermission } from './useMenuPermission';
import { useApiParams } from 'src/core/hooks/useApiParams';
import PermissionTree from 'src/core/components/PermissionTree';
import { Role } from 'src/core/types/role';

function MenuManagePage() {
  const {
    roleList,
    selectedRole,
    loading,
    updateLoading,
    fetchRoles,
    handleRoleSelect,
    handleSavePermissions
  } = useMenuPermission();
  const { t } = useTranslation('core/menuManage');

  const menuItems = useMemo(() => {
    return roleList.map((role) => ({
      key: role.roleId,
      label: role.roleName,
      icon: <TeamOutlined className='ml-6 mr-1' />,
    }));
  }, [roleList]);

  const onSelect = ({ key }: { key: string }) => {
    handleRoleSelect(key);
  };

  useEffect(() => {
    fetchRoles();
  }, []);

  return (
    <Layout style={{ height: 'calc(100vh - 100px)', overflow: 'hidden' }}>
      <LoadingSpinner loading={loading || updateLoading} />
      <Layout.Content className="p-0 flex gap-0 h-full">
          <Card
            className="w-48"
            style={{ width: '20%', borderTopRightRadius: '0px', borderBottomRightRadius: '0px' }}
            styles={{
              body: {
                height: '100%', padding: '0px', paddingBlockStart: '2px'
              },
              header: {
                backgroundColor: '#1d1d1d',
              }
            }}
            title={
              <Typography.Title level={5} className="mb-0 flex items-center">
                <TeamOutlined className="mr-2 text-blue-500" />
                {t('index.roleList')}
              </Typography.Title>
            }
            // bordered={false}
          >
            <Menu
              selectedKeys={[selectedRole?.roleId?.toString()]}
              mode="vertical"
              items={menuItems}
              // className="border-none"
              onSelect={onSelect}
              style={{ height: 'calc(100vh - 200px)', border: 'none', overflowY: 'auto' }}
            />
          </Card>

          <Card
            className="flex-1 shadow-md"
            style={{ borderTopLeftRadius: '0px', borderBottomLeftRadius: '0px' }}
            styles={{
              body: {
                padding: '0px', paddingBlockStart: '2px', paddingInlineEnd: '12px'
              },
              header: {
                backgroundColor: '#1d1d1d',
              }
            }}
            title={
              <Typography.Title level={5} className="mb-0 flex items-center">
                <SafetyCertificateOutlined className="mr-2 text-blue-500" />
                {selectedRole ? `${t('index.menuPermissionSetting')} - ${selectedRole.roleName}` : t('index.selectRole')}
              </Typography.Title>
            }
            // bordered={false}
          >
            {selectedRole ? (
              <PermissionTree
                subjectId={selectedRole.roleId}
                subjectType="role"
                onSave={handleSavePermissions}
                className="permission-tree"
                style={{ height: 'calc(100vh - 240px)', border: 'none' }}
              />
            ) : (
              <div className="text-center text-gray-500 py-8">
                {t('index.pleaseSelectRole')}
              </div>
            )}
          </Card>
      </Layout.Content>
    </Layout>
  );
}

export default MenuManagePage;