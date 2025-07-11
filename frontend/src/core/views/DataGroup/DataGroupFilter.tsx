/**
 * Copyright 2025 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import { Flex, Input, Select } from 'antd'
import React, { useState } from 'react'
import { useTranslation } from 'react-i18next'

interface DataGroupFilterProps {
  onSearch?: (searchValue: string) => void
}

const DataGroupFilter: React.FC<DataGroupFilterProps> = ({ onSearch }) => {
  const { t } = useTranslation('core/dataGroup')
  const [searchGroupName, setSearchGroupName] = useState<string>('')

  const handleSearchChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const value = e.target.value
    setSearchGroupName(value)
    onSearch?.(value)
  }

  return (
    <Flex align="center">
      <Flex className="flex-1 w-full mr-4" align="center">
        <p className="flex-shrink-0">{t('groupNameLabel')}</p>
        <Input value={searchGroupName} onChange={handleSearchChange} />
      </Flex>
      <Flex align="center">
        <p className="flex-shrink-0">{t('datasourceLabel')}</p>
        <Select className="w-[200px]" />
      </Flex>
    </Flex>
  )
}

export default DataGroupFilter
