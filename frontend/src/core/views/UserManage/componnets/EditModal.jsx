import { Modal, Flex, Form, Input, message, Divider, Button, Popconfirm } from "antd"
import { useEffect, useState } from "react"
import { AiFillWechatWork } from "react-icons/ai"
import { getUserListApi, updateEmailApi, updatePhoneApi, updateCorporationApi, updatePasswordWithNoOldPwd } from "core/api/user"
import { AiOutlineLoading } from "react-icons/ai";
import { GrPowerReset } from "react-icons/gr";
import { showToast } from "core/utils/toast";

const EditModal = ({ selectedUser, modalEditVisibility, setModalEditVisibility, getUserList }) => {
    const [loading, setLoading] = useState(false)
    const [saveStatus, setSaveStatus] = useState(false)
    const [form] = Form.useForm()

    useEffect(() => {
        console.log(selectedUser)
        getUserInfoByName()
    }, [selectedUser])

    const editUser = () => {
        if(saveStatus) return
        form.validateFields()
            .then(async ({ email, phone, corporation, newPassword }) => {
                let flag = false
                setSaveStatus(true)
                if (email) {
                    await updateEmailApi({ username: selectedUser, email })
                    flag = true
                }
                if (phone) {
                    await updatePhoneApi({ username: selectedUser, phone })
                    flag = true
                }
                if (corporation) {
                    await updateCorporationApi({ username: selectedUser, corporation })
                    flag = true
                }
                if (newPassword) {
                    await updatePasswordWithNoOldPwd({ username: selectedUser, newPassword })
                }
                if (flag) {
                    showToast({
                        title: "保存成功",
                        color: "success"
                    })
                } else {
                    showToast({
                        title: "未作任何修改"
                    })
                }
                setModalEditVisibility(false)
                getUserList()
            })
            .catch((error) => {
                console.log(error)
            }).finally(() => {
                setSaveStatus(false)
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
            console.log("result", users)
            form.setFieldsValue({
                username: users[0]?.username,
                email: users[0]?.email,
                phone: users[0]?.phone,
                corporation: users[0]?.corporation
            })
            setLoading(false)
        } catch (error) {
            console.log(error)
            setLoading(false)
        }
    }

    return (<>
        <Modal
            open={modalEditVisibility}
            onCancel={() => setModalEditVisibility(false)}
            maskClosable={false}
            title="编辑用户"
            okText={saveStatus ? (<div><AiOutlineLoading className="animate-spin" size={18} /></div>) : <span>保存</span>}
            onOk={editUser}
            width={1000}
        >
            <Flex vertical className="w-full mt-4 mb-4 justify-center align-center">
                {loading ? <AiOutlineLoading className="animate-spin" size={25} /> : <div>
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
                        {/* <Divider /> */}
                        <span>密码</span>
                        <div className="ml-7 mt-3">
                            <Form.Item
                                label="新密码"
                                name="newPassword"
                                rules={[
                                    {
                                        pattern: /^(?=.*[a-z])(?=.*[A-Z])(?=.*[!@#$%^&*(),.?":{}|<>]).{9,}$/,
                                        message: '密码必须包含大写字母、小写字母、特殊字符，且长度大于8'
                                    }
                                ]}
                            >
                                <Input placeholder="请输入密码" />
                            </Form.Item>
                            <Form.Item
                                label="重复新密码"
                                name="confirmPassword"
                                rules={[
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
                                <Input placeholder="请重复密码" />
                            </Form.Item>
                        </div>
                    </Form>
                </div>}
            </Flex>
        </Modal>
    </>)
}

export default EditModal
