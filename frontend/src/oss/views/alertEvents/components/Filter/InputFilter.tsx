/**
 * Copyright 2025 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */
import { Input } from 'antd'
import { useState, useEffect } from 'react'
import { FilterRenderProps } from './type'

const InputFilter = ({ item, addFilter, filters }: FilterRenderProps) => {
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
        placeholder="请输入,按下回车检索"
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
