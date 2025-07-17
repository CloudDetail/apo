/**
 * Copyright 2025 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import { useEffect, useState } from 'react'
import { getDatasourceByGroupApiV2 } from 'src/core/api/dataGroup'
import { DataGroupItem } from 'src/core/types/dataGroup'
import { Tree } from 'antd'
import Search from 'antd/es/input/Search'

interface DataGroupPermissionProps {
  id: string
  dataGroupList: any[]
  onChange: any
  type: 'team' | 'user'
  permissionSourceTeam: any[]
  readOnly?: boolean
}

const DataGroupPermission = (props: DataGroupPermissionProps) => {
  const { id, dataGroupList = [], onChange, readOnly = false } = props
  const [checkedKeys, setCheckedKeys] = useState<number[]>([])
  const [treeData, setTreeData] = useState<DataGroupItem[]>([])
  const [expandedKeys, setExpandedKeys] = useState([])
  const [flattenedData, setFlattenedData] = useState<
    Array<{ node: any; path: number[]; level: number }>
  >([])
  const flattenTreeData = (nodes: any[], path: number[] = [], level: number = 0) => {
    const flattened: Array<{ node: any; path: number[]; level: number }> = []
    for (const node of nodes) {
      const currentPath = [...path, node.groupId]
      flattened.push({ node, path: currentPath, level })
      if (node.subGroups && node.subGroups.length > 0) {
        flattened.push(...flattenTreeData(node.subGroups, currentPath, level + 1))
      }
    }
    return flattened
  }

  const getAllExpandableKeys = (flattened: typeof flattenedData) => {
    return flattened
      .filter((item) => item.node.subGroups && item.node.subGroups.length > 0)
      .map((item) => item.node.groupId)
  }

  const getSearchResult = (searchValue: string, flattened: typeof flattenedData) => {
    const expandedKeys: string[] = []
    const matchedKeys: string[] = []
    for (const item of flattened) {
      if (item.node.groupName.toLowerCase().includes(searchValue.toLowerCase())) {
        matchedKeys.push(item.node.groupId)
        expandedKeys.push(...item.path.slice(0, -1))
      }
    }
    return { expandedKeys: [...new Set(expandedKeys)], matchedKeys: [...new Set(matchedKeys)] }
  }

  const getDataGroups = () => {
    getDatasourceByGroupApiV2().then((res) => {
      const treeData = res.data || res
      setTreeData([treeData])
      const flattened = flattenTreeData([treeData])
      setFlattenedData(flattened)
      const allKeys = getAllExpandableKeys(flattened)
      const checkedKeysFromDataGroup = flattened
        .filter((item) => dataGroupList.includes(item.node.groupId))
        .map((item) => item.node.groupId)
      setExpandedKeys(allKeys)
      setCheckedKeys(checkedKeysFromDataGroup)
    })
  }

  useEffect(() => {
    getDataGroups()
  }, [])

  useEffect(() => {
    if (flattenedData.length > 0) {
      const checkedKeysFromDataGroup = flattenedData
        .filter((item) => dataGroupList.includes(item.node.groupId))
        .map((item) => item.node.groupId)
      const allChildKeys = Array.from(
        new Set(
          checkedKeysFromDataGroup.flatMap((key) =>
            flattenedData
              .filter((item) => item.path.includes(key))
              .map((item) => item.node.groupId),
          ),
        ),
      )
      setCheckedKeys(allChildKeys)
    }
  }, [dataGroupList, flattenedData])

  const onSearch = (value: string) => {
    if (value) {
      const { expandedKeys: searchExpandedKeys } = getSearchResult(value, flattenedData)
      setExpandedKeys(searchExpandedKeys)
    } else {
      const allKeys = getAllExpandableKeys(flattenedData)
      setExpandedKeys(allKeys)
    }
  }

  const getAllChildKeys = (nodeId: string): string[] => {
    const childKeys: string[] = []
    const targetItem = flattenedData.find((item) => item.node.groupId === nodeId)
    if (targetItem) {
      flattenedData.forEach((item) => {
        if (
          targetItem.path.every((pathId, index) => item.path[index] === pathId) &&
          item.path.length > targetItem.path.length
        ) {
          childKeys.push(item.node.groupId)
        }
      })
    }
    return childKeys
  }

  const getDisabledKeys = useMemo(() => {
    const disabledSet = new Set<string>()
    checkedKeys.forEach((key) => {
      const childKeys = getAllChildKeys(key)
      childKeys.forEach((child) => disabledSet.add(child))
    })
    return disabledSet
  }, [checkedKeys, flattenedData])

  const generateTreeDataWithDisabled = (nodes: any[]): any[] => {
    return nodes.map((node) => {
      const childKeys = node.subGroups ? node.subGroups.map((n: any) => n.groupId) : []
      return {
        ...node,
        disabled: getDisabledKeys.has(node.groupId),
        subGroups: node.subGroups ? generateTreeDataWithDisabled(node.subGroups) : undefined,
      }
    })
  }

  const onCheck = (checkedKeys: any, { checked, node }) => {
    const checkedArray = Array.isArray(checkedKeys) ? checkedKeys : checkedKeys.checked
    let result = []
    // checked
    if (checked) {
      const allChildKeys = Array.from(
        new Set(
          checkedArray.flatMap((key) =>
            flattenedData
              .filter((item) => item.path.includes(key))
              .map((item) => item.node.groupId),
          ),
        ),
      )
      result = allChildKeys
      setCheckedKeys(result)
    } else {
      const nodeChildKeys = flattenedData
        .filter((item) => item.path.includes(node.groupId))
        .map((item) => item.node.groupId)
      result = checkedArray.filter((key) => !nodeChildKeys.includes(key))
      setCheckedKeys(result)
    }
    onChange(result)
  }

  return (
    <div style={{ maxHeight: '30vh' }} className="flex flex-col overflow-hidden w-full" id={id}>
      <Search className="grow-0 shrink-0" placeholder="Search" onSearch={onSearch} />
      {treeData.length > 0 && (
        <Tree
          checkable
          selectable={false}
          checkedKeys={checkedKeys}
          expandedKeys={expandedKeys}
          onExpand={setExpandedKeys}
          onCheck={onCheck}
          treeData={generateTreeDataWithDisabled(treeData)}
          style={{ width: '100%' }}
          className="pr-3 h-full w-full overflow-auto"
          fieldNames={{ title: 'groupName', key: 'groupId', children: 'subGroups' }}
          blockNode
          disabled={readOnly}
          multiple
          checkStrictly
        />
      )}
    </div>
  )
}

export default DataGroupPermission
