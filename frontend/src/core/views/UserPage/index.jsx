import UserInfo from "./component/UserInfo"
import { Menu, Flex, Splitter, Divider } from "antd"
import { useState } from "react"
import { EditOutlined, PlusOutlined, KeyOutlined } from "@ant-design/icons"
import UpdatePassword from "./component/UpdatePassword"
import { LiaUser } from "react-icons/lia";
import { PiUsersDuotone } from "react-icons/pi";
import { IoMdLock } from "react-icons/io";

export default function UserPage() {
    return (
        <>
            <Flex vertical className="w-full mt-4 pl-12">
                <Divider orientation="left">基本信息</Divider>
                <UserInfo />
                <Divider orientation="left">修改密码</Divider>
                <UpdatePassword />
            </Flex>
        </>
    )
}