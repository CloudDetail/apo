/**
 * Copyright 2024 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import React, { useMemo } from 'react'
import ReactJson from 'react-json-view'
import BasicTable from 'src/core/components/Table/basicTable'
import { convertUTCToLocal } from 'src/core/utils/time'
import { useTranslation } from 'react-i18next'

export default function AlertInfoTable({ data }) {
  const { t } = useTranslation('oss/serviceInfo')
  const columns = useMemo(
    () => [
      {
        title: t('alertInfo.alertInfoTable.alertEventName'),
        accessor: 'name',
        customWidth: 250,
      },
      {
        title: t('alertInfo.alertInfoTable.alertList'),
        accessor: 'list',
        isNested: true,
        hide: true,
        children: [
          {
            title: t('alertInfo.alertInfoTable.recentAlertTime'),
            accessor: 'receivedTime',
            customWidth: 180,
            Cell: ({ value }) => {
              return convertUTCToLocal(value)
            },
          },
          {
            title: t('alertInfo.alertInfoTable.alertDetails'),
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
    ],
    [t],
  )

  const tableProps = useMemo(() => {
    return {
      columns: columns,
      data: data,
      loading: false,
      pagination: {
        pageSize: 5,
        pageIndex: 1,
        total: data?.length || 0,
      },
      scrollY: 300,
    }
  }, [columns, data])

  return <div>{data && <BasicTable {...tableProps} />}</div>
}
