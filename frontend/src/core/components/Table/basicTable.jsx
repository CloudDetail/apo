import React, { useEffect, useMemo, useRef, useState } from 'react'
import { useSortBy, useTable, usePagination } from 'react-table'
import './index.css'
import _ from 'lodash'
import TableBody from './tableBody'
import LoadingSpinner from '../Spinner'
import BasicPagination from './basicPagination'

const BasicTable = (props) => {
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
    emptyContent = '暂无数据',
    showLoading = true,
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
        sortBy: getSortByColumns(columns),
        pageIndex: pagination?.pageIndex ? pagination.pageIndex - 1 : 0,
        pageSize: pagination?.pageSize ? pagination.pageSize : 100000,
      },
      columns,
      data,
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
        <thead>
          {!noHeader &&
            headerGroups.map((headerGroup, idx) => (
              <tr {...headerGroup.getHeaderGroupProps()} key={`header_tr_${idx}`}>
                {headerGroup.headers.map((column, idx) => {
                  return (
                    <th
                      {...column.getHeaderProps({
                        style: {
                          width: column.customWidth,
                          flex: column.customWidth ? 'none' : '1',
                          justifyContent: column.justifyContent ? column.justifyContent : 'center',
                          padding: column.isNested ? 0 : 8,
                          minWidth: column.minWidth,
                        },
                      })}
                      className={
                        (column.isSorted ? (column.isSortedDesc ? 'sort-desc' : 'sort-asc') : '') +
                        (column.canSort ? '  cursor-pointer' : '')
                      }
                      key={`header_th_${idx}`}
                      onClick={() => {
                        if (column.canSort) {
                          column.toggleSortBy()
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
                              key={index}
                            >
                              {item.title}
                            </th>
                          )
                        })}
                      {!column.isNested && column.render('title')}
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
}

export default React.memo(BasicTable)
