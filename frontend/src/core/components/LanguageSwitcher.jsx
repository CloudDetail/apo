/**
 * Copyright 2024 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import { Select } from 'antd'
import i18next from 'i18next'
import { useEffect, useState } from 'react'
import { useDispatch, useSelector } from 'react-redux'
import { showToast } from '../utils/toast'
import { useTranslation } from 'react-i18next'

const LanguageSwitcher = () => {
  const { t } = useTranslation('core/systemConfiguration')
  const [selectedKeys, setSelectedKeys] = useState('')
  const dispatch = useDispatch()
  const currentLang = i18next.language
  const changeLanguage = (value) => {
    i18next
      .changeLanguage(value)
      .then(() => {
        dispatch({ type: 'setLanguage', payload: value }) // 更新 Redux 中的语言状态
      })
      .then(() => {
        showToast({
          title: t('languageSuccess'),
          color: 'success',
        })
        location.reload()
      })
  }

  const options = [
    { value: 'en', label: 'English' },
    { value: 'zh', label: '简体中文' },
  ]

  useEffect(() => {
    setSelectedKeys(currentLang)
  }, [currentLang])

  return (
    <Select
      value={selectedKeys}
      onChange={changeLanguage}
      options={options}
      className="w-1/2 rounded-none"
    />
  )
}
export default LanguageSwitcher
