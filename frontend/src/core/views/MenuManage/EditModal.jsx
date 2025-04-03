/**
 * Copyright 2024 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import { Button, Card, Tree } from 'antd'
import Checkbox from 'antd/es/checkbox/Checkbox'
import { useEffect, useState } from 'react'
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
import { updateRoleApi } from 'src/core/api/role'

function EditModal({ selectedRole }) {
  const { user, getUserPermission } = useUserContext()
  const [expandedKeys, setExpandedKeys] = useState([])
  const [checkedKeys, setCheckedKeys] = useState([])
  const [selectedKeys, setSelectedKeys] = useState([])
  const [autoExpandParent, setAutoExpandParent] = useState(true)
  const [permissionTreeData, setPermissionTreeData] = useState([])
  const [allKeys, setAllKeys] = useState([])
  const [loading, setLoading] = useState(true)
  const { t, i18n } = useTranslation('core/menuManage')
  const onExpand = (expandedKeysValue) => {
    setExpandedKeys(expandedKeysValue)
    setAutoExpandParent(false)
  }
  const onCheck = (checkedKeysValue) => {
    setCheckedKeys(checkedKeysValue)
  }
  const onSelect = (selectedKeysValue, info) => {
    setSelectedKeys(selectedKeysValue)
  }

  // Recursively traverses the tree structure to collect all keys and expanded keys
  const loopTree = (treeData = [], key = 'featureId') => {
    const allKeys = []
    const expandedKeys = []

    treeData.forEach((item) => {
      allKeys.push(item[key])

      // 如果有子节点，记录到 expandedKeys
      if (item?.children?.length > 0) {
        expandedKeys.push(item[key])

        const { allKeys: allResult, expandedKeys: expandedResult } = loopTree(item.children, key)
        expandedKeys.push(...expandedResult)
        allKeys.push(...allResult)
      }
    })
    return { allKeys, expandedKeys }
  }
  const fetchData = async () => {
    setLoading(true)
    try {
      console.log('selectedRole: ', selectedRole)
      const params = { language: i18n.language }
      const [allPermissions, subjectPermissions] = await Promise.all([
        getAllPermissionApi(params),
        getSubjectPermissionApi({
          subjectId: selectedRole?.roleId,
          subjectType: 'role',
        }),
      ])

      setPermissionTreeData(allPermissions || [])
      // 展开所有节点
      const { allKeys, expandedKeys } = loopTree(allPermissions || [])

      setExpandedKeys(expandedKeys)
      setAllKeys(allKeys)
      setCheckedKeys((subjectPermissions || []).map((permission) => permission.featureId))
      // 在这里处理两者的数据
    } catch (error) {
      console.error(t('index.errorFetchingPermissions'), error)
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    if (selectedRole?.roleId) fetchData()
  }, [user.userId, i18n.language, selectedRole?.roleId])

  //保存配置
  function updateRole() {
    setLoading(true)
    const params = new URLSearchParams()
    params.append('roleId', selectedRole?.roleId);
    params.append('roleName', selectedRole?.roleName);
    checkedKeys.forEach((value) => params.append('permissionList', value))
    updateRoleApi(params)
      .then((res) => {
        showToast({
          title: t('index.menuConfigSuccess'),
          color: 'success',
        })
      })
      .catch((error) => {
        console.error(error)
      })
      .finally(() => {
        fetchData()
        getUserPermission()
        setLoading(false)
      })
  }
  return (
    <>
      <Card style={{ height: 'calc(100vh - 60px)', overflow: 'auto', width: '100%' }}>
        <LoadingSpinner loading={loading} />
        <Button
          type="primary"
          className="mx-4 mb-4"
          onClick={() => setCheckedKeys(allKeys)}
          icon={<BsCheckAll />}
        >
          {t('index.selectAll')}
        </Button>
        <Tree
          checkable
          onExpand={onExpand}
          expandedKeys={expandedKeys}
          autoExpandParent={autoExpandParent}
          onCheck={onCheck}
          checkedKeys={checkedKeys}
          onSelect={onSelect}
          selectedKeys={selectedKeys}
          defaultExpandAll={false}
          treeData={permissionTreeData}
          fieldNames={{ title: 'featureName', key: 'featureId' }}
        />
        <Button type="primary" className="m-4" onClick={updateRole}>
          {t('index.save')}
        </Button>
      </Card>
    </>
  )
}
export default EditModal
