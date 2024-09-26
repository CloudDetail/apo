import { CAccordionBody } from '@coreui/react'
import React, { useMemo, useState, useEffect } from 'react'
import BasicTable from 'src/components/Table/basicTable'
import TempCell from 'src/components/Table/TempCell'
import { usePropsContext } from 'src/contexts/PropsContext'
import { selectSecondsTimeRange } from 'src/store/reducers/timeRangeReducer'
import { useSelector } from 'react-redux'
import { getStep } from 'src/utils/step'
import { getServiceSqlMetrics } from 'src/api/serviceInfo'
import LoadingSpinner from 'src/components/Spinner'
import { useDebounce, useUpdateEffect } from 'react-use'

export default function SqlMetrics() {
  const [data, setData] = useState()
  const [status, setStatus] = useState()
  const { serviceName } = usePropsContext()
  const [loading, setLoading] = useState(false)
  const { startTime, endTime } = useSelector(selectSecondsTimeRange)
  const [pageIndex, setPageIndex] = useState(1)
  const [pageSize, setPageSize] = useState(10)
  const [total, setTotal] = useState(0)
  const column = [
    {
      title: '数据库连接',
      accessor: 'dbUrl',
      Cell: ({ value }) => {
        return <>{value ? value : <span className="text-slate-400">N/A</span>}</>
      },
    },
    {
      title: '数据库操作',
      accessor: 'dbOperation',
      Cell: ({ value }) => {
        return <>{value ? value : <span className="text-slate-400">N/A</span>}</>
      },
    },

    {
      title: '数据库名',
      accessor: 'dbName',
    },
    {
      title: '平均响应时间',
      accessor: 'latency',
      minWidth: 200,
      Cell: (props) => {
        const { value } = props
        return <TempCell type="latency" data={value} timeRange={{ startTime, endTime }} />
      },
    },
    {
      title: '错误率',
      accessor: 'errorRate',
      minWidth: 200,

      Cell: (props) => {
        const { value } = props
        return <TempCell type="errorRate" timeRange={{ startTime, endTime }} data={value} />
      },
    },
    {
      title: '吞吐量',
      accessor: 'tps',

      minWidth: 200,
      Cell: (props) => {
        const { value } = props
        return <TempCell type="tps" timeRange={{ startTime, endTime }} data={value} />
      },
    },

    {
      title: '数据库源',
      accessor: 'dbSystem',
    },
  ]
  const getData = () => {
    if (startTime && endTime) {
      setLoading(true)
      getServiceSqlMetrics({
        startTime: startTime,
        endTime: endTime,
        service: serviceName,
        sortBy: 'latency',
        step: getStep(startTime, endTime),
        currentPage: pageIndex,
        pageSize: pageSize,
      })
        .then((res) => {
          setTotal(res?.pagination.total)
          setData(res.sqlOperationDetails ?? [])
          // setStatus(res.status)
          setLoading(false)
          // handlePanelStatus(res.status)
        })
        .catch((error) => {
          setData([])
          // handlePanelStatus('unknown')
          setLoading(false)
        })
    }
  }
  // useEffect(() => {
  //   if (pageIndex === 1) {
  //     getData()
  //   } else {
  //     setPageIndex(1)
  //   }
  // }, [serviceName, startTime, endTime])

  useDebounce(
    () => {
      if (pageIndex === 1) {
        getData()
      } else {
        setPageIndex(1)
      }
    },
    300, // 延迟时间 300ms
    [startTime, endTime, serviceName],
  )
  useUpdateEffect(() => {
    getData()
  }, [pageIndex, pageSize])

  const handleTableChange = (props) => {
    if (props.pageSize && props.pageIndex) {
      setPageSize(props.pageSize)
      setPageIndex(props.pageIndex)
    }
  }
  const tableProps = useMemo(() => {
    return {
      columns: column,
      data: data,
      showBorder: false,
      loading: false,
      onChange: handleTableChange,
      pagination: {
        pageSize: pageSize,
        pageIndex: pageIndex,
        pageCount: Math.ceil(total / pageSize),
      },
    }
  }, [serviceName, data])
  return (
    <>
      <CAccordionBody className="text-xs relative">
        <LoadingSpinner loading={loading} />
        {data && <BasicTable {...tableProps} />}
      </CAccordionBody>
    </>
  )
}
