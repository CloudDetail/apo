import { Modal, Flex, Form, Input, Divider, Button } from "antd"
import { useEffect, useState } from "react"
import { getUserListApi, updateEmailApi, updatePhoneApi, updateCorporationApi, updatePasswordWithNoOldPwd } from "core/api/user"
import { showToast } from "core/utils/toast";
import LoadingSpinner from 'src/core/components/Spinner'

const EditModal = ({ selectedUser, modalEditVisibility, setModalEditVisibility, getUserList }) => {
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
        form.validateFields(['email','phone','corporation'])
            .then(async ({ email = "", phone = "", corporation = "" }) => {
                setLoading(true)

                const params = {
                    email, phone,
                    corporation
                }

                await updateCorporationApi({ username: selectedUser, ...params })

                setModalEditVisibility(false)
                getUserList()
                showToast({ "title": "保存成功", "color": "success" })
                form.resetFields()
            })
            .catch((error) => {
                console.error(error)
            }).finally(() => {
                setLoading(false)
            })
    }

    const resetPassword = () => {
        if (loading) return
        form.validateFields(['newPassword','confirmPassword'])
            .then(async ({ newPassword, confirmPassword }) => {
                try {
                    setLoading(true)

                    const params = { newPassword, confirmPassword }
                    await updatePasswordWithNoOldPwd({ username: selectedUser, ...params })
                    showToast({
                        "title": "密码修改成功",
                        "color": "success"
                    })
                    setModalEditVisibility(false)
                } catch (error) {
                    console.error(error)
                    showToast({
                        title: error.response?.data?.message || "未知错误",
                        "color": "danger"
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
                "currentPage": 1,
                "pageSize": 1,
                "username": selectedUser,
                "role": "",
                "corporation": ""
            }
            const { users } = await getUserListApi(params)
            form.setFieldsValue({
                username: users[0]?.username,
                email: users[0]?.email,
                phone: users[0]?.phone,
                corporation: users[0]?.corporation
            })
        } catch (error) {
            showToast({
                title: "获取用户信息失败",
                color: "danger"
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
                title="编辑用户"
                width={1000}
                footer={null}
            >
                <LoadingSpinner loading={loading} />
                <Flex vertical className="w-full mt-4 mb-4 justify-center align-center">
                    <div>
                        <Form form={form} layout="vertical">
                            <Form.Item
                                label="用户名"
                                name="username"
                            >
                                <Input disabled={true} />
                            </Form.Item>
                            <Form.Item
                                label="邮件"
                                name="email"
                                rules={[
                                    {
                                        type: "email",
                                        message: "请输入有效的邮箱地址"
                                    }
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
                                        message: "请输入有效的电话号码",
                                    }
                                ]}
                            >
                                <Input placeholder="请输入电话号码" />
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
                                            required:true,
                                            message:"请输入密码"
                                        },
                                        {
                                            pattern: /^(?=.*[a-z])(?=.*[A-Z])(?=.*[!@#$%^&*(),.?":{}|<>]).{9,}$/,
                                            message: '密码必须包含大写字母、小写字母、特殊字符，且长度大于8'
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
                                            required:true,
                                            message:"请确认密码"
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
                                    <Input.Password  placeholder="请重复密码" />
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
