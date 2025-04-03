/**
 * Copyright 2024 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */
import React from 'react'
import { Modal, Flex, Form, Input, Divider, Button, Tooltip, Card, Tree } from 'antd'
import { useEffect, useState } from 'react'
import {
  getUserListApi,
  updateEmailApi,
  updatePhoneApi,
  updateCorporationApi,
  updatePasswordWithNoOldPwdApi,
} from 'core/api/user'
import { showToast } from 'core/utils/toast'
import LoadingSpinner from 'src/core/components/Spinner'
import { useTranslation } from 'react-i18next'
import { BsCheckAll } from 'react-icons/bs'
import { getAllPermissionApi, getSubjectPermissionApi, configMenuApi } from 'src/core/api/permission'
import { useUserContext } from 'src/core/contexts/UserContext'
import i18n from 'src/i18n'
import { updateRoleApi } from 'src/core/api/role'

const EditModal = React.memo(
  ({ selectedRole, modalEditVisibility, setModalEditVisibility, getRoleList }) => {
    const { t } = useTranslation('core/roleManage')
    const [loading, setLoading] = useState(false)
    const [form] = Form.useForm()

  const { user, getUserPermission } = useUserContext()
  const [expandedKeys, setExpandedKeys] = useState([])
  const [checkedKeys, setCheckedKeys] = useState([])
  const [selectedKeys, setSelectedKeys] = useState([])
  const [autoExpandParent, setAutoExpandParent] = useState(true)
  const [permissionTreeData, setPermissionTreeData] = useState([])
  const [allKeys, setAllKeys] = useState([])
  // const [loading, setLoading] = useState(true)
  // const { t, i18n } = useTranslation('core/menuManage')

  const onExpand = (expandedKeysValue) => {
    setExpandedKeys(expandedKeysValue)
    setAutoExpandParent(false)
  }
  const onCheck = (checkedKeysValue) => {
    setCheckedKeys(checkedKeysValue)
    form.setFieldsValue({ permissions: checkedKeysValue })
  }
  const onSelect = (selectedKeysValue, info) => {
    setSelectedKeys(selectedKeysValue)
  }

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
      const params = { language: i18n.language }
      const [allPermissions, subjectPermissions] = await Promise.all([
        getAllPermissionApi(params),
        getSubjectPermissionApi({
          subjectId: selectedRole.roleId,
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
    if (user.userId) fetchData()
  }, [user.userId, i18n.language, modalEditVisibility])

  //保存配置
  function configMenu() {
    setLoading(true)
    const params = new URLSearchParams()
    checkedKeys.forEach((value) => params.append('permissionList', value))
    configMenuApi(params)
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


    useEffect(() => {
      if (modalEditVisibility) {
        form.resetFields()
        console.log('useEffect in Edit:', selectedRole)
        console.log("permissions: ", checkedKeys)
        form.setFieldsValue({
          roleName: selectedRole.roleName,
          description: selectedRole.description,
          // Todo: 该角色的当前的权限需要展示
          permissions: checkedKeys,
        })
      }
    }, [modalEditVisibility, selectedRole, form])

    const editRole = () => {
      if (loading) return
      form
        .validateFields(['roleName', 'description', 'permissions'])
        .then(async ({ roleName = '', description = '', permissions }) => {
          setLoading(true)

          const params = {
            roleName,
            description,
            permissionList: permissions
          }

          // const params = new URLSearchParams()

          // params.append('roleName', roleName);
          // params.append('description', description);

          // checkedKeys.forEach((value) => params.append('permissionList', value))

          // console.log('editRole: ', params)

          await updateRoleApi({ roleId: selectedRole?.roleId, ...params })

          setModalEditVisibility(false)
          getRoleList()
          showToast({ title: t('editModal.saveSuccess'), color: 'success' })
          form.resetFields()
        })
        .catch((error) => {
          console.error(error)
        })
        .finally(() => {
          setLoading(false)
        })
    }

    return (
      <>
        <Modal
          open={modalEditVisibility}
          onCancel={() => {
            if (!loading) {
              setModalEditVisibility(false)
            }
          }}
          maskClosable={false}
          title={t('editModal.title')}
          width={1000}
          footer={null}
        >
          <LoadingSpinner loading={loading} />
          <Flex vertical className="w-full mt-4 mb-4">
          <Flex vertical className="w-full justify-center start">
            <Form form={form} layout="vertical">
              <Form.Item
                label={t('addModal.roleName')}
                name="roleName"
              >
                <Input />
              </Form.Item>
              <Form.Item
                label={t('addModal.description')}
                name="description"
              >
                <Input placeholder={t('addModal.descriptionPlaceholder')} />
              </Form.Item>
              <Form.Item
                label={t('addModal.permissions')}
                name="permissions"
              >
                <div className="flex justify-start items-start">
                  <Card style={{ overflow: 'auto', width: "100%" }}>
                    <LoadingSpinner loading={loading} />
                    <Button
                      type="primary"
                      className="mx-4 mb-4"
                      onClick={() => {
                        setCheckedKeys(allKeys)
                        form.setFieldsValue({ permissions: allKeys })
                      }}
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
                      defaultExpandAll={true}
                      treeData={permissionTreeData}
                      fieldNames={{ title: 'featureName', key: 'featureId' }}
                    />
                  </Card>
                </div>
              </Form.Item>
              <Button type="primary" onClick={editRole}>
                {t('editModal.save')}
              </Button>
            </Form>
          </Flex>
        </Flex>
        </Modal>
      </>
    )
  },
)

export default EditModal
