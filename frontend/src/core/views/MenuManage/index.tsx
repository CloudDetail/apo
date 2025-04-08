/**
 * Copyright 2024 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import React, { useMemo } from 'react';
import { Menu } from 'antd';
import { updateRoleApi } from 'src/core/api/role';
import LoadingSpinner from 'src/core/components/Spinner';
import { useUserContext } from 'src/core/contexts/UserContext';
import { showToast } from 'src/core/utils/toast';
import { useTranslation } from 'react-i18next';
import { useRoles } from 'src/core/hooks/useRoles';
import { useApiParams } from 'src/core/hooks/useApiParams';
import PermissionTree from 'src/core/components/PermissionTree';

function MenuManagePage() {
  const { roleList, selectedRole, loading, selectRole } = useRoles();
  const { t } = useTranslation('core/menuManage');
  const { getUserPermission } = useUserContext();

  // 使用 useApiParams 钩子处理角色权限更新
  const { sendRequest: updateRole, loading: updateLoading } = useApiParams(updateRoleApi);

  // 将角色列表转换为菜单项
  const menuItems = useMemo(() => {
    return roleList.map((role) => ({
      key: role.roleId,
      label: role.roleName
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
    <>
      <LoadingSpinner loading={loading || updateLoading} />
      <div className='flex'>
        <Menu
          selectedKeys={[selectedRole?.roleId?.toString()]}
          mode="vertical"
          items={menuItems}
          className='w-36'
          onSelect={onSelect}
        />

        {selectedRole && (
          <PermissionTree
            subjectId={selectedRole.roleId}
            subjectType="role"
            onSave={handleSavePermissions}
            readOnly={selectedRole.roleName === 'admin'}
            className="w-full"
          />
        )}
      </div>
    </>
  );
}

export default MenuManagePage;