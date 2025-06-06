/**
 * Copyright 2024 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import { Collapse, Tooltip, ConfigProvider, List, Button, Tag } from 'antd'
import Empty from 'src/core/components/Empty/Empty'
import React, { useEffect, useState } from 'react'
import { useLogsContext } from 'src/core/contexts/LogsContext'
import IndexCollapseItem from './IndexCollapseItem'
import { AiFillCaretDown, AiOutlineMinus, AiFillCaretUp } from 'react-icons/ai'
import { IoAddSharp, IoCloseOutline } from "react-icons/io5";
import { IoEye } from "react-icons/io5";
import { IoMdEyeOff } from "react-icons/io";
import style from "./IndexCollapse.module.less"
import { color } from 'echarts'

const IndexCollapse = (props) => {
  const { type } = props
  const [items, setItems] = useState([])
  const {
    defaultFields,
    hiddenFields,
    displayFields,
    addDisplayFields,
    removeDisplayFields,
    tableInfo,
    query = '',
    updateQuery
  } = useLogsContext()

  //由tableName和type组成的唯一标识
  const tableId = `${tableInfo.tableName}_${tableInfo.type}`

  const addToDisplayField = (event, field) => {
    event.stopPropagation()
    addDisplayFields({
      [tableId]: [...displayFields[tableId], field]
    })
  }
  const removeToDisplayField = (event, field) => {
    event.stopPropagation()
    const payload = {
      ...displayFields,
      [tableId]: displayFields[tableId]
        .filter((item) => item !== field)
    }
    removeDisplayFields(payload)
  }

  useEffect(() => {
    const fields = {
      'base': defaultFields,
      'log': hiddenFields,
    }[type]
    setItems(
      (fields ?? []).map((item) => {
        return {
          key: item.field,
          className: style.collapseItemContainer,
          label: (<div className={style.collapseItem}>
            <div className='flex-1 overflow-hidden whitespace-nowrap text-ellipsis'>
              {item}
            </div>
            {!displayFields[`${tableInfo.tableName}_${tableInfo.type}`].includes(item) ?
              <Button size='small' className='border-0' onClick={(e) => addToDisplayField(e, item)}>
                <IoMdEyeOff size={16} className='opacity-40' />
              </Button>
              :
              <Button size='small' className='border-0' onClick={(e) => removeToDisplayField(e, item)}>
                <IoEye size={16} className='opacity-60' />
              </Button>
            }
          </div>),
          children: <IndexCollapseItem field={item} />,
        }
      }),
    )
  }, [type, defaultFields, hiddenFields, displayFields])

  return (
    <>
      {items.length === 0 ? <Empty context="" width={80} /> : (
        <>
          < Collapse
            items={items}
            bordered={false}
            expandIconPosition="end"
            ghost
            className={style.collapse}
            style={tableInfo.type === 'database' ?
              { maxHeight: 'calc((100vh - 250px))', borderRadius: '0', } :
              { maxHeight: 'calc((100vh - 275px) / 2)', borderRadius: '0', }}
          // expandIcon={({ isActive }) => (isActive ? <AiFillCaretUp /> : <AiFillCaretDown />)}
          />
        </>
      )}
      {/* <List
        dataSource={items}
        bordered={false}
        renderItem={(item) => (
          <List.Item key={item} onClick={() => clickIndex(item)} className=" cursor-pointer">
            {item}
          </List.Item>
        )}
      /> */}
    </>
  )
}
export default IndexCollapse
