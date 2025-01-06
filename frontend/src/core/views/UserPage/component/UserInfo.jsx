/**
 * Copyright 2024 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import { Button, Popconfirm, Form, Input } from 'antd'
import { updateUserInfoApi, getUserInfoApi } from 'core/api/user'
import { showToast } from 'core/utils/toast'
import { useEffect } from 'react'
import { useUserContext } from 'src/core/contexts/UserContext'

export default function UserInfo() {
  const [form] = Form.useForm()
  const { user } = useUserContext()

  async function getUserInfo() {
    try {
      const user = await getUserInfoApi()
      if (user) {
        const { email, phone, corporation } = user
        form.setFieldValue('email', email)
        form.setFieldValue('phone', phone)
        form.setFieldValue('corporation', corporation == 'undefined' ? '' : corporation)
      }
    } catch (error) {
      console.error(error)
    }
  }

  useEffect(() => {
    getUserInfo()
  }, [])

  //更新用户信息
  function updateUserInfo() {
    form
      .validateFields(['email', 'corporation', 'phone'])
      .then(async ({ email, corporation, phone }) => {
        const params = {
          userId: user.userId,
          email,
          corporation,
          phone,
        }
        await updateUserInfoApi(params)
        showToast({
          title: '用户信息更新成功',
          color: 'success',
        })
        form.resetFields()
      })
      .then(() => {
        getUserInfo()
      })
  }

  return (
    <div className="w-2/3 flex flex-col flex-wrap justify-start">
      <Form form={form} requiredMark={true} layout="vertical">
        <Form.Item
          label={<p className="text-md">邮件</p>}
          name="email"
          rules={[
            {
              type: 'email',
              message: '请输入正确的邮箱格式',
            },
            {
              required: true,
              message: '邮箱不能为空',
            },
          ]}
        >
          <Input placeholder="请输入邮箱" className="w-80" />
        </Form.Item>
        <Form.Item
          label={<p className="text-md">手机号</p>}
          name="phone"
          rules={[
            { required: true, message: '请输入手机号' },
            { pattern: /^1[3-9]\d{9}$/, message: '请输入有效的手机号' },
          ]}
        >
          <Input placeholder="请输入手机号" className="w-80" />
        </Form.Item>
        <Form.Item label={<p className="text-md">组织</p>} name="corporation">
          <Input placeholder="请输入组织名" className="w-80" />
        </Form.Item>
        <Popconfirm title="确定要修改信息吗" onConfirm={updateUserInfo}>
          <Button type="primary">修改信息</Button>
        </Popconfirm>
      </Form>
    </div>
  )
}
