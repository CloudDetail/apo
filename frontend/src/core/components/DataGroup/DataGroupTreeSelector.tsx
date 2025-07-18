/**
 * Copyright 2025 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */
import { TreeSelect } from 'antd'
import React, { useEffect, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { useDataGroupContext } from 'src/core/contexts/DataGroupContext'

interface DataGroupInfo {
  groupId: number
  groupName: string
  description: string
  permissionType: 'known' | 'view' | 'edit'
  subGroups?: DataGroupInfo[]
  disabled?: boolean // 新增字段
}

interface DataGroupTreeProps {
  onChange: (data: DataGroupInfo) => void
  id?: string
  groupId?: number
  disabled?: boolean
  suffixIcon?: React.ReactNode
}

const DataGroupTreeSelector: React.FC<DataGroupTreeProps> = ({
  onChange,
  id,
  groupId,
  disabled = false,
  suffixIcon,
}) => {
  const [expandedKeys, setExpandedKeys] = useState<number[]>([])
  const [selectedGroupId, setSelectedGroupId] = useState<number>(groupId)
  const { t } = useTranslation('core/dataGroup')
  const dataGroup = useDataGroupContext((ctx) => ctx.dataGroup)
  const allNodeIds = useDataGroupContext((ctx) => ctx.allNodeIds)
  useEffect(() => {
    setExpandedKeys(allNodeIds)
  }, [dataGroup, allNodeIds])
  useEffect(() => {
    setSelectedGroupId(groupId)
  }, [groupId])

  return (
    <div id={id} className="w-full">
      {dataGroup && dataGroup.length > 0 && (
        <TreeSelect
          disabled={disabled}
          placeholder={t('dataGroupPlaceholder')}
          value={selectedGroupId}
          showSearch
          treeData={dataGroup}
          treeExpandedKeys={expandedKeys}
          className="h-full w-full"
          fieldNames={{
            label: 'groupName',
            value: 'groupId',
            children: 'subGroups',
            _title: ['groupName'],
          }}
          onTreeExpand={(keys) => {
            setExpandedKeys(keys)
          }}
          treeNodeFilterProp="groupName"
          onChange={onChange}
          onSelect={(value) => {
            setSelectedGroupId(Number(value))
          }}
          suffixIcon={suffixIcon}
          popupMatchSelectWidth={false}
          popupClassName="max-w-[700px]"
        />
      )}
    </div>
  )
}

export default DataGroupTreeSelector
