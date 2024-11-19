import React, { useState, useEffect } from "react";
import { Button, Input, Form, Flex, Checkbox } from "antd";
import { loginApi } from "src/core/api/user";
import { useNavigate } from "react-router-dom";
import { UserOutlined, LockOutlined } from "@ant-design/icons"
import { showToast } from "src/core/utils/toast";

export default function Login() {
    const navigate = useNavigate();
    const [form] = Form.useForm();
    const [remeberMe, setRemeberMe] = useState(true)

    const login = () => {
        form.validateFields()
            .then(async (values) => {
                try {
                    // @ts-ignore
                    const { accessToken, refreshToken } = await loginApi(values);
                    if (accessToken && refreshToken) {
                        window.localStorage.setItem("token", accessToken)
                        window.localStorage.setItem("refreshToken", refreshToken)
                        navigate("/service");
                        showToast({
                            title: "登录成功",
                            color: "success"
                        })
                        if (remeberMe) {
                            localStorage.setItem("username", values.username)
                        } else {
                            localStorage.removeItem("username")
                        }
                        localStorage.setItem("remeberMe", String(remeberMe))
                    }
                } catch (error) {
                    if (error.response && error.response.data) {
                        const { code, message } = error.response.data
                        switch (code) {
                            case 'B0902':
                                showToast({
                                    title: message,
                                    color: 'danger'
                                })
                                break
                            case 'B0901':
                                showToast({
                                    title: "用户不存在",
                                    color: 'danger'
                                })
                                break
                        }

                    }
                }
            })
            .catch((errorInfo) => {
                console.log("验证失败:", errorInfo);
            });
    };

    useEffect(() => {
        form.setFieldValue("username", localStorage.getItem("username"))
        const savedRememberMe = localStorage.getItem("remeberMe");
        setRemeberMe(savedRememberMe ? JSON.parse(savedRememberMe) : false);

        const handleKeyDown = (e) => {
            if (e.key === "Enter") {
                login();
            }
        };

        // 添加全局键盘事件监听器
        window.addEventListener("keydown", handleKeyDown);

        // 在组件卸载时移除事件监听器
        return () => {
            window.removeEventListener("keydown", handleKeyDown);
        };
    }, [form]);

    return (
        <Flex vertical className="w-screen h-screen justify-center items-center bg-[url('bg.jpg')] bg-no-repeat bg-cover">
            <Flex vertical className="w-3/12 bg-[rgba(0,0,0,0.4)] rounded-lg p-10 drop-shadow-xl">
                <Flex className="w-full justify-center items-center select-none">
                    <img src="mini_logo.png" className="w-16"/>
                    <p className="text-2xl">向导式可观测平台</p>
                </Flex>
                <Flex vertical className="w-full justify-center items-center mt-20">
                    <Form form={form} className="w-full">
                        <label className="text-xs">
                            用户名
                        </label>
                        <Form.Item
                            name="username"
                            rules={[
                                { required: true, message: "请输入用户名" }
                            ]}
                        >
                            <Input size="large" className="w-full bg-[rgba(17,18,23,0.5)] hover:bg-[rgba(17,18,23,0.5)]" prefix={<UserOutlined />} />
                        </Form.Item>
                        <label className="text-xs">
                            密码
                        </label>
                        <Form.Item
                            name="password"
                            rules={[
                                { required: true, message: "请输入密码" }
                            ]}
                        >

                            <Input.Password size="large" className="w-full bg-[rgba(17,18,23,0.5)] hover:bg-[rgba(17,18,23,0.5)]" prefix={<LockOutlined />} />
                        </Form.Item>
                    </Form>
                    <Flex className="w-full justify-between items-start mt-14">
                        <Button size="large" onClick={login} className="bg-[#455EEB] border-none w-full border-none">登录</Button>
                    </Flex>
                </Flex>
            </Flex>
        </Flex>
    );
}
