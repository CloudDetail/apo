/**
 * Copyright 2024 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */
import { TreeSelect } from 'antd'
import React, { useEffect, useMemo } from 'react'
import { useDataGroupContext } from 'src/core/contexts/DataGroupContext'
import { useDispatch } from 'react-redux'
import { useSelector } from 'react-redux'
import { useNavigate, useLocation, useSearchParams } from 'react-router-dom'

const DataGroupSelector = ({ readonly = false }) => {
  const dataGroup = useDataGroupContext((ctx) => ctx.dataGroup)
  const dispatch = useDispatch()
  const { dataGroupId } = useSelector((state: any) => state.dataGroupReducer)
  const navigate = useNavigate()
  const location = useLocation()
  const [searchParams, setSearchParams] = useSearchParams()

  const flattenedAvailableNodes = useMemo(() => {
    const result: any[] = []
    const flattenTree = (nodes: any[]) => {
      if (!nodes || nodes.length === 0) return
      for (const node of nodes) {
        if (!node.disabled) {
          result.push(node)
        }
        if (node.subGroups?.length) {
          flattenTree(node.subGroups)
        }
      }
    }
    flattenTree(dataGroup)
    return result
  }, [dataGroup])

  const availableNodeIds = useMemo(() => {
    return new Set(flattenedAvailableNodes.map((node) => Number(node.groupId)))
  }, [flattenedAvailableNodes])

  // --- 初始化：从 URL 设置 groupId（仅首次）
  useEffect(() => {
    if (flattenedAvailableNodes.length > 0) {
      const urlParams = searchParams
      const groupIdStr = urlParams.get('groupId')
      const groupIdFromUrl = groupIdStr ? Number(groupIdStr) : null
      const isUrlGroupValid = groupIdFromUrl !== null && availableNodeIds.has(groupIdFromUrl)
      if (isUrlGroupValid) {
        if (groupIdFromUrl !== dataGroupId) {
          dispatch({
            type: 'setSelectedDataGroupId',
            payload: groupIdFromUrl,
          })
        }
      } else {
        const isCurrentValid = typeof dataGroupId === 'number' && availableNodeIds.has(dataGroupId)
        if (!isCurrentValid) {
          const firstAvailableNode = flattenedAvailableNodes[0]
          if (firstAvailableNode) {
            const newGroupId = Number(firstAvailableNode.groupId)
            dispatch({
              type: 'setSelectedDataGroupId',
              payload: newGroupId,
            })
            // 👇 把默认 groupId 写入 URL
            urlParams.set('groupId', newGroupId.toString())
            navigate(`${location.pathname}?${urlParams.toString()}`, { replace: true })
          }
        } else if (!groupIdStr) {
          urlParams.set('groupId', dataGroupId.toString())
          navigate(`${location.pathname}?${urlParams.toString()}`, { replace: true })
        }
      }
    } else if (dataGroupId) {
      dispatch({ type: 'setSelectedDataGroupId', payload: null })
    }
  }, [flattenedAvailableNodes, availableNodeIds, dispatch, location, searchParams])

  // --- 双向绑定：dataGroupId 变化 → 同步到 URL（只在实际变化时）
  useEffect(() => {
    if (typeof dataGroupId !== 'number') return
    const urlParams = new URLSearchParams(location.search)
    if (Number(urlParams.get('groupId')) !== dataGroupId) {
      urlParams.set('groupId', dataGroupId.toString())
      navigate(`${location.pathname}?${urlParams.toString()}`, { replace: true })
    }
  }, [dataGroupId, location.pathname, location.search, navigate])

  const onChange = (newValue: number) => {
    dispatch({ type: 'setSelectedDataGroupId', payload: newValue })
  }

  return (
    <TreeSelect<number>
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
