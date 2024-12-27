/**
 * Copyright 2024 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */
import React from 'react'
import { Modal, Flex, Form, Input, Divider, Button, Tooltip } from 'antd'
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
import { useTranslation } from 'react-i18next' // 添加i18n

const EditModal = React.memo(
  ({ selectedUser, modalEditVisibility, setModalEditVisibility, getUserList }) => {
  const { t } = useTranslation('oss/userManage') // 使用i18n
  const [loading, setLoading] = useState(false)
  const [form] = Form.useForm()

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
                <Flex vertical className="w-full mt-4 mb-4 justify-center align-center">
                    <div>
                        <Form form={form} layout="vertical">
                            <Form.Item
                                label={t('editModal.username')}
                                name="username"
                            >
                                <Input disabled={true} />
                            </Form.Item>
                            <Form.Item
                                label={t('editModal.email')}
                                name="email"
                                rules={[
                                    {
                                        type: "email",
                                        message: t('editModal.emailInvalid'),
                                    }
                                ]}
                            >
                                <Input placeholder={t('editModal.emailPlaceholder')} />
                            </Form.Item>
                            <Form.Item
                                label={t('editModal.phone')}
                                name="phone"
                                rules={[
                                    {

                                        pattern: /^1[3-9]\d{9}$/, // 中国大陆手机号正则
                                        message: t('editModal.phoneInvalid'),
                                    }
                                ]}
                            >
                                <Input placeholder={t('editModal.phonePlaceholder')} />
                            </Form.Item>
                            <Form.Item
                                label="组织"
                                name="corporation"
                            >
                                <Input placeholder="请输入组织" />
                            </Form.Item>
                            <Button type="primary" onClick={editUser}>修改用户信息</Button>
                            <Divider />
                            <div className="mt-3">
                                <Form.Item
                                    label="新密码"
                                    name="newPassword"
                                    rules={[
                                        {
                                            required: true,
                                            message: "请输入密码"
                                        },
                                        {
                                            pattern: /^(?=.*[a-z])(?=.*[A-Z])(?=.*[!@#$%^&*()\-_+=<>?/{}[\]|:;.,~]).{9,}$/,
                                            message: <p>密码必须包含大写字母、小写字母、<Tooltip title="(! @ # $ % ^ & * ( ) - _ + = < > ? / { } [ ] | : ; . , ~)" ><span className="underline">特殊字符</span></Tooltip>，且长度大于8</p>
                                        }
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
                                            message: "请确认密码"
                                        },
                                        ({ getFieldValue }) => ({
                                            validator(_, value) {
                                                if (!value || getFieldValue('newPassword') === value) {
                                                    return Promise.resolve();
                                                }
                                                return Promise.reject(new Error('两次输入的密码不一致'));
                                            },
                                        }),
                                    ]}
                                >
                                    <Input.Password placeholder="请重复密码" />
                                </Form.Item>
                            </div>
                            <Button type="primary" onClick={resetPassword}>修改密码</Button>
                        </Form>
                    </div>
                </Flex>
            </Modal>
        </>
    )
}

export default EditModal
