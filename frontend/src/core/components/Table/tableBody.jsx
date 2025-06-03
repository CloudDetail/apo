/**
 * Copyright 2024 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

/* eslint-disable react/prop-types */
import React from 'react'
import TableRow from './tableRow'
import Empty from '../Empty/Empty'

function TableBody(props) {
  const { page, prepareRow, rowKey, loading, clickRow, emptyContent, scrollY, tdPadding } =
    props.tableBodyProps
  const getRowKeyValue = (row) => {
    if (row) {
      return row.id
    } else if (typeof rowKey === 'function') {
      return rowKey(row.original)
    } else {
      return row.original[rowKey]
    }
  }
  return (
    <tbody
      style={{
        maxHeight: scrollY ? scrollY : 'auto',
        overflowY: scrollY ? 'scroll' : 'auto',
      }}
    >
      {(page &&
        page.length > 0 &&
        page.map((row, idx) => {
          prepareRow(row)
          return (
            <TableRow
              row={row}
              key={getRowKeyValue(row)}
              clickRow={clickRow}
              tdPadding={tdPadding}
            />
          )
        })) ||
        loading || <Empty context={emptyContent} />}
    </tbody>
  )
}

export default React.memo(TableBody)
