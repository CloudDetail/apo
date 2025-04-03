import { useState, useEffect } from 'react';
import { getAllRolesApi } from 'src/core/api/role';

// 定义角色接口
export interface Role {
  roleId: string | number;
  roleName: string;
  [key: string]: any;
}

/**
 * 获取角色列表的钩子
 * 简单直观，只处理角色数据获取
 */
export function useRoles() {
  const [roleList, setRoleList] = useState<Role[]>([]);
  const [selectedRole, setSelectedRole] = useState<Role | null>(null);
  const [loading, setLoading] = useState<boolean>(true);

  const fetchRoles = async (): Promise<Role[]> => {
    setLoading(true);
    try {
      const roles = await getAllRolesApi();
      setRoleList(roles || []);

      if (roles?.length > 0) {
        setSelectedRole(roles[0]);
      }

      return roles;
    } catch (error) {
      console.error("获取角色列表失败:", error);
      return [];
    } finally {
      setLoading(false);
    }
  };

  // 选择角色的方法
  const selectRole = (roleId: string | number): Role | null => {
    const role = roleList.find(role => role.roleId == roleId);
    if (role) {
      setSelectedRole(role);
      return role;
    }
    return null;
  };

  useEffect(() => {
    fetchRoles();
  }, []);

  return {
    roleList,
    selectedRole,
    loading,
    fetchRoles,
    selectRole
  };
}