/**
 * Copyright 2025 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import React, { useEffect, useState } from 'react';
import { Button, Flex } from 'antd';
import { useTranslation } from 'react-i18next';
import LoadingSpinner from 'src/core/components/Spinner';
import { TeamOutlined } from '@ant-design/icons';
import { Role } from 'src/core/types/role';
import { useRoleManage } from './useRoleManage';
import { RoleTable } from './components/RoleTable';
import { AddRoleModal } from './components/AddRoleModal';
import { EditRoleModal } from './components/EditRoleModal';
import { PermissionModal } from './components/PermissionModal';
import { BasicCard } from 'src/core/components/Card/BasicCard';

export default function RoleManage() {
  const { t } = useTranslation('core/roleManage');
  const {
    roleList,
    loading,
    addLoading,
    updateLoading,
    fetchRoles,
    addRole,
    updateRole,
    removeRole,
  } = useRoleManage();

  const [selectedRole, setSelectedRole] = useState<Role | null>(null);
  const [addModalVisible, setAddModalVisible] = useState(false);
  const [editModalVisible, setEditModalVisible] = useState(false);
  const [permissionModalVisible, setPermissionModalVisible] = useState(false);

  useEffect(() => {
    fetchRoles();
  }, []);

  const handleAddRole = async (values: { roleName: string; description: string; permissionList: any[] }) => {
    await addRole(values);
    setAddModalVisible(false);
  };

  const handleEditRole = async (values: { roleName: string; description: string }) => {
    if (!selectedRole) return;
    await updateRole(selectedRole.roleId, values);
    setEditModalVisible(false);
  };

  const handleSavePermissions = async (checkedKeys: React.Key[]) => {
    if (!selectedRole) return;
    await updateRole(selectedRole.roleId, {
      roleName: selectedRole.roleName,
      description: selectedRole.description,
      permissionList: checkedKeys
    });
    setPermissionModalVisible(false);
  };

  const showEditModal = (role: Role) => {
    setSelectedRole(role);
    setEditModalVisible(true);
  };

  const showPermissionModal = (role: Role) => {
    setSelectedRole(role);
    setPermissionModalVisible(true);
  };

  return (
    <BasicCard>
      <LoadingSpinner loading={loading} />

      <BasicCard.Header>
        <div className="w-full flex justify-end mt-2">
          <Button
            type="primary"
            onClick={() => setAddModalVisible(true)}
            icon={<TeamOutlined />}
          >
            {t('index.addRole')}
          </Button>
        </div>
      </BasicCard.Header>

      <BasicCard.Table>
        <RoleTable
          roleList={roleList}
          onEdit={showEditModal}
          onDelete={removeRole}
          onConfigPermission={showPermissionModal}
        />
      </BasicCard.Table>

      <AddRoleModal
        visible={addModalVisible}
        loading={addLoading}
        onCancel={() => setAddModalVisible(false)}
        onFinish={handleAddRole}
      />
      <EditRoleModal
        visible={editModalVisible}
        loading={updateLoading}
        selectedRole={selectedRole}
        onCancel={() => setEditModalVisible(false)}
        onFinish={handleEditRole}
      />
      <PermissionModal
        visible={permissionModalVisible}
        selectedRole={selectedRole}
        onCancel={() => setPermissionModalVisible(false)}
        onSave={handleSavePermissions}
      />
    </BasicCard>
  );
}
