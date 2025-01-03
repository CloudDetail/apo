/**
 * Copyright 2024 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import React, { useMemo } from 'react'
import ReactJson from 'react-json-view'
import BasicTable from 'src/core/components/Table/basicTable'
import { convertUTCToBeijing } from 'src/core/utils/time'

export default function AlertInfoTable({ data }) {
  const column = [
    {
      title: '告警事件名',
      accessor: 'name',
      customWidth: 250,
    },
    {
      title: '告警列表',
      accessor: 'list',
      isNested: true,
      hide: true,
      children: [
        {
          title: '最近告警时间',
          accessor: 'receivedTime',
          customWidth: 180,
          Cell: ({ value }) => {
            return convertUTCToBeijing(value)
          },
        },
        {
          title: '告警详情',
          accessor: 'detail',
          Cell: ({ value }) => (
            <ReactJson
              src={JSON.parse(value)}
              theme="brewer"
              collapsed={false}
              displayDataTypes={false}
              style={{ width: '100%' }}
              enableClipboard={false}
            />
          ),
        },
      ],
    },
  ]
  const tableProps = useMemo(() => {
    return {
      columns: column,
      data: data,

      loading: false,
    }
  }, [data])
  return <>{data && <BasicTable {...tableProps} />}</>
}
