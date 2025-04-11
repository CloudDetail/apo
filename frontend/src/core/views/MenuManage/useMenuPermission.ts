import { useState } from "react";
import { getAllRolesApi, updateRoleApi } from "src/core/api/role";
import { useApiParams } from "src/core/hooks/useApiParams";
import { Role } from "src/core/types/role";
import { useUserContext } from "src/core/contexts/UserContext";
import { showToast } from "src/core/utils/toast";
import { useTranslation } from 'react-i18next';

export const useMenuPermission = () => {
  const [roleList, setRoleList] = useState<Role[]>([]);
  const [selectedRole, setSelectedRole] = useState<Role | null>(null);
  const [loading, setLoading] = useState(true);
  const { getUserPermission } = useUserContext();
  const { sendRequest: fetchRolesRequest } = useApiParams(getAllRolesApi);
  const { sendRequest: updateRole, loading: updateLoading } = useApiParams(updateRoleApi);
  const { t } = useTranslation('core/menuManage');

  const fetchRoles = async () => {
    setLoading(true);
    try {
      const roles = await fetchRolesRequest({}, { useURLSearchParams: false });
      setRoleList(roles || []);

      // Initial set
      if (roles?.length > 0) {
        setSelectedRole(roles[0])
      }
    } catch (error) {
      console.error('Error fetch role list:', error);
    } finally {
      setLoading(false);
    }
  };

  const handleRoleSelect = (roleId: string) => {
    const role = roleList.find(role => role.roleId == roleId);
    setSelectedRole(role || null);
  };

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
            title: t('index.menuSetSuccess'),
            color: 'success',
          });
          getUserPermission();
        },
        onError: (error) => {
          console.error('Error save permission:', error);
        }
      }
    );
  };

  return {
    roleList,
    selectedRole,
    loading,
    updateLoading,
    fetchRoles,
    handleRoleSelect,
    handleSavePermissions
  };
};