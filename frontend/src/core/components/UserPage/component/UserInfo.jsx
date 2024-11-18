import { Button, Flex, Popconfirm, Form, Input, Collapse, Divider } from "antd"
import { MailOutlined, ApartmentOutlined, LockOutlined, PhoneOutlined } from '@ant-design/icons'
import { updateEmailApi, updateCorporationApi, updatePhoneApi, getUserInfoApi } from "src/core/api/user"
import { showToast } from "src/core/utils/toast"
import { useEffect, useState } from "react"

export default function UserInfo() {
    const [form] = Form.useForm()
    const [username, setUsername] = useState("")

    async function getUserInfo() {
        try {
            const result = await getUserInfoApi()
            form.setFieldValue("email", result.email)
            form.setFieldValue("phone", result.phone)
            form.setFieldValue("corporation", result.corporation == 'undefined' ? "" : result.corporation)
            setUsername(result.username)
        } catch (error) {
            showToast({
                title: error,
                color: 'danger'
            })
        }
    }

    useEffect(() => {
        getUserInfo()
    }, [])

    //更新邮箱
    function updateEmail() {
        form.validateFields(['email'])
            .then(async (values) => {
                await updateEmailApi(values)
                showToast({
                    title: '邮箱更新成功',
                    color: 'success'
                })
                form.resetFields(['email'])
            })
            .then(() => {
                getUserInfo()
            })
    }

    //更新个人信息
    function updateCorporation() {
        form.validateFields(['corporation'])
            .then(async (values) => {
                await updateCorporationApi(values)
                showToast({
                    title: '个人信息更新成功',
                    color: 'success'
                })
                form.resetFields(['corporation'])
            })
            .then(() => {
                getUserInfo()
            })

    }

    //修改手机号
    function updatePhone() {
        form.validateFields(['phone'])
            .then(async (values) => {
                await updatePhoneApi(values)
                showToast({
                    title: '手机号修改成功',
                    color: 'success'
                })
                form.resetFields(['phone'])
            })
            .then(() => {
                getUserInfo()
            })
    }

    return (
        <Flex vertical className="w-full flex-wrap">
            <Flex vertical className="mb-10">
                <Divider orientation="left">个人信息: {username}</Divider>
            </Flex>
            <Flex vertical className="w-2/3">
                <Flex vertical justify="start" className="w-full">
                    <Form form={form} requiredMark={false}>
                        <Flex className="flex flex-col justify-between">
                            <Flex>
                                <Form.Item
                                    label="邮&#12288;件"
                                    name="email"
                                    rules={[
                                        {
                                            type: 'email',
                                            message: '请输入正确的邮箱格式'
                                        },
                                        {
                                            required: true,
                                            message: '邮箱不能为空'
                                        }
                                    ]}
                                >
                                    <Input placeholder="请输入邮箱" prefix={<MailOutlined />} className="w-60" />
                                </Form.Item>
                                <Popconfirm
                                    title="确定要修改邮箱吗"
                                    okText="确定"
                                    onConfirm={updateEmail}
                                >
                                    <Button type="link">修改邮箱</Button>
                                </Popconfirm>
                            </Flex>
                        </Flex>
                        <Flex className="flex flex-col justify-betwwen w-full">
                            <Flex>
                                <Form.Item
                                    label="手机号"
                                    name="phone"
                                    rules={[
                                        { required: true, message: '请输入手机号' },
                                        { pattern: /^1[3-9]\d{9}$/, message: '请输入有效的手机号' }
                                    ]}
                                >
                                    <Input placeholder="请输入手机号" prefix={<PhoneOutlined />} className="w-60" />
                                </Form.Item>
                                <Popconfirm
                                    title="确定要修改手机号吗"
                                    okText="确定"
                                    onConfirm={updatePhone}
                                >
                                    <Button type="link">修改手机号</Button>
                                </Popconfirm>
                            </Flex>
                        </Flex>

                        <Flex className="flex flex-col justify-betwwen">
                            <Flex >
                                <Form.Item
                                    label="组&#12288;织"
                                    name="corporation"
                                >
                                    <Input placeholder="请输入组织名" prefix={<ApartmentOutlined />} className="w-60" />
                                </Form.Item>
                                <Popconfirm
                                    title="确定要修改组织吗"
                                    okText="确定"
                                    onConfirm={updateCorporation}
                                >
                                    <Button type="link">修改组织</Button>
                                </Popconfirm>
                            </Flex>
                        </Flex>
                    </Form>
                </Flex>
            </Flex>
        </Flex>
    )
}