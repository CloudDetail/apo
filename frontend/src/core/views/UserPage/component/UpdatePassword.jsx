/**
 * Copyright 2024 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import { Form, Input, Popconfirm, Button, Flex, Tooltip } from 'antd'
import { showToast } from 'core/utils/toast'
import { logoutApi, updatePasswordApi } from 'core/api/user'
import { LockOutlined } from '@ant-design/icons'
import { useNavigate } from 'react-router-dom'
import { useUserContext } from 'src/core/contexts/UserContext'
import { useTranslation } from 'react-i18next'
import { redirectToLogin } from 'src/core/utils/redirectToLogin'

export default function UpdatePassword() {
  const { t } = useTranslation('core/userPage')
  const [form] = Form.useForm()
  const navigate = useNavigate()
  const { user, dispatch } = useUserContext()

  //更新密码
  function updatePassword() {
    form
      .validateFields(['oldPassword', 'newPassword', 'confirmPassword'])
      .then(async ({ oldPassword, newPassword, confirmPassword }) => {
        try {
          const paramsForUpdatePassword = {
            oldPassword,
            newPassword,
            confirmPassword,
            userId: user?.userId,
          }
          await updatePasswordApi(paramsForUpdatePassword)
          form.resetFields(['oldPassword', 'newPassword', 'confirmPassword'])
          const paramsForLogout = {
            userId: user?.userId,
            accessToken: localStorage.getItem('token'),
            refreshToken: localStorage.getItem('refreshToken'),
          }
          await logoutApi(paramsForLogout)
          localStorage.removeItem('token')
          localStorage.removeItem('refreshToken')
          redirectToLogin(false)
          showToast({
            title: t('updatePassword.updateSuccess'),
            color: 'success',
          })
        } catch (error) {
          console.log(error.response.data.code)
          const errorMessage =
            error.response?.data?.message || error.message || t('updatePassword.updateFail')
          showToast({
            title: t('updatePassword.error'),
            message: errorMessage,
            color: 'danger',
          })
        }
      })
      .catch((error) => {
        showToast({
          title: t('updatePassword.formValidationFail'),
          message: error.message || t('updatePassword.formValidationFailMessage'),
          color: 'danger',
        })
      })
  }

  return (
    <>
      <Flex vertical className="w-full">
        <Flex vertical className="w-1/3">
          <Flex vertical justify="start">
            <Form form={form} requiredMark={true} layout="vertical">
              <Form.Item
                label={<p className="text-md">{t('updatePassword.oldPassword')}</p>}
                name="oldPassword"
                rules={[{ required: true, message: t('updatePassword.oldPasswordRequired') }]}
              >
                <Input.Password
                  placeholder={t('updatePassword.oldPasswordPlaceholder')}
                  type="password"
                  className="w-80"
                />
              </Form.Item>
              <Form.Item
                label={<p className="text-md">{t('updatePassword.newPassword')}</p>}
                name="newPassword"
                rules={[
                  { required: true, message: t('updatePassword.newPasswordRequired') },
                  {
                    pattern: /^(?=.*[a-z])(?=.*[A-Z])(?=.*[!@#$%^&*()\-_+=<>?/{}[\]|:;.,~]).{9,}$/,
                    message: <p>{t('updatePassword.newPasswordPattern')}</p>,
                  },
                ]}
              >
                <Input.Password
                  placeholder={t('updatePassword.newPasswordPlaceholder')}
                  type="password"
                  className="w-80"
                />
              </Form.Item>
              <Form.Item
                label={<p className="text-md">{t('updatePassword.confirmPassword')}</p>}
                name="confirmPassword"
                dependencies={['newPassword']}
                rules={[
                  { required: true, message: t('updatePassword.confirmPasswordRequired') },
                  ({ getFieldValue }) => ({
                    validator(_, value) {
                      if (!value || getFieldValue('newPassword') === value) {
                        return Promise.resolve()
                      }
                      return Promise.reject(new Error(t('updatePassword.confirmPasswordMismatch')))
                    },
                  }),
                ]}
              >
                <Input.Password
                  placeholder={t('updatePassword.confirmPasswordPlaceholder')}
                  type="password"
                  className="w-80"
                />
              </Form.Item>
              <div className="w-auto flex justify-start">
                <Popconfirm
                  title={t('updatePassword.confirmUpdate')}
                  okText={t('updatePassword.okText')}
                  onConfirm={updatePassword}
                >
                  <Button type="primary" className="text-md">
                    {t('updatePassword.updatePassword')}
                  </Button>
                </Popconfirm>
              </div>
            </Form>
          </Flex>
        </Flex>
      </Flex>
    </>
  )
}
