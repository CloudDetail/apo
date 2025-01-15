/**
 * Copyright 2024 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import React, { useMemo } from 'react'
import BasicTable from 'src/core/components/Table/basicTable'
import { useTranslation } from 'react-i18next'
import { Card } from 'antd'
import { SlowErrorType } from 'src/constants'

function GlossaryTable() {
  const { t } = useTranslation('core/polarisMetrics')
  const data = Object.entries(SlowErrorType).map(([key, value]) => ({
    code: key,
    value: value,
  }))
  const columns = [
    {
      title: t('glossaryTable.type'),
      accessor: 'code',
      customWidth: 120,
    },
    {
      title: t('glossaryTable.meaning'),
      justifyContent: 'left',
      accessor: 'value',
    },
  ]

  const tableProps = useMemo(() => {
    return {
      columns: columns,
      data: data,
      loading: false,
    }
  }, [columns, data])

  return (
    <Card classNames={{ body: 'py-0' }}>
      <BasicTable {...tableProps}></BasicTable>
    </Card>
  )
}

export default GlossaryTable
