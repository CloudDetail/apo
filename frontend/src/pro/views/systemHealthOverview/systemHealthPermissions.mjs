export const canManageSystemHealthConfig = (user) => {
  if (!user) return false
  if (user.role === 'admin' || user.roleList === 'admin') return true
  return Array.isArray(user.roleList) && user.roleList.some((role) => role?.roleName === 'admin')
}
