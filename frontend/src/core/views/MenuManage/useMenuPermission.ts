import { useState } from "react";
import { getAllRolesApi, updateRoleApi } from "src/core/api/role";
import { useApiParams } from "src/core/hooks/useApiParams";
import { Role } from "src/core/types/role";
import { useUserContext } from "src/core/contexts/UserContext";
import { showToast } from "src/core/utils/toast";

export const useMenuPermission = () => {
  const [roleList, setRoleList] = useState<Role[]>([]);
  const [selectedRole, setSelectedRole] = useState<Role | null>(null);
  const [loading, setLoading] = useState(true);
  const { getUserPermission } = useUserContext();
  const { sendRequest: fetchRolesRequest } = useApiParams(getAllRolesApi);
  const { sendRequest: updateRole, loading: updateLoading } = useApiParams(updateRoleApi);

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
      console.error('获取角色列表失败:', error);
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
            title: '菜单配置成功',
            color: 'success',
          });
          getUserPermission();
        },
        onError: (error) => {
          console.error('保存权限失败:', error);
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