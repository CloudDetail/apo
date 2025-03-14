/**
 * Copyright 2025 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import { Select } from 'antd'
import React from 'react'

const FilterSelector = ({ label, placeholder, value, onChange, options, id }) => (
  <div className="flex flex-row items-center mr-5 text-sm min-w-[280px]">
    <span className="text-nowrap">{label}：</span>
    <Select
      mode="multiple"
      allowClear
      className="w-full"
      id={id}
      placeholder={placeholder}
      value={value}
      onChange={onChange}
      options={options}
      popupMatchSelectWidth={false}
      maxTagCount={2}
      maxTagPlaceholder={(omittedValues) => `+${omittedValues.length}...`}
    />
  </div>
)

export  default React.memo(FilterSelector)
