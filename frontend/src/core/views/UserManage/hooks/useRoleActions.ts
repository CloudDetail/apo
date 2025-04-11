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

  // 获取角色列表
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

  // 更改用户角色
  const handleRevokeUserRole = async (userId: string | number, roleId: string | number) => {
    await revokeUserRole(
      {
        userId,
        roleList: [roleId]
      },
      {
        onSuccess: () => {
          showToast({
            title: t('index.revokeSuccess'),
            color: 'success',
          });
        },
        onError: (error) => {
          console.error('撤销用户角色失败:', error);
        }
      }
    );
  };

  // 转换角色列表为下拉菜单选项
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