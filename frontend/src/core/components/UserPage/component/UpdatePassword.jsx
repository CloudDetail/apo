import { Form, Input, Popconfirm, Button, Flex, Divider } from "antd"
import { showToast } from "src/core/utils/toast"
import { updatePasswordApi } from "src/core/api/user"
import { LockOutlined } from "@ant-design/icons"


export default function UpdatePassword() {
    const [form] = Form.useForm()

    //更新密码
    function updatePassword() {
        form.validateFields(['oldPassword', 'newPassword'])
            .then(async (values) => {
                try {
                    await updatePasswordApi(values)
                    showToast({
                        title: '密码重设成功',
                        color: 'success'
                    })
                    form.resetFields(['oldPassword', 'newPassword'])
                } catch (error) {
                    console.log(error.response.data.code)
                    if (error.response.data.code) {
                        showToast({
                            title: "旧密码错误,请检查",
                            color: "danger"
                        })
                    }
                }
            })
    }

    return (
        <>
            <Flex vertical className="w-full">
                <Flex className="mb-10">
                    <Divider orientation="left">密码修改</Divider>
                </Flex>
                <Flex vertical className="w-1/3">
                    <Flex vertical justify="start">
                        <Form form={form} requiredMark={false}>
                            <Form.Item
                                label="旧密码"
                                name="oldPassword"
                                validateTrigger={false}
                                rules={[
                                    { required: true, message: '请输入旧密码' }
                                ]}
                            >
                                <Input.Password prefix={<LockOutlined />} placeholder="请输入旧密码" type="password" className="w-60" />
                            </Form.Item>
                            <Form.Item
                                label="新密码"
                                name="newPassword"
                                validateTrigger={false}
                                rules={[
                                    { required: true, message: '请输入新密码' }
                                ]}
                            >
                                <Input.Password prefix={<LockOutlined />} placeholder="请输入新密码" type="password" className="w-60" />
                            </Form.Item>
                            <div className="w-60 flex justify-start">
                                <Popconfirm
                                    title="确定要修改密码吗"
                                    okText="确定"
                                    onConfirm={updatePassword}
                                >
                                    <Button>修改密码</Button>
                                </Popconfirm>
                            </div>
                        </Form>
                    </Flex>
                </Flex>
            </Flex>
        </>
    )
}