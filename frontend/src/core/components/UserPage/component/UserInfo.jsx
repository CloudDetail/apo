import { Button, Flex, Popconfirm, Form, Input, Collapse, Divider } from "antd"
import { MailOutlined, ApartmentOutlined, LockOutlined, PhoneOutlined } from '@ant-design/icons'
import { updateEmailApi, updateCorporationApi, updatePhoneApi, getUserInfoApi } from "src/core/api/user"
import { showToast } from "src/core/utils/toast"
import { useEffect, useState } from "react"

export default function UserInfo() {
    const [form] = Form.useForm()

    async function getUserInfo() {
        try {
            const result = await getUserInfoApi()
            form.setFieldValue("email", result.email)
            form.setFieldValue("phone", result.phone)
            form.setFieldValue("corporation", result.corporation == 'undefined' ? "" : result.corporation)
            localStorage.setItem("user", JSON.stringify(result))
        } catch (error) {
            showToast({
                title: error,
                color: 'danger'
            })
        }
    }

    useEffect(() => {
        const user = JSON.parse(localStorage.getItem("user"))
        if (user) {
            form.setFieldValue("email", user.email)
            form.setFieldValue("phone", user.phone)
            form.setFieldValue("corporation", user.corporation == 'undefined' ? "" : user.corporation)
        }
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
            <Flex vertical className="w-2/3">
                <Flex vertical justify="start" className="w-full">
                    <Form form={form} requiredMark={true} layout="vertical">
                        <Flex className="flex flex-col justify-between">
                            <Flex className="flex items-center">
                                <Form.Item
                                    label={<p className="text-base">邮件</p>}
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
                                    <Input placeholder="请输入邮箱" className="w-80 h-10" />
                                </Form.Item>
                                <Popconfirm
                                    title="确定要修改邮箱吗"
                                    okText="确定"
                                    onConfirm={updateEmail}
                                >
                                    <Button type="link" className="text-base">修改邮箱</Button>
                                </Popconfirm>
                            </Flex>
                        </Flex>
                        <Flex className="flex flex-col justify-betwwen w-full">
                            <Flex className="flex items-center">
                                <Form.Item
                                    label={<p className="text-base">手机号</p>}
                                    name="phone"
                                    rules={[
                                        { required: true, message: '请输入手机号' },
                                        { pattern: /^1[3-9]\d{9}$/, message: '请输入有效的手机号' }
                                    ]}
                                >
                                    <Input placeholder="请输入手机号" className="w-80 h-10" />
                                </Form.Item>
                                <Popconfirm
                                    title="确定要修改手机号吗"
                                    okText="确定"
                                    onConfirm={updatePhone}
                                >
                                    <Button type="link" className="text-base">修改手机号</Button>
                                </Popconfirm>
                            </Flex>
                        </Flex>

                        <Flex className="flex flex-col justify-betwwen">
                            <Flex className="flex items-center">
                                <Form.Item
                                    label={<p className="text-base">组织</p>}
                                    name="corporation"
                                >
                                    <Input placeholder="请输入组织名" className="w-80 h-10" />
                                </Form.Item>
                                <Popconfirm
                                    title="确定要修改组织吗"
                                    okText="确定"
                                    onConfirm={updateCorporation}
                                >
                                    <Button type="link" className="text-base">修改组织</Button>
                                </Popconfirm>
                            </Flex>
                        </Flex>
                    </Form>
                </Flex>
            </Flex>
        </Flex>
    )
}