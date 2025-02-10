import { Collapse, Empty } from 'antd'
import BasicTable from 'src/core/components/Table/basicTable'
import React, { useState } from 'react'

export default function StackChart(props: any) {
  const { stackList } = props
  const staticColumns = [
    {
      title: '序号',
      accessor: 'index',
      customWidth: 60,
    },
    {
      title: '堆栈',
      accessor: 'value',
      justifyContent: 'start',
    },
  ]

  const getItems = () => {
    return stackList?.map((stack: any, index: number) => {
      const tableData = stack.stack
        .split(';')
        .filter((item: any) => item.trim() !== '')
        .map((item: any, index: number) => {
          return { index: index, value: item }
        })
      return {
        key: index,
        label: (
          <div className="row">
            <div className="col flex-shrink-0">锁开始时间：{stack.waitTime}</div>
            <div className="col" style={{ width: 150 }}>
              耗时：{stack.waitThread}
            </div>
          </div>
        ),
        children: (
          <div className="row">
            <div className="col">锁结束时间：{stack.wakeTime}</div>
            <div className="col">锁地址：{stack.lockObjectAddress}</div>
            <div className="align-items-start flex flex-basis-1 ">
              <div className="flex-shrink-0" style={{ marginTop: 15 }}>
                当前堆栈：
              </div>
              {tableData.length > 0 ? (
                <BasicTable columns={staticColumns} data={tableData} />
              ) : (
                <Empty />
              )}
            </div>
          </div>
        ),
      }
    })
  }
  return (
    <>
      <Collapse items={getItems()} />
    </>
  )
}
