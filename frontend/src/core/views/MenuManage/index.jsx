/**
 * Copyright 2024 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import { Button, Card, Menu, Tree } from 'antd'
import Checkbox from 'antd/es/checkbox/Checkbox'
import { useEffect, useMemo, useState } from 'react'
import { BsCheckAll } from 'react-icons/bs'
import {
  configMenuApi,
  getAllPermissionApi,
  getSubjectPermissionApi,
} from 'src/core/api/permission'
import LoadingSpinner from 'src/core/components/Spinner'
import { useUserContext } from 'src/core/contexts/UserContext'
import { showToast } from 'src/core/utils/toast'
import { useTranslation } from 'react-i18next'
import EditModal from './EditModal'
import { getAllRolesApi } from 'src/core/api/role'

function MenuManagePage() {
  const [loading, setLoading] = useState(true)
  const [roleList, setRoleList] = useState([])
  const [selectedRole, setSelectedRole] = useState()
  const [selectedKey, setSelectedKey] = useState(0)
  const { t, i18n } = useTranslation('core/menuManage')

  async function fetchRoles() {
    try {
      const roles = await getAllRolesApi(); // 等待 API 返回数据
      setRoleList(roles)
      if (roles?.length > 0) {
        setSelectedRole(roles[0]) // 默认选择第一项
        setSelectedKey(roles[0].roleId) // 设置默认选中项
      }
    } catch (error) {
      console.error("Failed to fetch roles: ", error); // 捕获并处理错误
    }
  }

  useEffect(() => {
    fetchRoles(); // 调用异步函数
  }, []); // 空依赖数组，确保只在组件挂载时调用一次

  const menuItems = useMemo(() => {
    return roleList.map((role) => ({
      key: role.roleId,
      label: role.roleName
    }))
  }, [roleList])

  const onSelect = ({ key }) => {
    setSelectedKey(key)
    setSelectedRole(roleList.find((role) => role.roleId == key))
  }

  return (
    <>
      <div className='flex'>
        <Menu
          selectedKeys={[selectedKey.toString()]}
          mode="vertical"
          // theme="dark"
          // inlineCollapsed={collapsed}
          items={menuItems}
          className='w-36'
          onSelect={onSelect}
        />
        <EditModal selectedRole={selectedRole} />
      </div>
    </>
  )
}
export default MenuManagePage
