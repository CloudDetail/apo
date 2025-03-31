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

const EditModal = React.memo(
  ({ selectedUser, modalEditVisibility, setModalEditVisibility, getUserList }) => {
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
          subjectId: user.userId,
          subjectType: 'user',
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
  }, [user.userId, i18n.language])

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
        form.setFieldsValue({
          username: selectedUser?.username,
          email: selectedUser?.email,
          phone: selectedUser?.phone,
          corporation: selectedUser?.corporation,
        })
      }
    }, [modalEditVisibility])

    const editUser = () => {
      if (loading) return
      form
        .validateFields(['email', 'phone', 'corporation'])
        .then(async ({ email = '', phone = '', corporation = '' }) => {
          setLoading(true)

          const params = {
            email,
            phone,
            corporation,
          }

          await updateCorporationApi({ userId: selectedUser?.userId, ...params })

          setModalEditVisibility(false)
          getUserList()
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

    const resetPassword = () => {
      if (loading) return
      form
        .validateFields(['newPassword', 'confirmPassword'])
        .then(async ({ newPassword, confirmPassword }) => {
          try {
            setLoading(true)
            const params = { newPassword, confirmPassword }
            await updatePasswordWithNoOldPwdApi({ userId: selectedUser?.userId, ...params })
            showToast({
              title: t('editModal.resetPasswordSuccess'),
              color: 'success',
            })
            setModalEditVisibility(false)
          } catch (error) {
            console.error(error)
            showToast({
              title: error.response?.data?.message || t('editModal.resetPasswordFail'),
              color: 'danger',
            })
            setModalEditVisibility(false)
          } finally {
            form.resetFields()
            setLoading(false)
          }
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
                <div className="flex justify-start items-start">
                  <Input disabled={true} value={selectedUser?.roleName} />
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
                name="description"
                rules={[{ required: true, message: t('addModal.roleNameRequired') }]}
              >
                <div className="flex justify-start items-start">
                  <Card style={{ height: 'calc(100vh - 60px)', overflow: 'auto', width: "100%" }}>
                    <LoadingSpinner loading={loading} />
                    <Button
                      type="primary"
                      className="mx-4 mb-4"
                      onClick={() => setCheckedKeys(allKeys)}
                      icon={<BsCheckAll />}
                    >
                      {/* {t('index.selectAll')} */}
                      选择全部
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
          {/* <div className="flex flex-col w-full justify-start items-start">
            <MenuManagePage />
          </div> */}
              </Form.Item>
              <Button type="primary" onClick={() => {}}>
                确认修改
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
