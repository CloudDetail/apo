/**
 * Copyright 2025 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import { Flex, Input, Select } from 'antd'
import { useState } from 'react'

const DataGroupFilter = ({ onSearch }) => {
  const [searchGroupName, setSearchGroupName] = useState('')

  return (
    <Flex align="center ">
      <Flex className="flex-1 w-full mr-4" align="center">
        <p className="flex-shrink-0">数据组名：</p>
        <Input value={searchGroupName} onChange={(e) => setSearchGroupName(e.target.value)}></Input>
      </Flex>
      <Flex align="center">
        <p className="flex-shrink-0">数据源：</p>
        <Select className="w-[200px]" />
      </Flex>
    </Flex>
  )
}
export default DataGroupFilter
