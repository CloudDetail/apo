import { useState } from "react";
import { getAllRolesApi, revokeUserRoleApi } from "src/core/api/role";
import { useApiParams } from "src/core/hooks/useApiParams";
import { Role } from "src/core/types/role";
import { showToast } from "src/core/utils/toast";
import { useTranslation } from "react-i18next";

export interface RoleOption {
  label: string;
  key: string | number;
  value: string | number;
}

export const useRoleActions = () => {
  const { t } = useTranslation('core/userManage');
  const [roleList, setRoleList] = useState<Role[]>([]);
  const [loading, setLoading] = useState(true);

  const { sendRequest: fetchRolesRequest } = useApiParams(getAllRolesApi);
  const { sendRequest: revokeUserRole } = useApiParams(revokeUserRoleApi);

  const fetchRoles = async () => {
    setLoading(true);
    try {
      const roles = await fetchRolesRequest({}, { useURLSearchParams: false });
      setRoleList(roles || []);
    } catch (error) {
      console.error('获取角色列表失败:', error);
    } finally {
      setLoading(false);
    }
  };

  const handleRevokeUserRole = async (userId: string | number, roleId: string | number) => {
    await revokeUserRole(
      {
        userId,
        roleList: [roleId]
      }
    );
  };

  // Get role dropdown menu options
  const getRoleOptions = (): RoleOption[] => {
    return roleList.map((role) => ({
      label: role.roleName,
      key: role.roleId,
      value: role.roleId
    }));
  };

  return {
    loading,
    fetchRoles,
    handleRevokeUserRole,
    roleOptions: getRoleOptions()
  };
};