import { Modal, Flex, Form, Input } from "antd"
import { showToast } from "core/utils/toast"
import { createUserApi, updateEmailApi, updatePhoneApi, updateCorporationApi } from "core/api/user"
import { AiOutlineLoading } from "react-icons/ai";
import { useState } from "react"
import LoadingSpinner from 'src/core/components/Spinner'


const AddModal = ({ modalAddVisibility, setModalAddVisibility, getUserList }) => {
    const [addStatus, setAddStatus] = useState(false)
    const [form] = Form.useForm()

    //创建用户
    async function createUser() {
        if (addStatus) return
        form.validateFields()
            .then(async ({ username, password, confirmPassword, email, phone, corporation }) => {
                try {
                    setAddStatus(true)
                    const params = {
                        username: username,
                        password: password,
                        confirmPassword: confirmPassword
                    }
                    await createUserApi(params)
                    // @ts-ignore
                    if (email) {
                        await updateEmailApi({ username, email })
                    }
                    if (phone) {
                        await updatePhoneApi({ username, phone })
                    }
                    if (corporation) {
                        await updateCorporationApi({ username, corporation })
                    }
                    showToast({
                        title: "用户添加成功",
                        color: "success"
                    })
                    setAddStatus(false)
                    setModalAddVisibility(false)
                    await getUserList()
                } catch (error) {
                    setAddStatus(false)
                    console.log(error)
                    showToast({
                        title: error.response ? error.response.data.message : "未知错误",
                        color: 'danger'
                    })
                }
                form.resetFields()
            })
    }

    return (<>
        <Modal
            open={modalAddVisibility}
            onCancel={() => {
                if (!addStatus) {
                    setModalAddVisibility(false)
                }
            }}
            maskClosable={false}
            title="新增用户"
            okText={<span>新增</span>}
            onOk={createUser}
            width={1000}
        >
            <LoadingSpinner loading={addStatus} />
            <Flex vertical className="w-full mt-4 mb-4">
                <Flex vertical className="w-full justify-center start">
                    <Form form={form} layout="vertical">
                        <Form.Item
                            label="用户名;"
                            name="username"
                            rules={[
                                { required: true, message: '请输入用户名' }
                            ]}
                        >
                            <div className="flex justify-start items-start">
                                <Input placeholder="请输入用户名" />
                            </div>
                        </Form.Item>
                        <Form.Item
                            label="密码"
                            name="password"
                            rules={[
                                { required: true, message: '请输入密码' },
                                {
                                    pattern: /^(?=.*[a-z])(?=.*[A-Z])(?=.*[!@#$%^&*(),.?":{}|<>]).{9,}$/,
                                    message: '密码必须包含大写字母、小写字母、特殊字符，且长度大于8'
                                }
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
                                            return Promise.resolve();
                                        }
                                        return Promise.reject(new Error('两次输入的密码不一致'))
                                    }
                                })
                            ]}
                        >
                            <Input.Password placeholder="请再次输入密码" />
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
                    </Form>
                </Flex>
            </Flex>
        </Modal>
    </>)
}

export default AddModal
