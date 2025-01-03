/**
 * Copyright 2024 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

// @ts-nocheck
import React, { useEffect, useState } from 'react'
import { commonStyle } from './tableRow'
import _ from 'lodash'
import LoadingSpinner from '../Spinner'

const Td = React.memo(function Td(props) {
  const { column, value, onExpand, canExpand, isExpanded } = props
  return (
    <td
      style={{
        width: column.customWidth,
        flex: column.customWidth ? 'none' : '1',
        justifyContent: column.justifyContent ? column.justifyContent : 'center',
        minWidth: column.minWidth,
      }}
    >
      <>{column.Cell ? React.createElement(column.Cell, props) : value}</>
    </td>
  )
})

export const NestedTrs = React.memo(function NestedTrs(props) {
  const { cell, trs } = props
  const [isExpanded, setIsExpanded] = useState(false)
  const [children, setChildren] = useState([])
  useEffect(() => {
    setChildren(trs.value)
  }, [trs])
  const onExpand = (props) => {
    setIsExpanded(!isExpanded)
  }
  return (
    <>
      <tr
        style={{ width: '100%', flexGrow: 1, borderBottom: 0 }}
        onClick={() => cell.column.clickCell(props)}
        className={cell.column.clickCell ? 'cursor-pointer' : ''}
      >
        {cell.column.children?.map((column, tdIndex) => {
          const valueKey =
            typeof column.accessor === 'function' ? column.accessor(0) : column.accessor
          const value = _.get(trs, valueKey)
          return (
            <Td
              key={`body_nested_tr_td_${tdIndex}`}
              column={column}
              cell={cell}
              value={value}
              isExpanded={isExpanded}
              onExpand={onExpand}
              trs={trs}
            />
          )
        })}
      </tr>
      {/* {cell.column.showMore(props) && <Button onClick={() => cell.column.clickMore(props)}>更多</Button>} */}
      {/* {children.length > 1 &&
        isExpanded &&
        children?.map((child, childIndex) => {
          return (
            childIndex > 0 && (
              <tr
                key={`body_nested_tr_expand_${childIndex}`}
                style={{ width: "100%", flexGrow: 1, flexWrap: "wrap", overflow: "visible" }}
              >
                {cell.column.children.map((column, tdIndex) => {
                  const valueKey =
                    typeof column.accessor === "function"
                      ? column.accessor(childIndex)
                      : column.accessor;
                  const value = _.get(trs, valueKey);
                  return (
                    <Td
                      key={`body_nested_tr_expand_${childIndex}_td_${tdIndex}`}
                      cell={cell}
                      column={column}
                      value={value ?? ""}
                      onExpand={onExpand}
                      canExpand={false}
                      trs={trs}
                    />
                  );
                })}
                {childIndex === 4 && trs.rootCauseCount > 5 && (
                  <div style={{ width: "100%", textAlign: "center" }}>
                    {" "}
                    <Button onClick={() => cell.column.clickMore(props)}>更多</Button>{" "}
                  </div>
                )}
              </tr>
            )
          );
        })} */}
      {/* {children?.length > 0 && <tr style={{ width: '100%' ,padding: 10}}><Pagination numberOfPages={pagination.totalPage} currentPage={pagination.currentPage} onNavigate={onNavigate} /></tr>} */}
    </>
  )
})

export const NestedTd = React.memo(function NestedTd(props) {
  const { cell, value } = props
  const [trsData, setTrsData] = useState([])
  const [loading, setLoading] = useState(false)

  useEffect(() => {
    if (cell.column.api) {
      setTrsData([])
      setLoading(true)
      cell.column.api(props.originalRow).then(({ data, error }) => {
        if (error) {
        } else {
          setTrsData(data)
          // 这里可以处理接收到的数据
        }
        setLoading(false)
      })
    }
  }, [cell, value])

  return (
    <td
      style={{
        ...commonStyle(cell),
        display: 'flex',
        flexDirection: 'column',
        borderLeft: '1px solid var(--cui-border-color)',
        borderRight: '1px solid var(--cui-border-color)',
        // flexDirection: 'column', padding: 0, display: 'inline-flex', boxSizing: 'content-box', flexWrap: 'nowrap'
      }}
      className="nested-td relative"
    >
      {value?.map((trs, rowIndex) => {
        return <NestedTrs key={`body_nested_tr_${rowIndex}`} cell={cell} trs={trs} />
      })}
      {trsData?.map((trs, rowIndex) => {
        return <NestedTrs key={`body_nested_tr_${rowIndex}`} cell={cell} trs={trs} />
      })}
      <LoadingSpinner loading={loading} />
      {cell.column.showMore?.(cell.row.original)}
    </td>
  )
})
