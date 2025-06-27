/**
 * Copyright 2024 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import { useLogsContext } from 'src/core/contexts/LogsContext'
import LogValueTag from './LogValueTag'
import LogKeyTag from './LogKeyTag'
import { useMemo } from 'react'

const LogItemFold = ({ tags }) => {
  const { tableInfo, displayFields } = useLogsContext()
  //由tableName和type组成的唯一标识
  const tableId = `${tableInfo.tableName}_${tableInfo.type}`
  // 计算过滤后的 tags
  const filteredTags = useMemo(() => {
    if (!tags) return []

    const isFieldValid = ([key, value]) =>
      displayFields[tableId]?.includes(key) &&
      value !== '' &&
      key !== (tableInfo?.timeField || 'timestamp') &&
      typeof value !== 'object'

    return Object.entries(tags).filter(isFieldValid)
  }, [tags, displayFields, tableInfo?.timeField])

  return (
    <>
      {/* 渲染 tags */}
      <div className="text-ellipsis text-wrap flex" style={{ display: '-webkit-box' }}>
        {filteredTags?.map(([key, value]) => (
          <LogValueTag key={key} objKey={key} value={String(value)} />
        ))}
      </div>
    </>
  )
}

export default LogItemFold
