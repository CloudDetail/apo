import UserInfo from './component/UserInfo'
import { Menu, Flex, Splitter, Divider } from 'antd'
import { useState } from 'react'
import { EditOutlined, PlusOutlined, KeyOutlined } from '@ant-design/icons'
import UpdatePassword from './component/UpdatePassword'
import { LiaUser } from 'react-icons/lia'
import { PiUsersDuotone } from 'react-icons/pi'
import { IoMdLock } from 'react-icons/io'
import { useTranslation } from 'react-i18next' // 添加i18n

export default function UserPage() {
  const { t } = useTranslation('oss/userPage') // 使用i18n
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
