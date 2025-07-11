/**
 * Copyright 2025 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import { Select } from 'antd'
import React from 'react'

const FilterSelector = ({
  label,
  placeholder,
  value,
  onChange,
  options,
  id,
  mode = 'multiple',
}) => (
  <div className="flex flex-row items-center mr-5 text-sm min-w-[280px]  mb-1">
    <span className="text-nowrap">{label}：</span>
    <Select
      mode={mode}
      allowClear
      className="w-full"
      id={id}
      placeholder={placeholder}
      value={mode === 'multiple' ? value : value?.[0] || null}
      onChange={(e) => {
        if (mode === 'multiple') {
          onChange(e)
        } else {
          onChange(e ? [e] : [])
        }
      }}
      options={options}
      popupMatchSelectWidth={false}
      maxTagCount={2}
      maxTagPlaceholder={(omittedValues) => `+${omittedValues.length}...`}
    />
  </div>
)

export default React.memo(FilterSelector)
