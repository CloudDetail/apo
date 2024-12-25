import { CAccordionBody, CToast, CToastBody } from '@coreui/react'
import React, { useState, useEffect, useMemo } from 'react'
import BasicTable from 'src/core/components/Table/basicTable'
import DelayLineChart from 'src/core/components/Chart/DelayLineChart'
import Timeline from './TimeLine'
import { usePropsContext } from 'src/core/contexts/PropsContext'
import { IoMdInformationCircleOutline } from 'react-icons/io'
import { selectSecondsTimeRange } from 'src/core/store/reducers/timeRangeReducer'
import { useSelector } from 'react-redux'
import { getStep } from 'src/core/utils/step'
import { getTraceMetricsApi } from 'core/api/serviceInfo'
import { useDebounce } from 'react-use'
import { useTranslation } from 'react-i18next'

function TraceInfo() {
  const { t } = useTranslation('oss/serviceInfo')
  const [data, setData] = useState()
  const [loading, setLoading] = useState(false)
  const { serviceName, endpoint } = usePropsContext()
  const { startTime, endTime } = useSelector(selectSecondsTimeRange)
  const column = [
    {
      title: t('traceInfo.instanceName'),
      accessor: 'name',
      customWidth: 150,
    },
    {
      title: t('traceInfo.logErrorCount'),
      accessor: 'logs',
      Cell: (props) => {
        const { value } = props
        return <DelayLineChart data={value} timeRange={{ startTime, endTime }} type={'logs'} />
      },
    },
    {
      title: t('traceInfo.responseTimeP90'),
      accessor: 'latency',
      Cell: (props) => {
        const { value } = props
        return <DelayLineChart data={value} timeRange={{ startTime, endTime }} type={'p90'} />
      },
    },
    {
      title: t('traceInfo.errorRate'),
      accessor: 'errorRate',
      Cell: (props) => {
        const { value } = props
        return <DelayLineChart data={value} timeRange={{ startTime, endTime }} type={'errorRate'} />
      },
    },
    {
      title: t('traceInfo.traceInfo'),
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
    if (startTime && endTime) {
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
  }
  // useEffect(() => {
  //   getData()
  // }, [serviceName, startTime, endTime])
  useDebounce(
    () => {
      getData()
    },
    300, // 延迟时间 300ms
    [startTime, endTime, serviceName, endpoint],
  )
  const tableProps = useMemo(() => {
    return {
      columns: column,
      data: data,
      showBorder: false,
      loading: false,
    }
  }, [data, serviceName, column])
  return (
    <>
      <CAccordionBody className="text-xs">
        <CToast autohide={false} visible={true} className="align-items-center w-full my-2">
          <div className="d-flex">
            <CToastBody className=" flex flex-row items-center text-xs">
              <IoMdInformationCircleOutline size={20} color="#f7c01a" className="mr-1" />
              {t('traceInfo.toastMessage')}
            </CToastBody>
          </div>
        </CToast>
        {data && <BasicTable {...tableProps} />}
      </CAccordionBody>
    </>
  )
}
export default TraceInfo
