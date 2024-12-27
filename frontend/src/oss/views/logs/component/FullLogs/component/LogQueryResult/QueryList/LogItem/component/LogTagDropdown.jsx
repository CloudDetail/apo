/**
 * Copyright 2024 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import { Dropdown } from 'antd'
import React from 'react'
import { copyValue } from 'src/core/components/CopyButton'
import { useLogsContext } from 'src/core/contexts/LogsContext'
import { useTranslation } from 'react-i18next' // 引入i18n

const LogTagDropDown = ({ objKey, value, children, trigger = ['click', 'contextMenu'] }) => {
  const { t } = useTranslation('oss/fullLogs') // 使用i18n
  const { query, updateQuery } = useLogsContext()
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
        getPopupContainer={(triggerNode) => triggerNode.parentNode}
      >
        <span>{children}</span>
      </Dropdown>
    </>
  )
}
export default LogTagDropDown
