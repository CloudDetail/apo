/**
 * Copyright 2024 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */
import React from 'react'
import { Modal, Form, Input, Divider, Button, Tooltip, Select } from 'antd'
import { useEffect, useState } from 'react'
import { updateUserInfoApi, updatePasswordWithNoOldPwdApi, getRoleListApi } from 'core/api/user'
import { showToast } from 'core/utils/toast'
import LoadingSpinner from 'src/core/components/Spinner'

const EditModal = React.memo(
  ({ selectedUser, modalEditVisibility, setModalEditVisibility, getUserList }) => {
    const [loading, setLoading] = useState(false)
    const [roleOptions, setRoleOptions] = useState(null)
    const [form] = Form.useForm()

    //获取角色列表
    const getRoleList = () => {
      setLoading(true)
      getRoleListApi()
        .then((res) => {
          const list = res?.map((item) => {
            return {
              value: item.roleId,
              label: item.roleName,
            }
          })
          setRoleOptions(list)
        })
        .catch((error) => {
          console.log(error)
        })
        .finally(() => {
          setLoading(false)
        })
    }

    //初始化编辑表单
    const initForm = () => {
      if (modalEditVisibility) {
        form.resetFields()
        getRoleList()
        form.setFieldsValue({
          username: selectedUser?.username,
          email: selectedUser?.email,
          phone: selectedUser?.phone,
          corporation: selectedUser?.corporation,
          roleList: selectedUser?.roleList?.map((role) => role.roleId),
        })
      }
    }

    useEffect(() => {
      initForm()
    }, [modalEditVisibility])

    const editUser = () => {
      if (loading) return
      form
        .validateFields(['email', 'phone', 'corporation', 'roleList'])
        .then(async ({ email = '', phone = '', corporation = '', roleList = [] }) => {
          setLoading(true)

          const params = new URLSearchParams()
          params.append('userId', selectedUser?.userId)
          params.append('email', email)
          params.append('phone', phone)
          params.append('corporation', corporation)
          roleList.forEach((role) => {
            params.append('roleList', role)
          })

          await updateUserInfoApi(params)

          setModalEditVisibility(false)
          getUserList()
          showToast({ title: '保存成功', color: 'success' })
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
              title: '密码修改成功',
              color: 'success',
            })
            setModalEditVisibility(false)
          } catch (error) {
            console.error(error)
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
          title="编辑用户"
          width={1000}
          footer={null}
          centered
        >
          <LoadingSpinner loading={loading} />
          <div className="flex flex-col overflow-auto w-full mt-4 mb-4 justify-center">
            <Form form={form} layout="vertical">
              <Form.Item label="用户名" name="username">
                <Input placeholder="请输入用户名" className="h-8" disabled />
              </Form.Item>
              <Form.Item label="角色" name="roleList">
                <Select
                  mode="multiple"
                  placeholder="请选择角色"
                  options={roleOptions}
                  className="h-8"
                  maxTagCount={7}
                  maxTagPlaceholder={(omittedValues) => `+${omittedValues.length} 更多`}
                  optionLabelProp="label"
                  allowClear
                />
              </Form.Item>
              <Form.Item
                label="邮件"
                name="email"
                rules={[
                  {
                    type: 'email',
                    message: '请输入有效的邮箱地址',
                  },
                ]}
              >
                <Input placeholder="请输入用户邮箱" />
              </Form.Item>
              <Form.Item
                label="电话号码"
                name="phone"
                rules={[
                  {
                    pattern: /^1[3-9]\d{9}$/, // 中国大陆手机号正则
                    message: '请输入有效的电话号码',
                  },
                ]}
              >
                <Input placeholder="请输入电话号码" />
              </Form.Item>
              <Form.Item label="组织" name="corporation">
                <Input placeholder="请输入组织" />
              </Form.Item>
              <Button type="primary" onClick={editUser}>
                修改用户信息
              </Button>
              <Divider />
              <div className="mt-3">
                <Form.Item
                  label="新密码"
                  name="newPassword"
                  rules={[
                    {
                      required: true,
                      message: '请输入密码',
                    },
                    {
                      pattern:
                        /^(?=.*[a-z])(?=.*[A-Z])(?=.*[!@#$%^&*()\-_+=<>?/{}[\]|:;.,~]).{9,}$/,
                      message: (
                        <p>
                          密码必须包含大写字母、小写字母、
                          <Tooltip title="(! @ # $ % ^ & * ( ) - _ + = < > ? / { } [ ] | : ; . , ~)">
                            <span className="underline">特殊字符</span>
                          </Tooltip>
                          ，且长度大于8
                        </p>
                      ),
                    },
                  ]}
                >
                  <Input.Password placeholder="请输入密码" />
                </Form.Item>
                <Form.Item
                  label="重复新密码"
                  name="confirmPassword"
                  rules={[
                    {
                      required: true,
                      message: '请确认密码',
                    },
                    ({ getFieldValue }) => ({
                      validator(_, value) {
                        if (!value || getFieldValue('newPassword') === value) {
                          return Promise.resolve()
                        }
                        return Promise.reject(new Error('两次输入的密码不一致'))
                      },
                    }),
                  ]}
                >
                  <Input.Password placeholder="请重复密码" />
                </Form.Item>
              </div>
              <Button type="primary" onClick={resetPassword}>
                修改密码
              </Button>
            </Form>
          </div>
        </Modal>
      </>
    )
  },
)

export default EditModal
