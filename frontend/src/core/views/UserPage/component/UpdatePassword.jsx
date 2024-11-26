import { Form, Input, Popconfirm, Button, Flex, Divider } from "antd"
import { showToast } from "core/utils/toast"
import { logoutApi, updatePasswordApi } from "core/api/user"
import { LockOutlined } from "@ant-design/icons"
import { useNavigate } from "react-router-dom"


export default function UpdatePassword() {
    const [form] = Form.useForm()
    const navigate = useNavigate();

    //更新密码
    function updatePassword() {
        form.validateFields(['oldPassword', 'newPassword', 'confirmPassword'])
            .then(async ({ oldPassword, newPassword, confirmPassword }) => {
                try {
                    const paramsForUpdatePassword = {
                        oldPassword,
                        newPassword,
                        confirmPassword,
                        username: JSON.parse(localStorage.getItem("user")).username
                    }
                    await updatePasswordApi(paramsForUpdatePassword)
                    form.resetFields(['oldPassword', 'newPassword', 'confirmPassword'])
                    const paramsForLogout = {
                        accessToken: localStorage.getItem("token"),
                        refreshToken: localStorage.getItem("refreshToken")
                    }
                    await logoutApi(paramsForLogout)
                    localStorage.removeItem("token")
                    localStorage.removeItem("refreshToken")
                    navigate("/login")
                    showToast({
                        title: '密码重设成功,请重新登录',
                        color: 'success'
                    })
                } catch (error) {
                    console.log(error.response.data.code)
                    showToast({
                        title: "错误",
                        message: error,
                        color: "danger"
                    })
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
                                label={<p className="text-md">旧密码</p>}
                                name="oldPassword"
                                rules={[
                                    { required: true, message: '请输入旧密码' }
                                ]}
                            >
                                <Input.Password placeholder="请输入旧密码" type="password" className="w-80" />
                            </Form.Item>
                            <Form.Item
                                label={<p className="text-md">新密码</p>}
                                name="newPassword"
                                rules={[
                                    { required: true, message: '请输入新密码' },
                                    {
                                        pattern: /^(?=.*[a-z])(?=.*[A-Z])(?=.*[!@#$%^&*(),.?":{}|<>]).{9,}$/,
                                        message: '密码必须包含大写字母、小写字母、特殊字符，且长度大于8'
                                    }
                                ]}
                            >
                                <Input.Password placeholder="请输入新密码" type="password" className="w-80" />
                            </Form.Item>
                            <Form.Item
                                label={<p className="text-md">确认新密码</p>}
                                name="confirmPassword"
                                dependencies={['newPassword']}
                                rules={[
                                    { required: true, message: '请再次输入新密码' },
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
                                <Input.Password placeholder="请输入新密码" type="password" className="w-80" />
                            </Form.Item>
                            <div className="w-auto flex justify-start">
                                <Popconfirm
                                    title="确定要修改密码吗"
                                    okText="确定"
                                    onConfirm={updatePassword}
                                >
                                    <Button className="text-md">修改密码</Button>
                                </Popconfirm>
                            </div>
                        </Form>
                    </Flex>
                </Flex>
            </Flex>
        </>
    )
}
