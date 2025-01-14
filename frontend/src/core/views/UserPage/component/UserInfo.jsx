/**
 * Copyright 2024 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import { Button, Flex, Popconfirm, Form, Input, Collapse, Divider } from 'antd'
import { MailOutlined, ApartmentOutlined, LockOutlined, PhoneOutlined } from '@ant-design/icons'
import { updateEmailApi, updateCorporationApi, updatePhoneApi, getUserInfoApi } from 'core/api/user'
import { showToast } from 'core/utils/toast'
import { useEffect, useState } from 'react'
import { useUserContext } from 'src/core/contexts/UserContext'
import { useTranslation } from 'react-i18next'

export default function UserInfo() {
  const { t } = useTranslation('core/userPage')
  const [form] = Form.useForm()
  const { user, dispatch } = useUserContext()

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
  function updateEmail() {
    form
      .validateFields(['email', 'corporation', 'phone'])
      .then(async ({ email, corporation, phone }) => {
        const params = {
          email,
          corporation,
          phone,
        }
        await updateCorporationApi({ userId: user.userId, ...params })
        showToast({
          title: t('userInfo.updateSuccess'),
          color: 'success',
        })
        form.resetFields()
      })
      .then(() => {
        getUserInfo()
      })
  }

  return (
    <Flex vertical className="w-full flex-wrap">
      <Flex vertical className="w-2/3">
        <Flex vertical justify="start" className="w-full">
          <Form form={form} requiredMark={true} layout="vertical">
            <Flex className="flex flex-col justify-between">
              <Flex className="flex items-center">
                <Form.Item
                  label={<p className="text-md">{t('userInfo.email')}</p>}
                  name="email"
                  rules={[
                    {
                      type: 'email',
                      message: t('userInfo.emailInvalid'),
                    },
                    {
                      required: true,
                      message: t('userInfo.emailRequired'),
                    },
                  ]}
                >
                  <Input placeholder={t('userInfo.emailPlaceholder')} className="w-80" />
                </Form.Item>
              </Flex>
            </Flex>
            <Flex className="flex flex-col justify-betwwen w-full">
              <Flex className="flex items-center">
                <Form.Item
                  label={<p className="text-md">{t('userInfo.phone')}</p>}
                  name="phone"
                  rules={[
                    { required: true, message: t('userInfo.phoneRequired') },
                    { pattern: /^1[3-9]\d{9}$/, message: t('userInfo.phoneInvalid') },
                  ]}
                >
                  <Input placeholder={t('userInfo.phonePlaceholder')} className="w-80" />
                </Form.Item>
              </Flex>
            </Flex>
            <Flex className="flex flex-col justify-betwwen">
              <Flex className="flex items-center">
                <Form.Item
                  label={<p className="text-md">{t('userInfo.corporation')}</p>}
                  name="corporation"
                >
                  <Input placeholder={t('userInfo.corporationPlaceholder')} className="w-80" />
                </Form.Item>
              </Flex>
            </Flex>
            <Popconfirm title={t('userInfo.confirmUpdate')} onConfirm={updateEmail}>
              <Button type="primary">{t('userInfo.okText')}</Button>
            </Popconfirm>
          </Form>
        </Flex>
      </Flex>
    </Flex>
  )
}
