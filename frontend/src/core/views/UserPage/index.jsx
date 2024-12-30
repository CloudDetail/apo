/**
 * Copyright 2024 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import UserInfo from './component/UserInfo'
import { Flex, Divider } from 'antd'
import UpdatePassword from './component/UpdatePassword'

export default function UserPage() {
  return (
    <>
      <Flex vertical className="w-full mt-4 pl-12 pb-20">
        <Divider orientation="left">基本信息</Divider>
        <UserInfo />
        <Divider orientation="left">修改密码</Divider>
        <UpdatePassword />
      </Flex>
    </>
  )
}
