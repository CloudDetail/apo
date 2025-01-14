/**
 * Copyright 2024 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import UserInfo from './component/UserInfo'
import { Menu, Flex, Splitter, Divider } from 'antd'
import { useState } from 'react'
import { EditOutlined, PlusOutlined, KeyOutlined } from '@ant-design/icons'
import UpdatePassword from './component/UpdatePassword'
import { LiaUser } from 'react-icons/lia'
import { PiUsersDuotone } from 'react-icons/pi'
import { IoMdLock } from 'react-icons/io'
import { useTranslation } from 'react-i18next'

export default function UserPage() {
  const { t } = useTranslation('core/userPage')
  return (
    <>
      <Flex vertical className="w-full mt-4 pl-12 pb-20">
        <Divider orientation="left">{t('index.basicInfo')}</Divider>
        <UserInfo />
        <Divider orientation="left">{t('index.updatePassword')}</Divider>
        <UpdatePassword />
      </Flex>
    </>
  )
}
