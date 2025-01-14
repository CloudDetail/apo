/**
 * Copyright 2024 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import React, { useMemo } from 'react'
import BasicTable from 'src/core/components/Table/basicTable'
import { useTranslation } from 'react-i18next'
import { Card } from 'antd'

function GlossaryTable() {
  const { t } = useTranslation('core/polarisMetrics')
  const data = useMemo(
    () => [
      {
        code: 'network_time',
        value: t('glossaryTable.networkTime'),
      },
      {
        code: 'CPU_time',
        value: t('glossaryTable.cpuTime'),
      },
      {
        code: 'lock_gc_time',
        value: t('glossaryTable.lockGcTime'),
      },
      {
        code: 'disk_io_time',
        value: t('glossaryTable.diskIoTime'),
      },
      {
        code: 'schedule_time',
        value: t('glossaryTable.scheduleTime'),
      },
    ],
    [t],
  )

  const columns = useMemo(
    () => [
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
    ],
    [t],
  )

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
