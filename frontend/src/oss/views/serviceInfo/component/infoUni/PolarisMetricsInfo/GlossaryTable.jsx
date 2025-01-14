/**
 * Copyright 2024 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import { CCard } from '@coreui/react'
import React, { useMemo } from 'react'
import BasicTable from 'src/core/components/Table/basicTable'
import { useTranslation } from 'react-i18next'

function GlossaryTable() {
  const { t } = useTranslation('oss/serviceInfo')
  const data = useMemo(
    () => [
      {
        code: 'network_time',
        value: t('polarisMetricsInfo.glossaryTable.networkTime'),
      },
      {
        code: 'CPU_time',
        value: t('polarisMetricsInfo.glossaryTable.cpuTime'),
      },
      {
        code: 'lock_gc_time',
        value: t('polarisMetricsInfo.glossaryTable.lockGcTime'),
      },
      {
        code: 'disk_io_time',
        value: t('polarisMetricsInfo.glossaryTable.diskIoTime'),
      },
      {
        code: 'schedule_time',
        value: t('polarisMetricsInfo.glossaryTable.scheduleTime'),
      },
    ],
    [t],
  )

  const columns = useMemo(
    () => [
      {
        title: t('polarisMetricsInfo.glossaryTable.type'),
        accessor: 'code',
        customWidth: 100,
      },
      {
        title: t('polarisMetricsInfo.glossaryTable.meaning'),
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
    <CCard>
      <BasicTable {...tableProps}></BasicTable>
    </CCard>
  )
}

export default GlossaryTable
