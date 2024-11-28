import { Flex, Popover } from "antd";
import { LogoutOutlined, UserOutlined } from "@ant-design/icons"
import { showToast } from "core/utils/toast";
import { useNavigate } from 'react-router-dom'
import { logoutApi } from "core/api/user";
import { HiUserCircle } from "react-icons/hi";


const UserToolBox = () => {
    const navigate = useNavigate()

    const content = (
        <>
            <Flex vertical className={"flex items-center w-36 rounded-lg z-50"} >
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
        <Popover content={content}>
            <div
                className="relative flex items-center select-none w-auto pl-2 pr-2 rounded-md hover:bg-[#30333C] cursor-pointer"
            >
                <div>
                    <HiUserCircle className="w-8 h-8" />
                </div>
                <div className="h-1/2 flex flex-col justify-center">
                    <p className="text-base relative -top-0.5">{JSON.parse(localStorage.getItem("user"))?.username || "获取用户信息失败"}</p>
                </div>
            </div>
        </Popover>
    )
}

export default UserToolBox
