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
        showToast({ title: t('EditModal.saveSuccess'), color: 'success' })
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
            title: t('EditModal.resetPasswordSuccess'),
            color: 'success',
          })
          setModalEditVisibility(false)
        } catch (error) {
          console.error(error)
          showToast({
            title: error.response?.data?.message || t('EditModal.resetPasswordFail'),
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
        title: t('EditModal.getUserInfoFail'),
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
        title={t('EditModal.title')}
        width={1000}
        footer={null}
      >
        <LoadingSpinner loading={loading} />
        <Flex vertical className="w-full mt-4 mb-4 justify-center align-center">
          <div>
            <Form form={form} layout="vertical">
              <Form.Item label={t('EditModal.username')} name="username">
                <Input disabled={true} />
              </Form.Item>
              <Form.Item
                label={t('EditModal.email')}
                name="email"
                rules={[
                  {
                    type: 'email',
                    message: t('EditModal.emailInvalid'),
                  },
                ]}
              >
                <Input placeholder={t('EditModal.emailPlaceholder')} />
              </Form.Item>
              <Form.Item
                label={t('EditModal.phone')}
                name="phone"
                rules={[
                  {
                    pattern: /^1[3-9]\d{9}$/, // 中国大陆手机号正则
                    message: t('EditModal.phoneInvalid'),
                  },
                ]}
              >
                <Input placeholder={t('EditModal.phonePlaceholder')} />
              </Form.Item>
              <Form.Item label={t('EditModal.corporation')} name="corporation">
                <Input placeholder={t('EditModal.corporationPlaceholder')} />
              </Form.Item>
              <Button type="primary" onClick={editUser}>
                {t('EditModal.save')}
              </Button>
              <Divider />
              <div className="mt-3">
                <Form.Item
                  label={t('EditModal.newPassword')}
                  name="newPassword"
                  rules={[
                    {
                      required: true,
                      message: t('EditModal.newPasswordRequired'),
                    },
                    {
                      pattern:
                        /^(?=.*[a-z])(?=.*[A-Z])(?=.*[!@#$%^&*()\-_+=<>?/{}[\]|:;.,~]).{9,}$/,
                      message: t('EditModal.newPasswordPattern'),
                    },
                  ]}
                >
                  <Input.Password placeholder={t('EditModal.newPasswordPlaceholder')} />
                </Form.Item>
                <Form.Item
                  label={t('EditModal.confirmPassword')}
                  name="confirmPassword"
                  rules={[
                    {
                      required: true,
                      message: t('EditModal.confirmPasswordRequired'),
                    },
                    ({ getFieldValue }) => ({
                      validator(_, value) {
                        if (!value || getFieldValue('newPassword') === value) {
                          return Promise.resolve()
                        }
                        return Promise.reject(new Error(t('EditModal.confirmPasswordMismatch')))
                      },
                    }),
                  ]}
                >
                  <Input.Password placeholder={t('EditModal.confirmPasswordPlaceholder')} />
                </Form.Item>
              </div>
              <Button type="primary" onClick={resetPassword}>
                {t('EditModal.resetPassword')}
              </Button>
            </Form>
          </div>
        </Flex>
      </Modal>
    </>
  )
}

export default EditModal
