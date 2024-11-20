import UserInfo from "./component/UserInfo"
import { Menu, Flex } from "antd"
import UserManage from "./component/UserManage"
import { useState } from "react"
import { EditOutlined, PlusOutlined, KeyOutlined } from "@ant-design/icons"
import UpdatePassword from "./component/UpdatePassword"
import { LiaUser } from "react-icons/lia";
import { PiUsersDuotone } from "react-icons/pi";
import { IoMdLock } from "react-icons/io";

export default function UserPage() {
    const [currentItem, setCurrentItem] = useState("update")

    const items = [
        {
            key: 'update',
            label: '个人信息',
            icon: <LiaUser size={16} />
        },
        {
            key: 'add',
            label: '用户管理',
            icon: <PiUsersDuotone size={16} />
        },
        {
            key: 'updatePassword',
            label: '密码修改',
            icon: <IoMdLock size={16} />
        }
    ]

    function handleSelect(e) {
        setCurrentItem(e.key)
    }

    return (
        <>
            <Flex className="w-full mt-4">
                <Menu
                    mode="vertical"
                    items={items}
                    className="w-40 bg-[#1E222B] border-0"
                    defaultSelectedKeys={['update']}
                    onSelect={handleSelect}
                />
                <Flex className="w-11/12 ml-10 h-fit">
                    {currentItem == "update" ? <UserInfo /> : currentItem == "add" ? <UserManage /> : <UpdatePassword />}
                </Flex>
            </Flex>
        </>
    )
}