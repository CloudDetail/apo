import { CCard } from '@coreui/react'
import React, { useMemo } from 'react'
import BasicTable from 'src/components/Table/basicTable'

function GlossaryTable() {
  const data = [
    {
      code: 'network_time',
      value: '耗时为对外的服务调用和任何数据库和中间件，DNS等网络操作耗时',
    },
    {
      code: 'CPU_time',
      value:
        '程序代码在cpu上的执行时间.如果cpu_time过长，是由于嵌入层次过多或者循环过多导致在cpu上执行时间',
    },
    {
      code: 'lock_gc_time',
      value:
        '程序由于GC或者锁等待产生时延，如果是GC频繁，延时应该分布在不同时间点，如果是锁，一般是连续时间',
    },
    {
      code: 'disk_io_time',
      value: '程序写文件时间',
    },
    {
      code: 'schedule_time',
      value: '由于CPU争抢而产生的等待时间',
    },
  ]
  const column = [
    {
      title: '类型',
      accessor: 'code',
      customWidth: 100,
    },

    {
      title: '含义',
      justifyContent: 'left',
      accessor: 'value',
    },
  ]
  const tableProps = useMemo(() => {
    return {
      columns: column,
      data: data,

      loading: false,
    }
  }, [])
  return (
    <CCard>
      <BasicTable {...tableProps}></BasicTable>
    </CCard>
  )
}

export default GlossaryTable
