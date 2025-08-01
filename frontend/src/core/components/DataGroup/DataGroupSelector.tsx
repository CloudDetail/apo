/**
 * Copyright 2024 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */
import React, { useEffect } from 'react'
import { useDispatch } from 'react-redux'
import { useSelector } from 'react-redux'
import { useNavigate, useLocation } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import DataGroupTreeSelector from './DataGroupTreeSelector'
import { useDataGroupContext } from 'src/core/contexts/DataGroupContext'

const DataGroupSelector = ({ readonly = false, hidden = false }) => {
  const { t } = useTranslation('core/dataGroup')
  const flattenedAvailableNodes = useDataGroupContext((ctx) => ctx.flattenedAvailableNodes)
  const availableNodeIds = useDataGroupContext((ctx) => ctx.availableNodeIds)
  const getDataGroup = useDataGroupContext((ctx) => ctx.getDataGroup)
  const dispatch = useDispatch()
  const { dataGroupId } = useSelector((state: any) => state.dataGroupReducer)
  const navigate = useNavigate()
  const location = useLocation()

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

  useEffect(() => {
    getDataGroup()
  }, [])
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

  return import.meta.env.VITE_APP_CODE_VERSION !== 'CE' && !hidden ? (
    <div className="w-[200px]">
      <DataGroupTreeSelector
        onChange={onChange}
        groupId={dataGroupId}
        disabled={readonly}
        suffixIcon={<span className="mr-3">{t('dataGroup')}</span>}
      />
    </div>
  ) : (
    <></>
  )
}

export default DataGroupSelector
