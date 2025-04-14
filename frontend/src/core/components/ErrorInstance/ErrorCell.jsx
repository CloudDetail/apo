/**
 * Copyright 2024 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import { ConfigProvider, Modal, Tabs, Typography } from 'antd'
import React, { useEffect, useState } from 'react'
import Empty from 'src/core/components/Empty/Empty'
import { convertTime } from 'src/core/utils/time'
import { MdOutlineOpenInFull } from 'react-icons/md'
import CopyPre from '../CopyPre'
import { ErrorChain } from './ErrorChain'
import { v4 as uuidv4 } from 'uuid'

function safeCall(fn, ...args) {
  if (typeof fn === 'function') {
    return fn(...args)
  }
  return undefined
}

function ErrorCell(props) {
  const { data, instance, canClickTo = true, update } = props
  const [options, setOptions] = useState([])
  const [visible, setVisible] = useState(false)
  const [errorMessage, setErrorMessage] = useState(null)
  const chartId = 'chart_' + uuidv4()
  useEffect(() => {
    const options = []
    data.map((item) => {
      item.errors.map((error) => {
        const key = error.type + convertTime(item.timestamp, 'yyyy-mm-dd hh:mm:ss')
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
    safeCall(update, options[0])
  }, [data])
  const PreCom = (value) => (
    <div className="relative max-h-14 overflow-auto mb-1">
      <pre className="text-xs p-2 bg-[#161b22] text-wrap" style={{ background: '#161b22' }}>
        {value}
      </pre>
    </div>
  )
  return options && options.length > 0 ? (
    <div className="w-full text-xs  rounded p-2">
      <ConfigProvider
        theme={{
          components: {
            Tabs: {
              verticalItemMargin: '0',
              verticalItemPadding: '4px 12px 0 0',
            },
          },
        }}
      >
        <Tabs
          destroyInactiveTabPane
          tabPosition="left"
          style={{ height: 220 }}
          onChange={(key) => {
            const selected = options.find((item, i) => item.value + i === key)
            if (selected) {
              safeCall(update, selected)
            }
          }}
          items={options.map((item, i) => {
            return {
              label: (
                <div className="flex-shrink w-48 ">
                  <div className=" overflow-x-hidden whitespace-pre-wrap w-full flex flex-row text-xs">
                    <div className="text-gray-400 flex-shrink-0 text-[10px]">Time：</div>
                    <div className="flex-1 w-0 whitespace-nowrap text-wrap break-all">
                      {convertTime(item?.customAbbreviation.timestamp, 'yyyy-mm-dd hh:mm:ss')}
                    </div>
                  </div>
                  <div className=" overflow-x-hidden  w-full flex flex-row  text-xs">
                    <div className="text-gray-400 flex-shrink-0 text-[10px]">ErrorType：</div>
                    <div className="flex-1 w-0 whitespace-nowrap text-wrap break-all text-left">
                      {item?.customAbbreviation.error.type}
                    </div>
                  </div>
                </div>
              ),
              key: item.value + i,
              children: (
                <div
                  className=" text-xs h-[220px] flex justify-between flex-col"
                  key={item.value + i}
                >
                  <div className="h-1/3">
                    <span className="text-gray-400 flex-shrink-0 text-[10px]">Error Message：</span>
                    <div
                      className="relative cursor-pointer"
                      onClick={() => {
                        setErrorMessage(item?.customAbbreviation?.error?.message)
                        setVisible(true)
                      }}
                    >
                      {PreCom(item?.customAbbreviation?.error?.message)}
                      <MdOutlineOpenInFull className=" absolute top-1 right-1" color="#3b82f6" />
                    </div>
                  </div>
                  <span className="text-gray-400 flex-shrink-0 text-[10px]">
                    Error Propagation Chain：
                  </span>
                  <div className="h-0 flex-1">
                    <ErrorChain
                      data={item?.customAbbreviation}
                      instance={instance}
                      chartId={chartId}
                      canClickTo={canClickTo}
                    />
                  </div>
                </div>
              ),
            }
          })}
        />
      </ConfigProvider>
      <Modal
        open={visible}
        width="60vw"
        onCancel={() => setVisible(false)}
        onOk={() => setVisible(false)}
        destroyOnClose
        centered
        footer={(_, { OkBtn }) => <></>}
        maskClosable={true}
      >
        <Typography className=" relative w-full max-h-[500px]">
          <CopyPre code={errorMessage} />
        </Typography>
      </Modal>
    </div>
  ) : (
    <Empty />
  )
}
export default ErrorCell
