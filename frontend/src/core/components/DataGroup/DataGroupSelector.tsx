/**
 * Copyright 2024 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */
import React, { useEffect, useMemo, useState } from 'react'
import { useDataGroupContext } from 'src/core/contexts/DataGroupContext'
import { useDispatch } from 'react-redux'
import { useSelector } from 'react-redux'
import { useNavigate, useLocation } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import DataGroupTreeSelector from './DataGroupTreeSelector'

const DataGroupSelector = ({ readonly = false }) => {
  const { t } = useTranslation('core/dataGroup')
  const dataGroup = useDataGroupContext((ctx) => ctx.dataGroup)
  const dispatch = useDispatch()
  const { dataGroupId } = useSelector((state: any) => state.dataGroupReducer)
  const navigate = useNavigate()
  const location = useLocation()
  const [treeData, setTreeData] = useState(dataGroup)
  // 移除未使用的 searchParams 变量
  useEffect(() => {
    setTreeData([
      {
        groupName: t('dataGroup'),
        title: t('dataGroup'),
        options: dataGroup,
      },
    ])
  }, [dataGroup])
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

  // 只以 URL 为真源，初始化时如果 URL 没有 groupId 且 redux 有，则补充到 URL
  useEffect(() => {
    if (flattenedAvailableNodes.length > 0) {
      let urlParams
      const URLSearchParamsCtor =
        typeof globalThis !== 'undefined' ? globalThis['URLSearchParams'] : undefined
      if (URLSearchParamsCtor) {
        urlParams = new URLSearchParamsCtor(location.search)
      } else {
        // SSR fallback，空实现
        urlParams = { get: () => null, set: () => {}, toString: () => '' }
      }
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
        if (isCurrentValid && !groupIdStr) {
          urlParams.set('groupId', dataGroupId.toString())
          navigate(`${location.pathname}?${urlParams.toString()}`, { replace: true })
        } else if (!isCurrentValid) {
          const firstAvailableNode = flattenedAvailableNodes[0]
          if (firstAvailableNode) {
            const newGroupId = Number(firstAvailableNode.groupId)
            dispatch({
              type: 'setSelectedDataGroupId',
              payload: newGroupId,
            })
            urlParams.set('groupId', newGroupId.toString())
            navigate(`${location.pathname}?${urlParams.toString()}`, { replace: true })
          }
        }
      }
    } else if (dataGroupId) {
      dispatch({ type: 'setSelectedDataGroupId', payload: null })
    }
  }, [flattenedAvailableNodes, availableNodeIds, dispatch, location, dataGroupId, navigate])

  // 删除监听 redux 变化同步 URL 的 useEffect

  const onChange = (newValue: number) => {
    dispatch({ type: 'setSelectedDataGroupId', payload: newValue })
    let urlParams
    const URLSearchParamsCtor =
      typeof globalThis !== 'undefined' ? globalThis['URLSearchParams'] : undefined
    if (URLSearchParamsCtor) {
      urlParams = new URLSearchParamsCtor(location.search)
    } else {
      urlParams = { set: () => {}, toString: () => '' }
    }
    urlParams.set('groupId', newValue.toString())
    navigate(`${location.pathname}?${urlParams.toString()}`, { replace: true })
  }

  return (
    // <Select<number>
    //   disabled={readonly}
    //   showSearch
    //   className="mx-2"
    //   style={{ width: 200 }}
    //   value={dataGroupId}
    //   placeholder="Please select"
    //   allowClear
    //   onChange={onChange}
    //   options={treeData}
    //   suffixIcon={<span className="mr-3">{t('dataGroup')}</span>}
    //   fieldNames={{
    //     label: 'groupName',
    //     value: 'groupId',
    //   }}
    // />
    <div className="w-[200px]">
      <DataGroupTreeSelector
        onChange={onChange}
        groupId={dataGroupId}
        disabled={readonly}
        suffixIcon={<span className="mr-3">{t('dataGroup')}</span>}
      />
    </div>
  )
}

export default DataGroupSelector
