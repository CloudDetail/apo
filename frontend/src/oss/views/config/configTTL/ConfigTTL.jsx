/**
 * Copyright 2024 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import { CCard, CCardHeader } from '@coreui/react'
import { Button, Collapse, Popconfirm, InputNumber, List, Space, Typography } from 'antd'
import React, { useEffect, useState } from 'react'
import { FaCheck } from 'react-icons/fa'
import { getTTLApi, setTTLApi } from 'core/api/config'
import { TableType } from 'src/constants'
import TTLTable from './TTLTable'
import { showToast } from 'src/core/utils/toast'
import { IoMdInformationCircleOutline } from 'react-icons/io'

function TTLConfigInput(props) {
  const [inputValue, setInputValue] = useState(null)
  const change = (value) => {
    if (value && value > 0) {
      setInputValue(value)
    }
  }
  useEffect(() => {
    if (props.value) {
      setInputValue(props.value)
    }
  }, [props])
  const confirm = () => {
    if (inputValue && inputValue > 0) {
      props.confirmTTL(inputValue)
    }
  }
  return (
    <Space>
      <InputNumber
        value={inputValue}
        min={1}
        addonAfter="天"
        controls={false}
        className="w-28"
        onChange={change}
        changeOnBlur={true}
        precision={0}
      />
      {props.value !== inputValue && (
        <Popconfirm
          description={'确认后将同步更新' + TableType[props.type] + '数据下所有表格的数据保留周期'}
          onConfirm={confirm}
          onCancel={() => setInputValue(props.value)}
          okText="确认"
          cancelText="取消"
          title={''}
        >
          <Typography.Link
            style={{
              marginInlineEnd: 8,
            }}
          >
            更新
          </Typography.Link>
        </Popconfirm>
      )}
    </Space>
  )
}

function CollapsePanelHeader(props) {
  const { type, list, refreshPage } = props
  const [value, setValue] = useState(null)
  const confirmTypeTTL = (value) => {
    setTTLApi({
      dataType: type,
      day: value,
    })
      .then(() => {
        showToast({
          title: '配置数据保留周期可能需一定时间生效，请稍后刷新页面查看结果',
          color: 'info',
        })
      })
      .catch(() => {})
      .finally(() => {
        refreshPage()
      })
  }
  useEffect(() => {
    if (list?.length > 0) {
      if (list.every((item) => item.originalDays === list[0].originalDays)) {
        setValue(list[0].originalDays)
      } else {
        const min = Math.min(...list.map((item) => item.originalDays))
        const max = Math.max(...list.map((item) => item.originalDays))
        setValue(`${min} - ${max}`)
      }
    }
  }, [type, list])
  return (
    <>
      <Space>
        <div className="w-36">{TableType[type]}数据</div>
        <TTLConfigInput value={value} confirmTTL={confirmTypeTTL} type={type} />
      </Space>
    </>
  )
}
export default function ConfigTTL() {
  const [data, setData] = useState(null)
  const getCollapseItems = () => {
    return Object.keys(TableType).map((key) => {
      const props = {
        list: data?.[key] ?? [],
        type: key,
        refreshPage: getTTLData,
      }
      return {
        key: key,
        label: <CollapsePanelHeader {...props} />,
        children: <TTLTable {...props} />,
      }
    })
  }
  const getTTLData = () => {
    getTTLApi().then((res) => {
      setData(res)
    })
  }
  useEffect(() => {
    getTTLData()
  }, [])
  return (
    <CCard className="w-full">
      <CCardHeader className="inline-flex items-center">
        数据保留周期配置
        <Space className="text-xs ml-3 text-gray-300">
          <IoMdInformationCircleOutline size={18} color="#f7c01a" />
          更新可能需一定时间生效，请稍后刷新页面查看结果。
        </Space>
      </CCardHeader>
      <Collapse ghost size="small" items={getCollapseItems()} collapsible="icon"></Collapse>
    </CCard>
  )
}
