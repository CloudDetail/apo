import { Button, Flex, Popconfirm } from "antd";
import { EditOutlined, LogoutOutlined, UserOutlined } from "@ant-design/icons"
import { showToast } from "src/core/utils/toast";
import { useNavigate } from 'react-router-dom'
import { logoutApi } from "src/core/api/user";
import { forwardRef, useState } from "react";


const ToolBox = forwardRef((props, ref) => {
    const { visiable, setVisiable } = props
    const navigate = useNavigate()

    //退出登录
    async function logout() {
        try {
            const params = {
                accessToken: localStorage.getItem("token"),
                refreshToken: localStorage.getItem("refreshToken")
            }
            await logoutApi(params)
            localStorage.removeItem("token")
            localStorage.removeItem("refreshToken")
            navigate('/login')
            showToast({
                title: '退出登录成功',
                color: 'success'
            })
        } catch (errorInfo) {
            showToast({
                title: '退出登录失败',
                message: '失败原因:' + errorInfo,
                color: 'danger'
            })
        }
    }


    return (
        <>
            <Flex ref={ref} vertical className={visiable ? "flex items-center absolute top-10 right-0 w-36 rounded-lg bg-[#171A21] pt-3 pb-3" : "hidden"} onClick={(e) => {
                setVisiable(false)
                e.stopPropagation()
            }}>
                <Flex vertical className="justify-center items-center w-full h-9 hover:bg-[#292E3B]" onClick={() => navigate('/user')}>
                    <Flex className="w-2/3 justify-around p-2">
                        <UserOutlined className="text-md" />
                        <p className="text-md select-none">
                            个人中心
                        </p>
                    </Flex>
                </Flex>
                <Flex vertical className="justify-center items-center w-full h-9 mt-2 hover:bg-[#292E3B]" onClick={logout}>
                    <Flex className="w-2/3 justify-around p-2" onClick={() => console.log("he;")}>
                        <LogoutOutlined className="text-md" />
                        <p className="text-md select-none">
                            退出登录
                        </p>
                    </Flex>
                </Flex>
            </Flex>
        </>
    )
})

export default ToolBox
