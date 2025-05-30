/**
 * Copyright 2024 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import { Button, Card, Collapse, Input, Tooltip } from 'antd'
import React, { useEffect, useRef, useState } from 'react'
import IndexCollapse from './component/IndexCollapse'
import './index.css'
import FullTextSearch from '../SerarchBar/RawLogQuery/FullTextSearch'
import { useLogsContext } from 'src/core/contexts/LogsContext'
import { IoEye } from 'react-icons/io5'
import { IoMdEyeOff } from 'react-icons/io'
import { useTranslation } from 'react-i18next'

const IndexList = () => {
  const { defaultFields, hiddenFields, displayFields, resetDisplayFields, tableInfo } =
    useLogsContext()

  //由tableName和type组成的唯一标识
  const tableId = `${tableInfo.tableName}_${tableInfo.type}`

  const { t } = useTranslation('oss/fullLogs')

  const showHiddenAll = (event, type, field) => {
    event.stopPropagation()
    if (field === 'tags') {
      if (type === 'show') {
        resetDisplayFields({
          ...displayFields,
          [tableId]: [
            ...defaultFields,
            ...hiddenFields.filter((item) => displayFields[tableId].includes(item)),
          ],
        })
      } else {
        resetDisplayFields({
          ...displayFields,
          [tableId]: displayFields[tableId].filter((item) => !defaultFields.includes(item)),
        })
      }
    } else {
      if (type === 'show') {
        resetDisplayFields({
          ...displayFields,
          [tableId]: [
            ...defaultFields.filter((item) => displayFields[tableId].includes(item)),
            ...hiddenFields,
          ],
        })
      } else {
        resetDisplayFields({
          ...displayFields,
          [tableId]: displayFields[tableId].filter((item) => !hiddenFields.includes(item)),
        })
      }
    }
  }

  const items = [
    {
      key: 'base',
      label: (
        <div className="flex items-center justify-between">
          <span className="select-none">{t('indexList.basicFieldLabel')}</span>
          <div className="flex items-center">
            {defaultFields.every((item) => displayFields[tableId].includes(item)) ? (
              <Button
                type="link"
                className="p-0 m-0 h-auto"
                icon={<IoMdEyeOff size={18} />}
                onClick={(e) => showHiddenAll(e, 'hidden', 'tags')}
              ></Button>
            ) : (
              <Button
                type="link"
                className="p-0 m-0 h-auto"
                icon={<IoEye size={18} />}
                onClick={(e) => showHiddenAll(e, 'show', 'tags')}
              ></Button>
            )}
          </div>
        </div>
      ),
      children: <IndexCollapse type="base" />,
    },
    {
      key: 'log',
      label: (
        <div className="flex justify-between">
          <span className="select-none">{t('indexList.logFieldLabel')}</span>
          <div className="flex items-center">
            {hiddenFields.every((item) => displayFields[tableId].includes(item)) ? (
              <Button
                type="link"
                className="p-0 m-0 h-auto"
                icon={<IoMdEyeOff size={18} />}
                onClick={(e) => showHiddenAll(e, 'hidden', 'logs')}
              ></Button>
            ) : (
              <Button
                type="link"
                className="p-0 m-0 h-auto"
                icon={<IoEye size={18} />}
                onClick={(e) => showHiddenAll(e, 'show', 'logs')}
              ></Button>
            )}
          </div>
        </div>
      ),
      children: <IndexCollapse type="log" />,
    },
  ]

  const getItems = () =>
    tableInfo.type === 'database' ? items.filter(({ key }) => key === 'log') : items

  return (
    <div className="flex flex-col h-full">
      <div className="flex-shrink-0 py-2">
        <FullTextSearch />
      </div>
      <Collapse
        items={getItems()}
        defaultActiveKey={['base', 'log']}
        size="small"
        className=" flex-1 indexList h-full overflow-auto mb-2 mt-2"
      />
    </div>
  )
}
export default IndexList
