/**
 * Copyright 2025 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */
import { Button, Popconfirm, Tree } from 'antd'
import Search from 'antd/es/input/Search'
import React, { useCallback, useEffect, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { LuShieldCheck } from 'react-icons/lu'
import { MdModeEdit, MdOutlineAdd } from 'react-icons/md'
import { RiDeleteBin5Line } from 'react-icons/ri'
import styles from './index.module.scss'

interface DataGroupInfo {
  groupId: number
  groupName: string
  description: string
  permissionType: 'known' | 'view' | 'edit'

  subGroups?: DataGroupInfo[]
}

interface DataGroupTreeProps {
  dataGroups: DataGroupInfo[]
  setParentGroupInfo: (data: DataGroupInfo) => void
  openAddModal: (groupId: number) => void
  openEditModal: (record: DataGroupInfo) => void
  openPermissionModal: (record: DataGroupInfo) => void
  deleteDataGroup: (record: DataGroupInfo) => void
}

const DataGroupTree: React.FC<DataGroupTreeProps> = ({
  dataGroups,
  setParentGroupInfo,
  openAddModal,
  openEditModal,
  openPermissionModal,
  deleteDataGroup,
}) => {
  const [treeData, setTreeData] = useState<DataGroupInfo[]>([])
  const [selectedKeys, setSelectedKeys] = useState<React.Key[]>([])
  const [expandedKeys, setExpandedKeys] = useState<React.Key[]>([])
  const [searchValue, setSearchValue] = useState<string>('')
  const [flattenedData, setFlattenedData] = useState<
    Array<{
      node: DataGroupInfo
      path: number[]
      level: number
    }>
  >([])
  const { t } = useTranslation('core/dataGroup')
  const { t: ct } = useTranslation('common')

  const onSelectTree = useCallback(
    (selectedKeys: React.Key[], { selectedNodes }: any) => {
      if (selectedNodes.length > 0) {
        setSelectedKeys(selectedKeys)
        setParentGroupInfo(selectedNodes[0])
      }
    },
    [setParentGroupInfo],
  )

  // Flatten tree data for efficient search
  const flattenTreeData = useCallback(
    (nodes: DataGroupInfo[], path: number[] = [], level: number = 0) => {
      const flattened: Array<{
        node: DataGroupInfo
        path: number[]
        level: number
      }> = []

      for (const node of nodes) {
        const currentPath = [...path, node.groupId]
        flattened.push({
          node,
          path: currentPath,
          level,
        })

        if (node.subGroups && node.subGroups.length > 0) {
          flattened.push(...flattenTreeData(node.subGroups, currentPath, level + 1))
        }
      }

      return flattened
    },
    [],
  )

  // Get all expandable keys from flattened data
  const getAllExpandableKeys = useCallback((flattened: typeof flattenedData): React.Key[] => {
    const keys: React.Key[] = []
    for (const item of flattened) {
      if (item.node.subGroups && item.node.subGroups.length > 0) {
        keys.push(item.node.groupId)
      }
    }
    return keys
  }, [])

  // Optimized search using flattened data
  const getSearchResult = useCallback((searchValue: string, flattened: typeof flattenedData) => {
    const expandedKeys: React.Key[] = []
    const matchedKeys: React.Key[] = []

    for (const item of flattened) {
      // Check if current node matches
      if (item.node.groupName.toLowerCase().includes(searchValue.toLowerCase())) {
        matchedKeys.push(item.node.groupId)
        // Add all parent keys to expanded keys (excluding the current node)
        expandedKeys.push(...item.path.slice(0, -1).map((key) => key as React.Key))
      }
    }

    // Remove duplicates
    return {
      expandedKeys: [...new Set(expandedKeys)],
      matchedKeys: [...new Set(matchedKeys)],
    }
  }, [])

  const onChange = useCallback(
    (e: React.ChangeEvent<HTMLInputElement>) => {
      const value = e.target.value
      setSearchValue(value)

      if (value) {
        // Use flattened data for search
        const { expandedKeys: searchExpandedKeys, matchedKeys } = getSearchResult(
          value,
          flattenedData,
        )
        setExpandedKeys(searchExpandedKeys)

        // Highlight first match if any
        if (matchedKeys.length > 0) {
          setSelectedKeys([matchedKeys[0]])
          // Find the first matched node from flattened data
          const matchedItem = flattenedData.find((item) => item.node.groupId === matchedKeys[0])
          if (matchedItem) {
            setParentGroupInfo(matchedItem.node)
          }
        }
      } else {
        // Clear search - expand all nodes
        const allKeys = getAllExpandableKeys(flattenedData)
        setExpandedKeys(allKeys)
      }
    },
    [flattenedData, getSearchResult, getAllExpandableKeys, setParentGroupInfo],
  )

  // Initialize tree data and flatten it
  useEffect(() => {
    setTreeData(dataGroups)
    const flattened = flattenTreeData(dataGroups)
    setFlattenedData(flattened)
  }, [dataGroups, flattenTreeData])

  // Set initial selection and expand all nodes
  useEffect(() => {
    if (treeData && treeData.length > 0 && !searchValue) {
      const allKeys = getAllExpandableKeys(flattenedData)
      setExpandedKeys(allKeys)

      const currentSelectedKey = selectedKeys[0]
      const allGroupIds = flattenedData.map((item) => item.node.groupId)

      if (selectedKeys.length === 0 || !allGroupIds.includes(currentSelectedKey as number)) {
        const firstAvailableNode = flattenedData.find(
          (item) => item.node.permissionType !== 'known',
        )?.node
        if (firstAvailableNode) {
          setSelectedKeys([firstAvailableNode.groupId])
          setParentGroupInfo(firstAvailableNode)
        }
      }
    }
  }, [
    treeData,
    selectedKeys.length,
    getAllExpandableKeys,
    setParentGroupInfo,
    searchValue,
    flattenedData,
  ])

  const titleRender = useCallback(
    (nodeData: DataGroupInfo) => {
      const strTitle = nodeData.groupName
      const index = strTitle.toLowerCase().indexOf(searchValue.toLowerCase())

      const renderTitle = () => {
        if (searchValue && index > -1) {
          const beforeStr = strTitle.substring(0, index)
          const afterStr = strTitle.slice(index + searchValue.length)
          const matchedStr = strTitle.substring(index, index + searchValue.length)

          return (
            <span>
              {beforeStr}
              <span className="text-[var(--ant-color-warning)] font-semibold">{matchedStr}</span>
              {afterStr}
            </span>
          )
        }
        return <span>{strTitle}</span>
      }

      return (
        <div className={styles.treeTitleRow}>
          <div className="flex-1 truncate">{renderTitle()}</div>
          <div className={styles.treeTitleActions}>
            {nodeData.permissionType === 'edit' && (
              <Button
                type="link"
                size="small"
                icon={<MdModeEdit />}
                onClick={(e: React.MouseEvent) => {
                  e.stopPropagation()
                  openEditModal(nodeData)
                }}
              />
            )}
            {nodeData.permissionType === 'edit' && (
              <>
                <Popconfirm
                  title={t('confirmDelete', {
                    groupName: nodeData.groupName,
                  })}
                  onConfirm={(e: React.MouseEvent) => {
                    e.stopPropagation()
                    deleteDataGroup(nodeData)
                  }}
                  onPopupClick={(e: React.MouseEvent) => {
                    e.stopPropagation()
                  }}
                  okText={ct('confirm')}
                  cancelText={ct('cancel')}
                >
                  <Button
                    type="text"
                    icon={<RiDeleteBin5Line />}
                    danger
                    size="small"
                    onClick={(e: React.MouseEvent) => e.stopPropagation()}
                  />
                </Popconfirm>
                <Button
                  type="text"
                  size="small"
                  icon={<LuShieldCheck />}
                  onClick={(e: React.MouseEvent) => {
                    e.stopPropagation()
                    e.preventDefault()
                    openPermissionModal(nodeData)
                  }}
                />
              </>
            )}
            {nodeData.permissionType !== 'known' && (
              <Button
                type="link"
                size="small"
                icon={<MdOutlineAdd />}
                onClick={(e: React.MouseEvent) => {
                  e.stopPropagation()
                  openAddModal(nodeData.groupId)
                }}
              />
            )}
          </div>
        </div>
      )
    },
    [t, ct, openEditModal, deleteDataGroup, openPermissionModal, openAddModal, searchValue],
  )

  return (
    <div className="h-full flex flex-col overflow-hidden ">
      <Search
        style={{ marginBottom: 8 }}
        className="h-[42px] grow-0 shrink-0"
        placeholder="Search"
        onChange={onChange}
      />
      {treeData && treeData.length > 0 && (
        <Tree
          selectedKeys={selectedKeys}
          expandedKeys={expandedKeys}
          onExpand={setExpandedKeys}
          onSelect={onSelectTree}
          treeData={treeData}
          titleRender={titleRender}
          className="pr-3 h-full overflow-auto flex-1"
          fieldNames={{
            title: 'groupName',
            key: 'groupId',
            children: 'subGroups',
          }}
          blockNode
          // autoExpandParent={autoExpandParent}
        />
      )}
    </div>
  )
}

export default DataGroupTree
