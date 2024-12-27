import { Modal, Flex, Form, Input, Divider, Button, Tooltip } from 'antd'
import { useEffect, useState } from 'react'
import {
  getUserListApi,
  updateEmailApi,
  updatePhoneApi,
  updateCorporationApi,
  updatePasswordWithNoOldPwd,
} from 'core/api/user'
import { showToast } from 'core/utils/toast'
import LoadingSpinner from 'src/core/components/Spinner'
import { useTranslation } from 'react-i18next' // 添加i18n

const EditModal = ({ selectedUser, modalEditVisibility, setModalEditVisibility, getUserList }) => {
  const { t } = useTranslation('oss/userManage') // 使用i18n
  const [loading, setLoading] = useState(false)
  const [form] = Form.useForm()

  useEffect(() => {
    if (modalEditVisibility) {
      form.resetFields()
      getUserInfoByName()
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

        await updateCorporationApi({ username: selectedUser, ...params })

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
          await updatePasswordWithNoOldPwd({ username: selectedUser, ...params })
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
          setLoading(false)
        }
      })
  }

  const getUserInfoByName = async () => {
    try {
      setLoading(true)
      const params = {
        currentPage: 1,
        pageSize: 1,
        username: selectedUser,
        role: '',
        corporation: '',
      }
      const { users } = await getUserListApi(params)
      form.setFieldsValue({
        username: users[0]?.username,
        email: users[0]?.email,
        phone: users[0]?.phone,
        corporation: users[0]?.corporation,
      })
    } catch (error) {
      showToast({
        title: t('editModal.getUserInfoFail'),
        color: 'danger',
      })
      console.log(error)
    } finally {
      setLoading(false)
    }
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
                      message: t('editModal.newPasswordRequired'),
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
}

export default EditModal
