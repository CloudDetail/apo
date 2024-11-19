import { Flex, Form, Input, Button, Popconfirm, Divider, Tooltip, Modal } from "antd"
import { UserOutlined, LockOutlined } from "@ant-design/icons"
import { createUserApi } from "src/core/api/user";
import { showToast } from "src/core/utils/toast";
import { IoPersonAdd } from "react-icons/io5";
import { useState } from "react";


export default function UserManage() {
    const [form] = Form.useForm()
    const [modalVisibility, setModalVisibility] = useState(false)
    function createUser() {
        form.validateFields()
            .then(async (values) => {
                console.log(values)
                if (values.password !== values.confirmPassword) {
                    showToast({
                        title: '两次密码不一致,请重新输入',
                        color: 'danger'
                    })
                    return
                }
                try {
                    await createUserApi(values)
                    showToast({
                        title: '新用户添加成功',
                        color: 'success'
                    })
                } catch (error) {
                    showToast({
                        title: error.response ? error.response.data.message : "未知错误",
                        color: 'danger'
                    })
                }
                form.resetFields()
            })
    }

    return (
        <Flex vertical className="w-full">
            <Flex className="mb-10">
                <Divider orientation="left">用户管理</Divider>
            </Flex>
            <Flex className="w-full justify-end">
                <Button onClick={() => setModalVisibility(true)}><Tooltip title='新增用户'><IoPersonAdd className="w-6 h-6" /></Tooltip></Button>
            </Flex>
            <Modal
                open={modalVisibility}
                onCancel={() => setModalVisibility(false)}
                maskClosable={false}
                title="新增用户"
                okText="新增"
                onOk={createUser}
            >
                <Flex vertical className="w-full mt-4 mb-4">
                    <Flex vertical className="w-full justify-center items-center">
                        <Form form={form} requiredMark={false}>
                            <Form.Item
                                label="用户名&#12288;"
                                name="username"
                                rules={[
                                    { required: true, message: '请输入用户名' }
                                ]}
                            >
                                <div className="flex justify-start items-start">
                                    <Input prefix={<UserOutlined />} placeholder="请输入用户名" className="w-80" />
                                </div>
                            </Form.Item>
                            <Form.Item
                                label="密&#12288;&#12288;码"
                                name="password"
                                rules={[
                                    { required: true, message: '请输入密码' }
                                ]}
                            >
                                <div className="flex justify-start items-start">
                                    <Input.Password prefix={<LockOutlined />} placeholder="请输入密码" className="w-80" />
                                </div>
                            </Form.Item>
                            <Form.Item
                                label="重复密码"
                                name="confirmPassword"
                                rules={[
                                    { required: true, message: '请再次输入密码' }
                                ]}
                            >
                                <Input.Password prefix={<LockOutlined />} placeholder="请再次输入密码" className="w-80" />
                            </Form.Item>
                        </Form>
                    </Flex>
                </Flex>
            </Modal>
        </Flex>
    )
}