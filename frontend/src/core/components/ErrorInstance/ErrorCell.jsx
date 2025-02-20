/**
 * Copyright 2024 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import { Select } from 'antd'
import React, { useEffect, useState } from 'react'
import Empty from 'src/core/components/Empty/Empty'
import { convertTime } from 'src/core/utils/time'

function ErrorCell(props) {
  const { data, update } = props
  const [options, setOptions] = useState([])
  const [selectTraceError, setSelectTraceError] = useState()
  const [menuIsOpen, setMenuIsOpen] = useState(false)
  useEffect(() => {
    const options = []
    data.map((item) => {
      item.errors.map((error) => {
        const key = convertTime(item.timestamp, 'yyyy-mm-dd hh:mm:ss') + ' ' + error.type
        options.push({
          value: key,
          label: key,
          customAbbreviation: {
            traceId: item.traceId,
            children: item.children,
            current: item.current,
            parents: item.parents,
            error: error,
            timestamp: item.timestamp,
          },
        })
      })
    })
    setOptions(options)
    setSelectTraceError(options[0])
    update(options[0])
  }, [data])
  const CustomSingleValue = (props) => {
    const { data } = props
    return (
      (data ?? selectTraceError) && (
        <div className="w-full flex-shrink" onClick={() => onSelect(data)}>
          <div className=" overflow-x-hidden whitespace-pre-wrap w-full flex flex-row">
            <div className="text-gray-400 flex-shrink-0">Time：</div>
            <div className="flex-1 w-0 whitespace-nowrap text-wrap break-all">
              {convertTime(
                (data ?? selectTraceError)?.customAbbreviation.timestamp,
                'yyyy-mm-dd hh:mm:ss',
              )}
            </div>
          </div>
          <div className=" overflow-x-hidden  w-full flex flex-row">
            <div className="text-gray-400 flex-shrink-0">ErrorType：</div>
            <div className="flex-1 w-0 whitespace-nowrap text-wrap break-all">
              {(data ?? selectTraceError)?.customAbbreviation.error.type}
            </div>
          </div>
        </div>
      )
    )
  }
  const CustomContainer = (props) => {
    const { data } = props
    return (
      <div className="w-full flex-shrink">
        <div className=" overflow-x-hidden whitespace-pre-wrap w-full flex flex-row">
          <div className="flex-1 w-0 whitespace-nowrap text-wrap break-all">
            {(data ?? selectTraceError) &&
              convertTime(
                (data ?? selectTraceError)?.customAbbreviation.timestamp,
                'yyyy-mm-dd hh:mm:ss',
              )}{' '}
            {''}
            {(data ?? selectTraceError)?.customAbbreviation.error.type}
          </div>
        </div>
      </div>
    )
  }
  const onSelect = (selectTraceError) => {
    setMenuIsOpen(false)
    setSelectTraceError(selectTraceError)
    update(selectTraceError)
  }
  return options && options.length > 0 ? (
    <div className="w-full h-full">
      <Select
        options={options}
        value={selectTraceError}
        optionRender={CustomSingleValue}
        className="w-full"
        optionFilterProp={'value'}
        dropdownStyle={{ minWidth: 200 }}
        showSearch
      ></Select>

      <div className="p-2">
        {selectTraceError && (
          <>
            <span className="text-gray-400 flex-shrink-0">Error Message：</span>
            {selectTraceError?.customAbbreviation?.error?.message}
          </>
        )}
      </div>
    </div>
  ) : (
    <Empty />
  )
}
export default ErrorCell
