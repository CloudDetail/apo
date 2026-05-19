import type React from 'react'

export type TableRecord = Record<string, React.ReactNode> & { _rowKey?: React.Key }

export const buildNodeTableDataSource = (
  rows: Array<Record<string, React.ReactNode>> = [],
): TableRecord[] =>
  rows.map((item, index) => ({
    ...item,
    _rowKey: (item?.key as React.Key | undefined) ?? index,
  }))
