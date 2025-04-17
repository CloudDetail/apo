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
import { Pagination, Tooltip } from 'antd'
import { VscTriangleUp, VscTriangleDown } from 'react-icons/vsc'
import { useTranslation } from 'react-i18next'
const BasicTable = React.memo((props) => {
  const { t } = useTranslation('core/table')
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
    showSizeChanger = false,
    scrollY = null,
    tdPadding = null,
    paginationSize = 'normal',
  } = props
  const tableRef = useRef(null)
  const debounceRef = useRef(null);
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
      manualPagination: typeof onChange === 'function',
      ...(typeof onChange === 'function'
        ? {
            pageCount: pagination?.total
              ? Math.ceil(pagination.total / pagination.pageSize)
              : undefined,
          }
        : {}),
    },
    useSortBy,
    usePagination,
  )
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
      scrollY,
      tdPadding,
    }),
    [
      pagination,
      page,
      data,
      pageIndex,
      pageSize,
      loading,
      rowKey,
      clickRow,
      emptyContent,
      scrollY,
      tdPadding,
    ],
  )
  const SortRender = (isSorted, isSortedDesc, sortType = ['asc', 'desc']) => {
    const hasAsc = sortType.includes('asc')
    const hasDesc = sortType.includes('desc')
    return (
      <Tooltip
        title={t(
          !isSorted
            ? hasAsc
              ? 'asc'
              : 'desc'
            : isSortedDesc
              ? 'unsort'
              : hasDesc
                ? 'desc'
                : 'unsort',
        )}
      >
        <div className="ml-3">
          <SortIcon isSorted={isSorted} isSortedDesc={isSortedDesc} sortType={sortType || []} />
        </div>
      </Tooltip>
    )
  }
  const getNextSortState = (isSorted, isSortedDesc, sortTypes) => {
    if (!isSorted) {
      return { desc: sortTypes.includes('desc') }
    }
    if (!isSortedDesc) {
      return { desc: true }
    }
    return null
  }
  const SortIcon = ({ isSorted, isSortedDesc, sortType }) => (
    <div className="flex flex-col cursor-pointer items-center">
      {sortType.includes('asc') && (
        <VscTriangleUp color={isSorted && !isSortedDesc ? '#5286e8' : 'grey'} />
      )}
      {sortType.includes('desc') && (
        <VscTriangleDown
          style={{ marginTop: sortType.length === 2 ? -6 : 0 }}
          color={isSorted && isSortedDesc ? '#5286e8' : 'grey'}
        />
      )}
    </div>
  )
  useEffect(() => {
    if (typeof onChange === 'function') {
      gotoPage(pagination.pageIndex - 1)
    }
  }, [pagination])

  useEffect(() => {
    const handler = (page, pageSize) => {
      setPaginationLoading(true)
      onChange?.(page, pageSize)
        ?.finally?.(() => setPaginationLoading(false))
      
      requestAnimationFrame(() => {
        gotoPage(page - 1);
        setPageSize(pageSize);
      });
    };
  
    debounceRef.current = _.debounce(handler, 300)
    
    return () => {
      debounceRef.current?.cancel();
      setPaginationLoading(false);
    };
  }, [gotoPage, setPageSize, onChange]);

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
                          const sortedColumn = sortBy.length > 0 ? sortBy[0] : null
                          const valueKey =
                            typeof item.accessor === 'function' ? item.accessor(0) : item.accessor
                          const isSorted = sortedColumn?.id === valueKey
                          const isSortedDesc = isSorted ? sortedColumn?.desc : undefined

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
                              className={
                                'hover:bg-[#303030] ' +
                                (item.sortType?.length > 0 && 'cursor-pointer  no-underline')
                              }
                              key={index}
                              onClick={() => {
                                if (!item.sortType?.length) return
                                const nextState = getNextSortState(
                                  isSorted,
                                  isSortedDesc,
                                  item.sortType,
                                )
                                if (nextState) {
                                  setSortBy([{ id: valueKey, ...nextState }])
                                } else {
                                  setSortBy([])
                                }
                              }}
                            >
                              {item.title}
                              {item.sortType?.length > 0 &&
                                SortRender(isSorted, isSortedDesc, item.sortType)}
                            </th>
                          )
                        })}
                      <div className="flex justify-between items-center">
                        {!column.isNested && column.render('title')}
                        {!column.disableSortBy && SortRender(isSorted, isSortedDesc)}
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
      {typeof pagination?.total === 'number' && pagination.total > 0 && (
        <Pagination
          defaultCurrent={1}
          total={pagination?.total}
          current={pageIndex + 1}
          pageSize={pageSize}
          align="end"
          showSizeChanger={showSizeChanger}
          onShowSizeChange={(current, pageSize) => setPageSize(pageSize)}
          onChange={(page, pageSize) => {
            setPaginationLoading(true)
            debounceRef.current(page, pageSize)
          }}
          className="mt-1"
          showTotal={(total) => (
            <span className="text-xs text-gray-400">{t('pagination', { total })}</span>
          )}
          size={paginationSize}
        />
      )}
    </div>
  )
})

export default BasicTable
