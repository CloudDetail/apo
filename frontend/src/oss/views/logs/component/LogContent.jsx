/**
 * Copyright 2024 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import React, { useEffect, useMemo, useState } from 'react'
import { CustomSelect } from 'src/core/components/Select'
import BasicTable from 'src/core/components/Table/basicTable'
import { convertTime } from 'src/core/utils/time'

const LogContent = (props) => {
  const { data, change } = props

  const column = [
    {
      title: 'Date',
      accessor: 'timestamp',
      canExpand: false,
      customWidth: 150,
      Cell: ({ value }) => {
        return convertTime(value, 'yyyy-mm-dd hh:mm:ss')
      },
    },
    {
      title: 'Massage',
      accessor: 'body',
      justifyContent: 'flex-start',
      canExpand: false,
    },
  ]
  const tableProps = useMemo(() => {
    return {
      columns: column,
      data: data?.logContents?.contents ?? [],
      showBorder: false,
      loading: false,
    }
  }, [data])
  return (
    <>
      {/* <div className="flex-grow flex-shrink overflow-hidden"></div> */}
      <BasicTable {...tableProps} />
    </>
  )
}
export default LogContent
