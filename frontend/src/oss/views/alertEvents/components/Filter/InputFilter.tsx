/**
 * Copyright 2025 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */
import { Input } from 'antd'
import { useState, useEffect } from 'react'
import { FilterRenderProps } from './type'
import { useTranslation } from 'react-i18next'

const InputFilter = ({ item, addFilter, filters }: FilterRenderProps) => {
  const { t } = useTranslation('oss/alertEvents')
  const [value, setValue] = useState(null)
  useEffect(() => {
    const oldValue = filters.find((filterItem) => filterItem.key === item.key)
    if (oldValue) setValue(oldValue.matchExpr)
  }, [filters])
  return (
    <>
      <div>{item.name}</div>
      <Input
        value={value}
        onChange={(e) => setValue(e.target.value)}
        placeholder={t('enterAndPress')}
        onPressEnter={() =>
          addFilter({
            key: item.key,
            matchExpr: value,
            name: item.name,
          })
        }
      />
    </>
  )
}
export default InputFilter
