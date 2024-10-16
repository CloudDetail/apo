import { Dropdown } from 'antd'
import React from 'react'
import { copyValue } from 'src/components/CopyButton'
import { useLogsContext } from 'src/contexts/LogsContext'

const LogTagDropDown = ({ objKey, value, children }) => {
  const { query, updateQuery } = useLogsContext()
  const addToQuery = () => {
    let newQueryPart = '`' + objKey + '` =' + "'" + value + "'"
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
          添加到查询
        </div>
      ),
    },
    {
      key: 'copy',
      label: (
        <div onClick={clickCopy} className="w-full">
          复制值
        </div>
      ),
    },
  ]
  return (
    <>
      <Dropdown
        menu={{ items }}
        trigger={['click', 'contextMenu']}
        overlayStyle={{ minWidth: 'auto' }}
        getPopupContainer={(triggerNode) => triggerNode.parentNode}
      >
        <span>{children}</span>
      </Dropdown>
    </>
  )
}
export default LogTagDropDown
