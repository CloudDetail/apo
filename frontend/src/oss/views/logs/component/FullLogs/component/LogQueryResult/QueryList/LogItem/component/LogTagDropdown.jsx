/**
 * Copyright 2024 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import { Dropdown } from 'antd'
import React, { useMemo } from 'react'
import { copyValue } from 'src/core/components/CopyButton'
import { useLogsContext } from 'src/core/contexts/LogsContext'
import { useTranslation } from 'react-i18next' // 引入i18n

// 判断字段值是否在当前查询中的工具函数（支持精确匹配和LIKE模糊匹配）
const isFieldInQuery = (query, objKey, value) => {
  if (!query || !objKey || value === undefined || value === null) return false
  
  // 精确匹配检查
  let exactQueryPart
  switch (typeof value) {
    case 'string':
      exactQueryPart = '`' + objKey + '` = ' + "'" + value + "'"
      break
    case 'number':
      exactQueryPart = '`' + objKey + '` = ' + value
      break
    case 'boolean':
      exactQueryPart = '`' + objKey + '` = ' + value
      break
    default:
      exactQueryPart = '`' + objKey + '` = -' + "'" + value + "'"
      break
  }
  
  // 检查精确匹配
  if (query.includes(exactQueryPart)) {
    return true
  }
  
  // 检查LIKE模糊匹配（仅对字符串类型）
  if (typeof value === 'string') {
    // 匹配模式：`field` LIKE '%value%'
    const escapedObjKey = objKey.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')
    const likePattern = new RegExp('`' + escapedObjKey + '`\\s+LIKE\\s+\'%([^%]*)%\'', 'gi')
    let match
    
    while ((match = likePattern.exec(query)) !== null) {
      const likeValue = match[1] // 提取LIKE中的值
      // 检查字段值是否包含查询中的部分内容，或查询中的部分内容包含字段值
      if (likeValue && (value.includes(likeValue) || likeValue.includes(value))) {
        return true
      }
    }
  }
  
  return false
}

const LogTagDropDown = ({ objKey, value, children, trigger = ['click', 'contextMenu'] }) => {
  const { t } = useTranslation('oss/fullLogs')
  const { query, updateQuery } = useLogsContext()
  
  // 计算是否高亮
  const isHighlighted = useMemo(() => {
    return isFieldInQuery(query, objKey, value)
  }, [query, objKey, value])
  const addToQuery = () => {
    let newQueryPart
    switch (typeof value) {
      case 'string':
        newQueryPart = '`' + objKey + '` = ' + "'" + value + "'"
        break
      case 'number':
        newQueryPart = '`' + objKey + '` = ' + value
        break
      case 'boolean':
        newQueryPart = '`' + objKey + '` = ' + value
        break
      default:
        newQueryPart = '`' + objKey + '` = -' + "'" + value + "'"
        break
    }
    // 检查 query 是否已经包含 newQueryPart
    if (!query.includes(newQueryPart)) {
      let newQuery = query
      if (newQuery.length > 0) {
        newQuery += ' AND '
      }
      newQuery += newQueryPart
      updateQuery(newQuery) // 更新查询
    }
  }
  const clickCopy = () => {
    copyValue(value)
  }
  const items = [
    {
      key: 'filter',
      label: (
        <div onClick={addToQuery} className="w-full">
          {t('logTagDropdown.addToQueryText')}
        </div>
      ),
    },
    {
      key: 'copy',
      label: (
        <div onClick={clickCopy} className="w-full">
          {t('logTagDropdown.copyValueText')}
        </div>
      ),
    },
  ]
  return (
    <>
      <Dropdown
        menu={{ items }}
        trigger={trigger}
        overlayStyle={{ minWidth: 'auto' }}
        getPopupContainer={(triggerNode) => triggerNode.parentNode || document.body}
      >
        <span>{React.isValidElement(children) ? React.cloneElement(children, { isHighlighted }) : children}</span>
      </Dropdown>
    </>
  )
}
export default LogTagDropDown
