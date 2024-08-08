import { CAccordionBody, CToast, CToastBody } from '@coreui/react'
import React, { useState, useEffect, useMemo } from 'react'
import TempCell from 'src/components/Table/TempCell'
import BasicTable from 'src/components/Table/basicTable'
import DelayLineChart from './DelayLineChart'
import Timeline from './TimeLine'
import { usePropsContext } from 'src/contexts/PropsContext'
import { IoMdInformationCircleOutline } from 'react-icons/io'
import { selectSecondsTimeRange } from 'src/store/reducers/timeRangeReducer'
import { useSelector } from 'react-redux'
import { getStep } from 'src/utils/step'
import { getTraceMetricsApi } from 'src/api/serviceInfo'
function TraceInfo() {
  const [data, setData] = useState()
  const [loading, setLoading] = useState(false)
  const { serviceName, endpoint } = usePropsContext()
  const { startTime, endTime } = useSelector(selectSecondsTimeRange)
  const column = [
    {
      title: '实例名',
      accessor: 'name',
      customWidth: 150,
    },
    {
      title: '日志错误数量',
      accessor: 'logs',
      Cell: (props) => {
        const { value } = props
        return (
          <DelayLineChart
            color="rgba(255, 158, 64, 1)"
            data={value}
            timeRange={{ startTime, endTime }}
            type={'logs'}
          />
        )
      },
    },
    {
      title: '平均响应时间P90',
      accessor: 'latency',
      Cell: (props) => {
        const { value } = props
        return (
          <DelayLineChart
            color="rgba(154, 102, 255, 1)"
            data={value}
            timeRange={{ startTime, endTime }}
            type={'latency'}
          />
        )
      },
    },
    {
      title: '错误率',
      accessor: 'errorRate',
      Cell: (props) => {
        const { value } = props
        return (
          <DelayLineChart
            color="rgba(46, 99, 255, 1)"
            data={value}
            timeRange={{ startTime, endTime }}
            type={'errorRate'}
          />
        )
      },
    },
    {
      title: '故障现场Trace',
      accessor: 'logInfo',
      customWidth: 320,
      Cell: (props) => {
        const { value, row } = props
        return (
          <Timeline
            instance={row.original.name}
            nodeName={row.original.nodeName}
            pid={row.original.pid}
            containerId={row.original.containerId}
            type="traceLogs"
            instanceName={row.values.name}
            startTime={startTime}
            endTime={endTime}
          />
        )
      },
    },
  ]
  const getData = () => {
    setLoading(true)
    getTraceMetricsApi({
      startTime: startTime,
      endTime: endTime,
      service: serviceName,
      endpoint: endpoint,
      step: getStep(startTime, endTime),
    })
      .then((res) => {
        setData(res ?? [])
        setLoading(false)
      })
      .catch((error) => {
        setData([])
        setLoading(false)
      })
  }
  useEffect(() => {
    getData()
  }, [serviceName, startTime, endTime])
  const tableProps = useMemo(() => {
    return {
      columns: column,
      data: data,
      showBorder: false,
      loading: false,
    }
  }, [data, serviceName])
  return (
    <>
      <CAccordionBody className="text-xs">
        <CToast autohide={false} visible={true} className="align-items-center w-full my-2">
          <div className="d-flex">
            <CToastBody className=" flex flex-row items-center text-xs">
              <IoMdInformationCircleOutline size={20} color="#f7c01a" className="mr-1" />
              根据日志错误数量、平均响应时间、错误率找到关键时刻Trace
            </CToastBody>
          </div>
        </CToast>
        {data && <BasicTable {...tableProps} />}
      </CAccordionBody>
    </>
  )
}
export default TraceInfo
