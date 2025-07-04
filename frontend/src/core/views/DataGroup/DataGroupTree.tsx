/**
 * Copyright 2025 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */
import { Button, Popconfirm, Tree } from 'antd'
import Search from 'antd/es/input/Search'
import { useEffect, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { LuShieldCheck } from 'react-icons/lu'
import { MdModeEdit, MdOutlineAdd } from 'react-icons/md'
import { RiDeleteBin5Line } from 'react-icons/ri'
import styles from './index.module.scss'

const DataGroupTree = ({
  dataGroups,
  setParentGroupInfo,
  openAddModal,
  openEditModal,
  openPermissionModal,
  deleteDataGroup,
}) => {
  const [treeData, setTreeData] = useState([])
  const [searchValue, setSearchValue] = useState('')
  const [selectedKeys, setSelectedKeys] = useState([])
  const [expandedKeys, setExpandedKeys] = useState([])
  const { t } = useTranslation('core/dataGroup')
  const { t: ct } = useTranslation('common')

  const onSelectTree = (selectedKeys, { selectedNodes }) => {
    if (selectedNodes.length > 0) {
      setSelectedKeys(selectedKeys)
      setParentGroupInfo(selectedNodes[0])
    }
  }

  const onChange = (e) => {
    setSearchValue(e.target.value)
    setTreeData(dataGroups.filter((item) => item.title.includes(e.target.value)))
  }

  // 递归获取所有节点的 key，同时检查 selectedKeys 是否存在
  const getAllKeysAndCheckSelected = (nodes, selectedKeys) => {
    const allKeys = []
    let selectedExists = false

    const traverse = (nodeList) => {
      for (const node of nodeList) {
        allKeys.push(node.groupId)
        if (selectedKeys.includes(node.groupId)) {
          selectedExists = true
        }
        if (node.subGroups && node.subGroups.length > 0) {
          traverse(node.subGroups)
        }
      }
    }

    traverse(nodes)
    return { allKeys, selectedExists }
  }

  useEffect(() => {
    setTreeData(dataGroups)
  }, [dataGroups])

  useEffect(() => {
    if (treeData && treeData.length > 0) {
      const { allKeys, selectedExists } = getAllKeysAndCheckSelected(treeData, selectedKeys)

      setExpandedKeys(allKeys)

      if (!selectedExists) {
        const firstNode = treeData[0]
        setSelectedKeys([firstNode.groupId])
        setParentGroupInfo(firstNode)
      }
    }
  }, [treeData, selectedKeys])

  const titleRender = (nodeData) => {
    return (
      <div className={styles.treeTitleRow}>
        <div className="flex-1 truncate">{nodeData.groupName}</div>
        <div className={styles.treeTitleActions}>
          {nodeData.permissionType === 'edit' && (
            <Button
              type="link"
              size="small"
              icon={<MdModeEdit />}
              onClick={(e) => {
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
                onConfirm={(e) => {
                  e.stopPropagation()
                  deleteDataGroup(nodeData)
                }}
                onPopupClick={(e) => {
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
                  onClick={(e) => e.stopPropagation()}
                />
              </Popconfirm>
              <Button
                type="text"
                size="small"
                icon={<LuShieldCheck />}
                onClick={(e) => {
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
              onClick={(e) => {
                e.stopPropagation()
                openAddModal(nodeData.groupId)
              }}
            />
          )}
        </div>
      </div>
    )
  }

  return (
    <>
      <Search
        style={{ marginBottom: 8 }}
        className="h-[42px]"
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
          className="pr-3 h-full"
          fieldNames={{
            title: 'groupName',
            key: 'groupId',
            children: 'subGroups',
          }}
          blockNode
        />
      )}
    </>
  )
}

export default DataGroupTree
