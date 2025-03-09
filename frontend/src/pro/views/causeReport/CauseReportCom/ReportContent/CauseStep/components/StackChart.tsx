import { Collapse, Empty } from 'antd'
import BasicTable from 'src/core/components/Table/basicTable'
import React, { useState } from 'react'
import { useTranslation } from 'react-i18next'

const i18nKey = 'stackChart'
export default function StackChart(props: any) {
  const { t } = useTranslation('pro/causeReport')

  const { stackList } = props
  const staticColumns = [
    {
      title: t(`${i18nKey}.serialNumber`),
      accessor: 'index',
      customWidth: 60,
    },
    {
      title: t(`${i18nKey}.stack`),
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
            <div className="col flex-shrink-0">
              {t(`${i18nKey}.lockStartTime`)}：{stack.waitTime}
            </div>
            <div className="col" style={{ width: 150 }}>
              {t(`${i18nKey}.elapsedTime`)}：{stack.waitThread}
            </div>
          </div>
        ),
        children: (
          <div className="row">
            <div className="col">
              {t(`${i18nKey}.lockEndTime`)}：{stack.wakeTime}
            </div>
            <div className="col">
              {t(`${i18nKey}.lockAddress`)}：{stack.lockObjectAddress}
            </div>
            <div className="align-items-start flex flex-basis-1">
              <div className="flex-shrink-0" style={{ marginTop: 15 }}>
                {t(`${i18nKey}.currentStack`)}：
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
