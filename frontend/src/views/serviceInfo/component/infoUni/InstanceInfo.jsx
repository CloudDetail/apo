import { CAccordion, CAccordionBody, CAccordionHeader, CAccordionItem } from '@coreui/react'
import React, { useMemo, useState, useEffect } from 'react'
import BasicTable from 'src/components/Table/basicTable'
import { instanceMock } from '../mock'
import StatusInfo from 'src/components/StatusInfo'
import TempCell from 'src/components/Table/TempCell'
import { usePropsContext } from 'src/contexts/PropsContext'
import { getServiceInstancesApi } from 'src/api/serviceInfo'
import { selectSecondsTimeRange } from 'src/store/reducers/timeRangeReducer'
import { useSelector } from 'react-redux'
import { getStep } from 'src/utils/step'
import { convertTime } from 'src/utils/time'

export default function InstanceInfo(props) {
  const { handlePanelStatus, prepareVariable } = props
  const [data, setData] = useState()
  const [status, setStatus] = useState()
  const { serviceName, endpoint } = usePropsContext()
  const [loading, setLoading] = useState(false)
  const { startTime, endTime } = useSelector(selectSecondsTimeRange)
  const column = [
    {
      title: '实例名',
      accessor: 'name',
      customWidth: 150,
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
      title: '日志错误数量',
      accessor: 'logs',

      minWidth: 200,
      Cell: (props) => {
        const { value } = props
        return <TempCell type="logs" timeRange={{ startTime, endTime }} data={value} />
      },
    },
    {
      title: '基础设施状态',
      accessor: 'infrastructureStatus',
      Cell: (props) => {
        const { value } = props
        return (
          <>
            <StatusInfo status={value} />
          </>
        )
      },
    },
    {
      title: '网络质量状态',
      accessor: 'netStatus',
      Cell: (props) => {
        const { value } = props
        return (
          <>
            <StatusInfo status={value} />
          </>
        )
      },
    },
    {
      title: 'K8s事件状态',
      accessor: 'k8sStatus',
      Cell: (props) => {
        const { value } = props
        return (
          <>
            <StatusInfo status={value} />
          </>
        )
      },
    },
    {
      title: '末次部署或重启时间',
      accessor: `timestamp`,
      Cell: (props) => {
        const { value } = props
        return <>{value !== null ? convertTime(value, 'yyyy-mm-dd hh:mm:ss') : 'N/A'} </>
      },
    },
  ]
  const getData = () => {
    setLoading(true)
    getServiceInstancesApi({
      startTime: startTime,
      endTime: endTime,
      serviceName: serviceName,
      endpoint: endpoint,
      step: getStep(startTime, endTime),
    })
      .then((res) => {
        setData(res.data ?? [])
        setStatus(res.status)
        setLoading(false)
        handlePanelStatus(res.status)
      })
      .catch((error) => {
        setData([])
        handlePanelStatus('unknown')
        setLoading(false)
      })
  }
  useEffect(() => {
    getData()
  }, [serviceName, startTime, endTime])

  useEffect(() => {
    const namespaceList = [
      ...new Set(data?.map((obj) => obj.namespace).filter((namespace) => namespace !== '')),
    ]
    const podList = [...new Set(data?.map((obj) => obj.name).filter((name) => name !== ''))]
    prepareVariable({ namespaceList, podList })
  }, [data])
  const tableProps = useMemo(() => {
    return {
      columns: column,
      data: data,
      showBorder: false,
      loading: false,
    }
  }, [serviceName, data])
  return (
    <>
      {/* <CAccordionHeader onClick={() => handleToggle('instance')}>
        {status && <StatusInfo status={status} />}
        <span className="ml-2">{serviceName}的应用URL实例</span>
      </CAccordionHeader> */}
      <CAccordionBody className="text-xs">{data && <BasicTable {...tableProps} />}</CAccordionBody>
    </>
  )
}
