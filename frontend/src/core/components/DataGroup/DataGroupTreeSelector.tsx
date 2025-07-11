/**
 * Copyright 2025 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */
import { TreeSelect } from 'antd'
import React, { useEffect, useState } from 'react'
import { getDatasourceByGroupApiV2 } from 'src/core/api/dataGroup'
import { useTranslation } from 'react-i18next'

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
  const [dataGroup, setDataGroups] = useState<DataGroupInfo[]>([])
  const [selectedGroupId, setSelectedGroupId] = useState<number>(groupId)
  const { t } = useTranslation('core/dataGroup')
  // 递归处理树结构，设置 disabled 字段，并收集所有 groupId
  const processTree = (nodes: DataGroupInfo[]): { tree: DataGroupInfo[]; keys: number[] } => {
    let keys: number[] = []
    const process = (list: DataGroupInfo[]): DataGroupInfo[] => {
      return list.map((node) => {
        keys.push(Number(node.groupId))
        const newNode: DataGroupInfo = {
          ...node,
          disabled: node.permissionType === 'known',
        }
        if (node.subGroups && node.subGroups.length > 0) {
          newNode.subGroups = process(node.subGroups)
        }
        return newNode
      })
    }
    const tree = process(nodes)
    return { tree, keys }
  }

  const getDataGroups = () => {
    getDatasourceByGroupApiV2().then((res: any) => {
      // 兼容 res 可能为单个对象或数组
      const rawList = Array.isArray(res) ? res : [res]
      const { tree, keys } = processTree(rawList)
      setDataGroups(tree)
      setExpandedKeys(keys)
    })
  }

  useEffect(() => {
    getDataGroups()
  }, [])

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
