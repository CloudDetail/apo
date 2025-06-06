/**
 * Copyright 2024 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

/* eslint-disable react/prop-types */
import React, { useEffect, useMemo, useState } from 'react'
import { NestedTd } from './NestedTd'

export const commonStyle = (cell) => {
  return {
    width: cell.column.customWidth,
    flex: cell.column.customWidth ? 'none' : '1',
    justifyContent: cell.column.justifyContent ? cell.column.justifyContent : 'center',
    minWidth: cell.column.minWidth,
    padding: cell?.padding ?? 8,
  }
}

const Td = React.memo(
  function Td({ cell, value, originalRow, selectedValues, updateSelectedValue, style }) {
    return (
      <td
        {...cell.getCellProps({
          style: style ? { ...commonStyle(cell), ...style } : commonStyle(cell),
        })}
      >
        {cell.column.dependsOn
          ? React.createElement(cell.column.Cell, {
              dependsValue: value,
              originalRow,
              updateSelectedValue,
              cell,
            })
          : cell.render('Cell', { selectedValues, updateSelectedValue })}
      </td>
    )
  },
  (prevProps, nextProps) => {
    return (
      prevProps.cell === nextProps.cell &&
      prevProps.value === nextProps.value &&
      prevProps.originalRow === nextProps.originalRow &&
      prevProps.selectedValues === nextProps.selectedValues &&
      prevProps.updateSelectedValue === nextProps.updateSelectedValue
    )
  },
)

function TableRow({ row, clickRow, tdPadding }) {
  const [selectedValues, setSelectedValues] = useState({})
  const updateSelectedValue = (key, value) => {
    setSelectedValues((prev) => ({ ...prev, [key]: value }))
  }
  useEffect(() => {
    return () => {
      setSelectedValues({})
    }
  }, [])

  const cellPropsArray = useMemo(() => {
    return row?.cells?.map((cell) => {
      const cellValue = cell.column.dependsOn ? selectedValues[cell.column.dependsOn] : undefined
      cell.padding = tdPadding
      return {
        cell: cell,
        originalRow: row.original,
        selectedValues: selectedValues,
        value: cellValue ?? cell.value,
        updateSelectedValue: (key, value) => {
          updateSelectedValue(key, value)
        },
        isNested: cell.column.isNested,
        style: cell.column.style,
      }
    })
  }, [row.cells, selectedValues])

  return (
    <tr
      {...row.getRowProps()}
      onClick={() => clickRow?.(row.original)}
      className={clickRow ? 'cursor-pointer' : ''}
    >
      {cellPropsArray.map((props, idx) => {
        return props.isNested ? (
          <NestedTd {...props} key={`${row.id}_body_td_${idx}`} />
        ) : (
          <Td {...props} key={`${row.id}_body_td_${idx}`} />
        )
      })}
    </tr>
  )
}

export default React.memo(TableRow, (prevProps, nextProps) => {
  return prevProps.row === nextProps.row
})
