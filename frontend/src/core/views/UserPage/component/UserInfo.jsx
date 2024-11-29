import { Button, Flex, Popconfirm, Form, Input, Collapse, Divider } from "antd"
import { MailOutlined, ApartmentOutlined, LockOutlined, PhoneOutlined } from '@ant-design/icons'
import { updateEmailApi, updateCorporationApi, updatePhoneApi, getUserInfoApi } from "core/api/user"
import { showToast } from "core/utils/toast"
import { useEffect, useState } from "react"

export default function UserInfo() {
    const [form] = Form.useForm()

    async function getUserInfo() {
        try {
            const user = await getUserInfoApi()
            if (user) {
                const { email, phone, corporation } = user
                form.setFieldValue("email", email)
                form.setFieldValue("phone", phone)
                form.setFieldValue("corporation", corporation == 'undefined' ? "" : corporation)
            }
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
        form.validateFields(['email', 'corporation', 'phone'])
            .then(async ({ email, corporation, phone }) => {
                const params = {
                    email,
                    corporation,
                    phone
                }
                await updateCorporationApi({ username: JSON.parse(localStorage.getItem("user"))?.username, ...params })
                showToast({
                    title: '邮箱更新成功',
                    color: 'success'
                })
                form.resetFields()
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
                                    label={<p className="text-md">邮件</p>}
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
                                    <Input placeholder="请输入邮箱" className="w-80" />
                                </Form.Item>
                            </Flex>
                        </Flex>
                        <Flex className="flex flex-col justify-betwwen w-full">
                            <Flex className="flex items-center">
                                <Form.Item
                                    label={<p className="text-md">手机号</p>}
                                    name="phone"
                                    rules={[
                                        { required: true, message: '请输入手机号' },
                                        { pattern: /^1[3-9]\d{9}$/, message: '请输入有效的手机号' }
                                    ]}
                                >
                                    <Input placeholder="请输入手机号" className="w-80" />
                                </Form.Item>
                            </Flex>
                        </Flex>
                        <Flex className="flex flex-col justify-betwwen">
                            <Flex className="flex items-center">
                                <Form.Item
                                    label={<p className="text-md">组织</p>}
                                    name="corporation"
                                >
                                    <Input placeholder="请输入组织名" className="w-80" />
                                </Form.Item>
                            </Flex>
                        </Flex>
                        <Button type="primary" onClick={updateEmail}>修改信息</Button>
                    </Form>
                </Flex>
            </Flex>
        </Flex>
    )
}
