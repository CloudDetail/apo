/**
 * Copyright 2024 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import React, { useEffect, useMemo, useRef, useState } from 'react'
import { useSortBy, useTable, usePagination } from 'react-table'
import './index.css'
import _ from 'lodash'
import TableBody from './tableBody'
import LoadingSpinner from '../Spinner'
import BasicPagination from './basicPagination'
import { Tooltip } from 'antd'
import { VscTriangleUp, VscTriangleDown } from 'react-icons/vsc'
const BasicTable = React.memo((props) => {
  const {
    data,
    columns,
    pagination,
    loading = false,
    onChange,
    rowKey,
    noHeader,
    showBorder,
    clickRow,
    emptyContent = null,
    showLoading = true,
    sortBy = [],
    setSortBy = null,
  } = props
  const tableRef = useRef(null)
  const getSortByColumns = (columns) => {
    let sortByColumns = []
    _.forEach(columns, (column) => {
      if (column.defaultSortBy) {
        sortByColumns.push({
          id: column.accessor,
          desc: column.defaultSortBy === 'desc' ? true : false,
        })
      }
    })
    return sortByColumns
  }
  const [paginationLoading, setPaginationLoading] = useState(false)
  const {
    getTableProps,
    getTableBodyProps,
    headerGroups,
    prepareRow,
    page,
    pageCount,
    gotoPage,
    nextPage,
    previousPage,
    setPageSize,
    state: { pageIndex, pageSize },
  } = useTable(
    {
      defaultColumn: {
        disableSortBy: true,
      },
      initialState: {
        sortBy: sortBy,
        pageIndex: pagination?.pageIndex ? pagination.pageIndex - 1 : 0,
        pageSize: pagination?.pageSize ? pagination.pageSize : 100000,
      },
      columns,
      data,
      manualSortBy: true,
      manualPagination: pagination?.pageCount !== undefined,
      ...(pagination?.pageCount !== undefined ? { pageCount: pagination.pageCount } : {}),
    },
    useSortBy,
    usePagination,
  )
  useEffect(() => {
    setPaginationLoading(true)
    if (onChange) {
      onChange({ pageSize: pageSize, pageIndex: pageIndex + 1 })
    }
    setPaginationLoading(false)
  }, [pageSize, pageIndex, onChange])

  const tableBodyProps = useMemo(
    () => ({
      page: page,
      prepareRow: prepareRow,
      rowKey: rowKey,
      loading: loading,
      pageIndex,
      pageSize,
      clickRow,
      emptyContent,
    }),
    [page, data, pageIndex, pageSize, loading, rowKey, clickRow, emptyContent],
  )
  return (
    <div className={showBorder ? 'basic-table border-table' : 'basic-table'}>
      <table {...getTableProps()} ref={tableRef}>
        <thead
          className="m-0 overflow-y-scroll bg-[#1d1d1d]"
          style={{ borderRadius: '8px 8px 0 0' }}
        >
          {!noHeader &&
            headerGroups.map((headerGroup, idx) => (
              <tr {...headerGroup.getHeaderGroupProps()} key={`header_tr_${idx}`}>
                {headerGroup.headers.map((column, idx) => {
                  const sortedColumn = sortBy.length > 0 ? sortBy[0] : null
                  const isSorted = sortedColumn?.id === column.id
                  const isSortedDesc = isSorted ? sortedColumn?.desc : undefined

                  return (
                    <th
                      {...column.getHeaderProps(column.getSortByToggleProps())}
                      {...column.getSortByToggleProps()}
                      {...column.getHeaderProps({
                        style: {
                          width: column.customWidth,
                          flex: column.customWidth ? 'none' : '1',
                          justifyContent: column.justifyContent ? column.justifyContent : 'center',
                          padding: column.isNested ? 0 : 8,
                          minWidth: column.minWidth,
                          textDecoration: 'none',
                        },
                      })}
                      className={
                        (isSorted ? (isSortedDesc ? 'sort-desc' : 'sort-asc') : '') +
                        (column.canSort ? 'cursor-pointer no-underline' : '') +
                        (!column.isNested && 'hover:bg-[#303030]') +
                        '    hover:no-underline'
                      }
                      key={`header_th_${idx}`}
                      onClick={() => {
                        console.log(column, isSorted, isSortedDesc)
                        if (!column.disableSortBy) {
                          if (!isSorted) {
                            if (isSortedDesc === undefined) {
                              // column.toggleSortBy(false, false)
                              setSortBy([{ id: column.id, desc: false }])
                            }
                          } else {
                            if (!isSortedDesc) {
                              // column.toggleSortBy(true, false)
                              setSortBy([{ id: column.id, desc: true }])
                            } else {
                              // column.toggleSortBy()
                              setSortBy([])
                            }
                          }
                        }
                      }}
                    >
                      {column.hide &&
                        column.isNested &&
                        column.children.map((item, index) => {
                          return (
                            <th
                              style={{
                                width: item.customWidth,
                                height: '100%',
                                flex: item.customWidth ? 'none' : '1',
                                justifyContent: item.justifyContent
                                  ? item.justifyContent
                                  : 'center',

                                minWidth: item.minWidth,
                              }}
                              className="hover:bg-[#303030]"
                              key={index}
                            >
                              {item.title}
                            </th>
                          )
                        })}
                      <div className="flex justify-between items-center">
                        {!column.isNested && column.render('title')}
                        {!column.disableSortBy && (
                          <Tooltip
                            title={isSorted ? (isSortedDesc ? '取消排序' : '点击降序') : '点击升序'}
                          >
                            <div
                              className="flex flex-col cursor-pointer ml-3"
                              //
                            >
                              <VscTriangleUp
                                color={isSorted && !isSortedDesc ? '#5286e8' : 'grey'}
                              />
                              <VscTriangleDown
                                style={{ marginTop: -6 }}
                                color={isSorted && isSortedDesc ? '#5286e8' : 'grey'}
                              />
                            </div>
                          </Tooltip>
                        )}
                      </div>
                    </th>
                  )
                })}
              </tr>
            ))}
        </thead>
        {showLoading && <LoadingSpinner loading={loading || paginationLoading} />}
        <TableBody {...getTableBodyProps()} tableBodyProps={tableBodyProps}></TableBody>
      </table>
      {pagination?.pageSize && (
        <BasicPagination
          pageSize={pageSize}
          pageIndex={pageIndex}
          page={page}
          pageCount={pageCount}
          previousPage={previousPage}
          gotoPage={gotoPage}
          nextPage={nextPage}
          setPageSize={setPageSize}
        />
      )}
    </div>
  )
})

export default BasicTable
