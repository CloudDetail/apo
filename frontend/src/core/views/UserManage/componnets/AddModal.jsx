/**
 * Copyright 2024 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import { Modal, Flex, Form, Input, Tooltip, Select } from 'antd'
import { showToast } from 'core/utils/toast'
import { createUserApi, getRoleListApi } from 'core/api/user'
import { useEffect, useState } from 'react'
import LoadingSpinner from 'src/core/components/Spinner'

const AddModal = ({ modalAddVisibility, setModalAddVisibility, getUserList }) => {
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

  useEffect(() => {
    getRoleList()
  }, [])

  //创建用户
  async function createUser() {
    if (loading) return
    form.validateFields().then(async (props) => {
      const {
        username,
        password,
        confirmPassword,
        email = '',
        phone = '',
        corporation = '',
        roleList = [],
      } = props
      try {
        //设置加载状态
        setLoading(true)
        //创建用户
        const params = new URLSearchParams()
        params.append('username', username)
        params.append('password', password)
        params.append('confirmPassword', confirmPassword)
        params.append('email', email)
        params.append('phone', phone)
        params.append('corporation', corporation)
        roleList.forEach((role) => {
          params.append('roleList', role)
        })

        await createUserApi(params)
        // 操作成功的反馈和状态清理
        setModalAddVisibility(false)
        await getUserList()
        showToast({ title: '用户添加成功', color: 'success' })
      } catch (error) {
        console.error(error)
      } finally {
        setLoading(false)
        form.resetFields()
      }
    })
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
        title="新增用户"
        okText={<span>新增</span>}
        onOk={createUser}
        width={1000}
      >
        <LoadingSpinner loading={loading} />
        <Flex vertical className="w-full mt-4 mb-4">
          <Flex vertical className="w-full justify-center start">
            <Form form={form} layout="vertical">
              <Form.Item
                label="用户名"
                name="username"
                rules={[{ required: true, message: '请输入用户名' }]}
              >
                <div className="flex justify-start items-start">
                  <Input placeholder="请输入用户名" className="h-8" />
                </div>
              </Form.Item>
              <Form.Item
                label="密码"
                name="password"
                rules={[
                  { required: true, message: '请输入密码' },
                  {
                    pattern: /^(?=.*[a-z])(?=.*[A-Z])(?=.*[!@#$%^&*()\-_+=<>?/{}[\]|:;.,~]).{9,}$/,
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
                <div className="flex justify-start items-start">
                  <Input.Password placeholder="请输入密码" />
                </div>
              </Form.Item>
              <Form.Item
                label="重复密码"
                name="confirmPassword"
                dependencies={['password']}
                rules={[
                  { required: true, message: '请再次输入密码' },
                  ({ getFieldValue }) => ({
                    validator(_, value) {
                      if (!value || getFieldValue('password') === value) {
                        return Promise.resolve()
                      }
                      return Promise.reject(new Error('两次输入的密码不一致'))
                    },
                  }),
                ]}
              >
                <Input.Password placeholder="请再次输入密码" />
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
            </Form>
          </Flex>
        </Flex>
      </Modal>
    </>
  )
}

export default AddModal
