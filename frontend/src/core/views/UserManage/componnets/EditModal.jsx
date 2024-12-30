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
import { useTranslation } from 'react-i18next'

const EditModal = React.memo(
  ({ selectedUser, modalEditVisibility, setModalEditVisibility, getUserList }) => {
    const { t } = useTranslation('core/userManage')
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
                <Form.Item label={t('editModal.username')} name="username">
                  <Input disabled={true} />
                </Form.Item>
                <Form.Item
                  label={t('editModal.email')}
                  name="email"
                  rules={[
                    {
                      type: 'email',
                      message: t('editModal.emailInvalid'),
                    },
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
                    },
                  ]}
                >
                  <Input placeholder={t('editModal.phonePlaceholder')} />
                </Form.Item>
                <Form.Item label={t('editModal.corporation')} name="corporation">
                  <Input placeholder={t('editModal.corporationPlaceholder')} />
                </Form.Item>
                <Button type="primary" onClick={editUser}>
                  {t('editModal.save')}
                </Button>
                <Divider />
                <div className="mt-3">
                  <Form.Item
                    label={t('editModal.newPassword')}
                    name="newPassword"
                    rules={[
                      {
                        required: true,
                        message: t('editModal.newPasswordPlaceholder'),
                      },
                      {
                        pattern:
                          /^(?=.*[a-z])(?=.*[A-Z])(?=.*[!@#$%^&*()\-_+=<>?/{}[\]|:;.,~]).{9,}$/,
                        message: t('editModal.newPasswordPattern'),
                      },
                    ]}
                  >
                    <Input.Password placeholder={t('editModal.newPasswordPlaceholder')} />
                  </Form.Item>
                  <Form.Item
                    label={t('editModal.confirmPassword')}
                    name="confirmPassword"
                    rules={[
                      {
                        required: true,
                        message: t('editModal.confirmPasswordRequired'),
                      },
                      ({ getFieldValue }) => ({
                        validator(_, value) {
                          if (!value || getFieldValue('newPassword') === value) {
                            return Promise.resolve()
                          }
                          return Promise.reject(new Error(t('editModal.confirmPasswordMismatch')))
                        },
                      }),
                    ]}
                  >
                    <Input.Password placeholder={t('editModal.confirmPasswordPlaceholder')} />
                  </Form.Item>
                </div>
                <Button type="primary" onClick={resetPassword}>
                  {t('editModal.resetPassword')}
                </Button>
              </Form>
            </div>
          </Flex>
        </Modal>
      </>
    )
  },
)

export default EditModal
