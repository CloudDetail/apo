/**
 * Copyright 2024 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import React, { useMemo } from 'react';
import { Menu, Layout, Card, Typography } from 'antd';
import { TeamOutlined, SafetyCertificateOutlined } from '@ant-design/icons';
import { updateRoleApi } from 'src/core/api/role';
import LoadingSpinner from 'src/core/components/Spinner';
import { useUserContext } from 'src/core/contexts/UserContext';
import { showToast } from 'src/core/utils/toast';
import { useTranslation } from 'react-i18next';
import { useRoles } from 'src/core/hooks/useRoles';
import { useApiParams } from 'src/core/hooks/useApiParams';
import PermissionTree from 'src/core/components/PermissionTree';
import classNames from 'classnames';

function MenuManagePage() {
  const { roleList, selectedRole, loading, selectRole } = useRoles();
  const { t } = useTranslation('core/menuManage');
  const { getUserPermission } = useUserContext();

  // 使用 useApiParams 钩子处理角色权限更新
  const { sendRequest: updateRole, loading: updateLoading } = useApiParams(updateRoleApi);

  // 将角色列表转换为菜单项
  // 优化菜单项，添加图标
  const menuItems = useMemo(() => {
    return roleList.map((role) => ({
      key: role.roleId,
      label: role.roleName,
      icon: <TeamOutlined className='ml-6 mr-1' />,
    }));
  }, [roleList]);

  // 处理菜单选择
  const onSelect = ({ key }: { key: string }) => {
    selectRole(key);
  };

  // 处理权限保存
  const handleSavePermissions = async (checkedKeys: React.Key[]) => {
    if (!selectedRole) return;

    await updateRole(
      {
        roleId: selectedRole.roleId,
        roleName: selectedRole.roleName,
        permissionList: checkedKeys
      },
      {
        onSuccess: () => {
          showToast({
            title: t('index.menuConfigSuccess'),
            color: 'success',
          });

          // 更新用户权限
          getUserPermission();
        },
        onError: (error) => {
          console.error('保存权限失败:', error);
        }
      }
    );
  };

  return (
    <Layout style={{ height: 'calc(100vh - 100px)', overflow: 'hidden' }}>
      <LoadingSpinner loading={loading || updateLoading} />
      <Layout.Content className="p-0 flex gap-0 h-full">
          <Card
            className="w-48"
            style={{ width: '20%', borderTopRightRadius: '0px', borderBottomRightRadius: '0px' }}
            styles={{ body: { height: '100%', padding: '0px', paddingBlockStart: '2px' }}}
            // title={t('index.roleList')}
            title={
              <Typography.Title level={5} className="mb-0 flex items-center">
                <TeamOutlined className="mr-2 text-blue-500" />
                角色列表
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
            styles={{ body: { padding: '0px', paddingBlockStart: '2px' }}}
            // title={selectedRole ? `${t('index.permissions')}: ${selectedRole.roleName}` : t('index.selectRole')}
            title={
              <Typography.Title level={5} className="mb-0 flex items-center">
                <SafetyCertificateOutlined className="mr-2 text-blue-500" />
                {selectedRole ? `菜单权限配置 - ${selectedRole.roleName}` : '请选择角色'}
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