import { Modal, Flex, Form, Input, Tooltip } from 'antd'
import { showToast } from 'core/utils/toast'
import { createUserApi } from 'core/api/user'
import { useState } from 'react'
import LoadingSpinner from 'src/core/components/Spinner'
import { useTranslation } from 'react-i18next' // 添加i18n

const AddModal = ({ modalAddVisibility, setModalAddVisibility, getUserList }) => {
  const { t } = useTranslation('oss/userManage') // 使用i18n
  const [loading, setLoading] = useState(false)
  const [form] = Form.useForm()

  //创建用户
  async function createUser() {
    if (loading) return //防止重复提交
    form
      .validateFields()
      .then(
        async ({
          username,
          password,
          confirmPassword,
          email = '',
          phone = '',
          corporation = '',
        }) => {
          try {
            //设置加载状态
            setLoading(true)
            //创建用户
            const params = { username, password, confirmPassword, email, phone, corporation }
            await createUserApi(params)
            // 操作成功的反馈和状态清理
            setModalAddVisibility(false)
            await getUserList()
            showToast({ title: t('addModal.addSuccess'), color: 'success' })
          } catch (error) {
            console.error(error)
          } finally {
            setLoading(false)
            form.resetFields()
          }
        },
      )
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
        title={t('addModal.title')}
        okText={<span>{t('addModal.add')}</span>}
        cancelText={<span>{t('addModal.cancel')}</span>}
        onOk={createUser}
        width={1000}
      >
        <LoadingSpinner loading={loading} />
        <Flex vertical className="w-full mt-4 mb-4">
          <Flex vertical className="w-full justify-center start">
            <Form form={form} layout="vertical">
              <Form.Item
                label={t('addModal.username')}
                name="username"
                rules={[{ required: true, message: t('addModal.usernameRequired') }]}
              >
                <div className="flex justify-start items-start">
                  <Input placeholder={t('addModal.usernamePlaceholder')} />
                </div>
              </Form.Item>
              <Form.Item
                label={t('addModal.password')}
                name="password"
                rules={[
                  { required: true, message: t('addModal.passwordRequired') },
                  {
                    pattern: /^(?=.*[a-z])(?=.*[A-Z])(?=.*[!@#$%^&*()\-_+=<>?/{}[\]|:;.,~]).{9,}$/,
                    message: t('addModal.passwordPattern'),
                  },
                ]}
              >
                <div className="flex justify-start items-start">
                  <Input.Password placeholder={t('addModal.passwordPlaceholder')} />
                </div>
              </Form.Item>
              <Form.Item
                label={t('addModal.confirmPassword')}
                name="confirmPassword"
                dependencies={['password']}
                rules={[
                  { required: true, message: t('addModal.confirmPasswordRequired') },
                  ({ getFieldValue }) => ({
                    validator(_, value) {
                      if (!value || getFieldValue('password') === value) {
                        return Promise.resolve()
                      }
                      return Promise.reject(new Error(t('addModal.confirmPasswordMismatch')))
                    },
                  }),
                ]}
              >
                <Input.Password placeholder={t('addModal.confirmPasswordPlaceholder')} />
              </Form.Item>
              <Form.Item
                label={t('addModal.email')}
                name="email"
                rules={[
                  {
                    type: 'email',
                    message: t('addModal.emailInvalid'),
                  },
                ]}
              >
                <Input placeholder={t('addModal.emailPlaceholder')} />
              </Form.Item>
              <Form.Item
                label={t('addModal.phone')}
                name="phone"
                rules={[
                  {
                    pattern: /^1[3-9]\d{9}$/, // 中国大陆手机号正则
                    message: t('addModal.phoneInvalid'),
                  },
                ]}
              >
                <Input placeholder={t('addModal.phonePlaceholder')} />
              </Form.Item>
              <Form.Item label={t('addModal.corporation')} name="corporation">
                <Input placeholder={t('addModal.corporationPlaceholder')} />
              </Form.Item>
            </Form>
          </Flex>
        </Flex>
      </Modal>
    </>
  )
}

export default AddModal
