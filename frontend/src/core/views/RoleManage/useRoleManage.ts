/**
 * Copyright 2025 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import { useState } from 'react';
import { useTranslation } from 'react-i18next';
import { useApiParams } from 'src/core/hooks/useApiParams';
import { getAllRolesApi, createRoleApi, updateRoleApi, deleteRoleApi } from 'src/core/api/role';
import { showToast } from 'src/core/utils/toast';
import { Role } from 'src/core/types/role';

export const useRoleManage = () => {
  const { t } = useTranslation('core/roleManage');
  const [roleList, setRoleList] = useState<Role[]>([]);
  const [loading, setLoading] = useState(true);

  const { sendRequest: fetchRolesRequest } = useApiParams(getAllRolesApi);
  const { sendRequest: addRoleRequest, loading: addLoading } = useApiParams(createRoleApi);
  const { sendRequest: updateRoleRequest, loading: updateLoading } = useApiParams(updateRoleApi);
  const { sendRequest: removeRoleRequest } = useApiParams(deleteRoleApi);

  const fetchRoles = async () => {
    setLoading(true);
    try {
      const roles = await fetchRolesRequest({}, { useURLSearchParams: false });
      setRoleList(roles || []);
    } catch (error) {
      console.error('Error fetch role list:', error);
    } finally {
      setLoading(false);
    }
  };

  const addRole = async (values: { roleName: string; description: string; permissionList: any[] }) => {
    await addRoleRequest(values, {
      onSuccess: () => {
        showToast({ title: t('addModal.addSuccess'), color: 'success' });
        fetchRoles();
        return true;
      },
      onError: (error) => {
        console.error('Error add role:', error);
        return false;
      }
    });
  };

  const updateRole = async (roleId: string | number, values: { roleName: string; description: string; permissionList?: any[] }) => {
    await updateRoleRequest(
      { ...values, roleId },
      {
        onSuccess: () => {
          showToast({ title: t('editModal.updateSuccess'), color: 'success' });
          fetchRoles();
          return true;
        },
        onError: (error) => {
          console.error('Error update role:', error);
          return false;
        }
      }
    );
  };

  const removeRole = async (roleId: string | number) => {
    await removeRoleRequest(
      { roleId },
      {
        useURLSearchParams: false,
        onSuccess: () => {
          showToast({ title: t('editModal.deleteSuccess'), color: 'success' });
          fetchRoles();
          return true;
        },
        onError: (error) => {
          console.error('Error delete role:', error);
          return false;
        }
      }
    );
  };

  return {
    roleList,
    loading,
    addLoading,
    updateLoading,
    fetchRoles,
    addRole,
    updateRole,
    removeRole,
  };
};