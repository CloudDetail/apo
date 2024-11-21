import { Form, Input, Popconfirm, Button, Flex, Divider } from "antd"
import { showToast } from "src/core/utils/toast"
import { updatePasswordApi } from "src/core/api/user"
import { LockOutlined } from "@ant-design/icons"


export default function UpdatePassword() {
    const [form] = Form.useForm()

    //更新密码
    function updatePassword() {
        form.validateFields(['oldPassword', 'newPassword', 'newPasswordAgain'])
            .then(async (values) => {
                try {
                    if (values.newPassword !== values.newPasswordAgain) {
                        showToast({
                            title: "两次密码不一致，请检查",
                            color: "danger"
                        })
                        return
                    }
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
                <Flex vertical className="w-1/3">
                    <Flex vertical justify="start">
                        <Form form={form} requiredMark={true} layout="vertical">
                            <Form.Item
                                label={<p className="text-base">旧密码</p>}
                                name="oldPassword"
                                validateTrigger={false}
                                rules={[
                                    { required: true, message: '请输入旧密码' }
                                ]}
                            >
                                <Input.Password placeholder="请输入旧密码" type="password" className="w-80 h-10" />
                            </Form.Item>
                            <Form.Item
                                label={<p className="text-base">新密码</p>}
                                name="newPassword"
                                validateTrigger={false}
                                rules={[
                                    { required: true, message: '请输入新密码' }
                                ]}
                            >
                                <Input.Password placeholder="请输入新密码" type="password" className="w-80 h-10" />
                            </Form.Item>
                            <Form.Item
                                label={<p className="text-base">确认新密码</p>}
                                name="newPasswordAgain"
                                validateTrigger={false}
                                rules={[
                                    { required: true, message: '请再次输入新密码' }
                                ]}
                            >
                                <Input.Password placeholder="请输入新密码" type="password" className="w-80 h-10" />
                            </Form.Item>
                            <div className="w-auto flex justify-start">
                                <Popconfirm
                                    title="确定要修改密码吗"
                                    okText="确定"
                                    onConfirm={updatePassword}
                                >
                                    <Button className="text-base">修改密码</Button>
                                </Popconfirm>
                            </div>
                        </Form>
                    </Flex>
                </Flex>
            </Flex>
        </>
    )
}