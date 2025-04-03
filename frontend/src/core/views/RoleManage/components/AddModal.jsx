/**
 * Copyright 2024 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import { Modal, Flex, Form, Input, Tooltip, Button, Card, Tree } from 'antd'
import { showToast } from 'core/utils/toast'
import { createUserApi } from 'core/api/user'
import { useEffect, useState } from 'react'
import LoadingSpinner from 'src/core/components/Spinner'
import { useTranslation } from 'react-i18next'
import MenuManagePage from '../../MenuManage'
import { getAllPermissionApi, getSubjectPermissionApi, configMenuApi } from 'src/core/api/permission'
import { useUserContext } from 'src/core/contexts/UserContext'
import i18n from 'src/i18n'
import { BsCheckAll } from 'react-icons/bs'
import { createRoleApi } from 'src/core/api/role'

const AddModal = ({ modalAddVisibility, setModalAddVisibility, getRoleList }) => {
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

  const onExpand = (expandedKeysValue) => {
    setExpandedKeys(expandedKeysValue)
    setAutoExpandParent(false)
  }
  const onCheck = (checkedKeysValue) => {
    setCheckedKeys(checkedKeysValue)
    form.setFieldsValue({ permissions: checkedKeysValue });
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
      const allPermissions = await getAllPermissionApi(params)
      // const [allPermissions, subjectPermissions] = await Promise.all([
      //   getAllPermissionApi(params),
      //   getSubjectPermissionApi({
      //     subjectId: user.userId,
      //     subjectType: 'user',
      //   }),
      // ])

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
  }, [user.userId, i18n.language])

  //创建 role
  async function createRole() {
    if (loading) return //防止重复提交
    form
      .validateFields()
      .then(
        async ({
          roleName,
          description = '',
          permissions,
        }) => {
          try {
            //设置加载状态
            setLoading(true)
            //创建用户
            const params = { roleName, description, permissionList: permissions }
            console.log("params: ", params)
            await createRoleApi(params)
            // 操作成功的反馈和状态清理
            setModalAddVisibility(false)
            await getRoleList() // 刷新
            showToast({ title: t('addModal.addSuccess'), color: 'success' })
          } catch (error) {
            console.error(error)
          } finally {
            setLoading(false)
            form.resetFields()
          }
        },
      )
  }

  return (
    <>
      <Modal
        open={modalAddVisibility}
        onCancel={() => {
          if (!loading) {
            setModalAddVisibility(false)
          }
        }}
        maskClosable={false}
        title={t('addModal.title')}
        okText={<span>{t('addModal.add')}</span>}
        cancelText={<span>{t('addModal.cancel')}</span>}
        onOk={createRole}
        width={1000}
      >
        <Flex vertical className="w-full mt-4 mb-4">
          <Flex vertical className="w-full justify-center start">
            <Form form={form} layout="vertical">
              <Form.Item
                label={t('addModal.roleName')}
                name="roleName"
                rules={[{ required: true, message: t('addModal.roleNameRequired') }]}
              >
                <div className="flex justify-start items-start">
                  <Input placeholder={t('addModal.roleNamePlaceholder')} />
                </div>
              </Form.Item>
              <Form.Item
                label={t('addModal.description')}
                name="description"
              >
                <div className="flex justify-start items-start">
                  <Input placeholder={t('addModal.descriptionPlaceholder')} />
                </div>
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
                      defaultExpandAll={false}
                      treeData={permissionTreeData}
                      fieldNames={{ title: 'featureName', key: 'featureId' }}
                    />
                  </Card>
                </div>
              </Form.Item>
            </Form>
          </Flex>
        </Flex>
        {/*  */}
      </Modal>
    </>
  )
}

export default AddModal
