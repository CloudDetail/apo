/**
 * Copyright 2024 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */
import { TreeSelect } from 'antd'
import React, { useEffect, useMemo } from 'react'
import { useDataGroupContext } from 'src/core/contexts/DataGroupContext'
import { useDispatch } from 'react-redux'
import { useSelector } from 'react-redux'

const DataGroupSelector = ({ readonly = false }) => {
  const dataGroup = useDataGroupContext((ctx) => ctx.dataGroup)
  const dispatch = useDispatch()
  const { dataGroupId } = useSelector((state: any) => state.dataGroupReducer)

  const flattenedAvailableNodes = useMemo(() => {
    const result: any[] = []

    const flattenTree = (nodes: any[]) => {
      if (!nodes || nodes.length === 0) return

      for (const node of nodes) {
        if (!node.disabled) {
          result.push(node)
        }
        if (node.subGroups && node.subGroups.length > 0) {
          flattenTree(node.subGroups)
        }
      }
    }

    flattenTree(dataGroup)
    return result
  }, [dataGroup])

  const availableNodeIds = useMemo(() => {
    return new Set(flattenedAvailableNodes.map((node) => node.groupId.toString()))
  }, [flattenedAvailableNodes])

  useEffect(() => {
    if (flattenedAvailableNodes.length > 0) {
      const isCurrentSelectionValid = dataGroupId && availableNodeIds.has(dataGroupId)

      if (!isCurrentSelectionValid) {
        const firstAvailableNode = flattenedAvailableNodes[0]
        if (firstAvailableNode) {
          dispatch({
            type: 'setSelectedDataGroupId',
            payload: firstAvailableNode.groupId,
          })
        }
      }
    } else if (dataGroupId) {
      dispatch({ type: 'setSelectedDataGroupId', payload: '' })
    }
  }, [flattenedAvailableNodes, availableNodeIds, dispatch])

  const onChange = (newValue: string) => {
    dispatch({ type: 'setSelectedDataGroupId', payload: newValue })
  }

  return (
    <TreeSelect
      disabled={readonly}
      showSearch
      className="mx-2"
      style={{ width: 200 }}
      value={dataGroupId}
      placeholder="Please select"
      allowClear
      treeDefaultExpandAll
      onChange={onChange}
      treeData={dataGroup}
      fieldNames={{
        label: 'groupName',
        value: 'groupId',
        children: 'subGroups',
      }}
    />
  )
}

export default DataGroupSelector
