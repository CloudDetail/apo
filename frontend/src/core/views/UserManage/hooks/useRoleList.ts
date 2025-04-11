import { useState } from "react";
import { getAllRolesApi } from "../api/role";
import { useApiParams } from "./useApiParams";
import { Role } from "./useRoles";

// Todo: 需要补充和完善
export const useRoleList = () => {
  const [roleList, setRoleList] = useState<Role[]>([]);
  const [loading, setLoading] = useState(true);
  const { sendRequest: fetchRolesRequest } = useApiParams(getAllRolesApi);

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

  return { roleList, loading, fetchRoles };
};