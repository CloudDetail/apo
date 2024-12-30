/**
 * Copyright 2024 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import { Button, Dropdown, Popover } from 'antd'
import i18next from 'i18next'
import { useEffect, useState } from 'react'
import { IoLanguage } from 'react-icons/io5'
import { useDispatch, useSelector } from 'react-redux'

const LanguageSwitcher = () => {
  const [selectedKeys, setSelectedKeys] = useState([])
  const dispatch = useDispatch()
  const currentLanguage = useSelector((state) => state.settingReducer.language)

  const changeLanguage = (item) => {
    i18next.changeLanguage(item.key).then(() => {
      dispatch({ type: 'setLanguage', payload: item.key }) // 更新 Redux 中的语言状态
    })
  }
  const items = [
    {
      key: 'en',
      label: <div className="w-24 text-center">English</div>,
    },
    {
      key: 'zh',
      label: <div className="w-24 text-center">简体中文</div>,
    },
  ]
  useEffect(() => {
    if (!currentLanguage) {
      changeLanguage('zh')
    } else {
      setSelectedKeys([currentLanguage])
    }
  }, [currentLanguage])
  return (
    //   <Popover content={content} title="Title">
    //     <Button type="text" icon={<IoLanguage />}></Button>
    //   </Popover>
    <Dropdown
      menu={{ items, selectable: true, selectedKeys, onClick: changeLanguage }}
      placement="bottomCenter"
    >
      <Button type="text" icon={<IoLanguage size={20} />}></Button>
    </Dropdown>
  )
}
export default LanguageSwitcher
