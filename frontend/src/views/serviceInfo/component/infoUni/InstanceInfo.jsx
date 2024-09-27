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
import { useDebounce } from 'react-use'

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
      minWidth: 150,
      Cell: (props) => {
        const { value } = props
        return <TempCell type="latency" data={value} timeRange={{ startTime, endTime }} />
      },
    },
    {
      title: '错误率',
      accessor: 'errorRate',
      minWidth: 150,

      Cell: (props) => {
        const { value } = props
        return <TempCell type="errorRate" timeRange={{ startTime, endTime }} data={value} />
      },
    },
    {
      title: '吞吐量',
      accessor: 'tps',

      minWidth: 150,
      Cell: (props) => {
        const { value } = props
        return <TempCell type="tps" timeRange={{ startTime, endTime }} data={value} />
      },
    },
    {
      title: '日志错误数量',
      accessor: 'logs',

      minWidth: 150,
      Cell: (props) => {
        const { value } = props
        return <TempCell type="logs" timeRange={{ startTime, endTime }} data={value} />
      },
    },
    {
      title: '基础设施状态',
      accessor: 'infrastructureStatus',
      Cell: (props) => {
        const { value, row, column } = props
        const alertReason = row.original?.alertReason?.[column.id]
        return (
          <>
            <StatusInfo status={value} alertReason={alertReason} title={column.title} />
          </>
        )
      },
    },
    {
      title: '网络质量状态',
      accessor: 'netStatus',
      Cell: (props) => {
        const { value, row, column } = props
        const alertReason = row.original?.alertReason?.[column.id]
        return (
          <>
            <StatusInfo status={value} alertReason={alertReason} title={column.title} />
          </>
        )
      },
    },
    {
      title: 'K8s事件状态',
      accessor: 'k8sStatus',
      Cell: (props) => {
        const { value, row, column } = props
        const alertReason = row.original?.alertReason?.[column.id]
        return (
          <>
            <StatusInfo status={value} alertReason={alertReason} title={column.title} />
          </>
        )
      },
    },

    {
      title: '主机节点信息',
      accessor: 'nodeName',
      minWidth: 150,
      Cell: (props) => {
        return (
          <div>
            <div className="flex ">
              <span className="text-gray-400 flex-shrink-0 flex-grow-0">主机名：</span>
              {props.value}
            </div>
            <div className="flex">
              <span className="text-gray-400 flex-shrink-0 flex-grow-0">主机IP：</span>
              {props.row.original.nodeIP}
            </div>
          </div>
        )
      },
    },
    {
      title: '末次部署或重启时间',
      accessor: `timestamp`,
      Cell: (props) => {
        const { value } = props
        return (
          <>
            {value !== null ? (
              convertTime(value, 'yyyy-mm-dd hh:mm:ss')
            ) : (
              <span className="text-slate-400">N/A</span>
            )}
          </>
        )
      },
    },
  ]
  const getData = () => {
    if (startTime && endTime) {
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
  }
  // useEffect(() => {
  //   getData()
  // }, [serviceName, startTime, endTime])
  //防抖避免跳转使用旧时间
  useDebounce(
    () => {
      getData()
    },
    300, // 延迟时间 300ms
    [startTime, endTime, serviceName, endpoint],
  )
  useEffect(() => {
    const namespaceList = [
      ...new Set(data?.map((obj) => obj.namespace).filter((namespace) => namespace !== '')),
    ]
    const podList = [...new Set(data?.map((obj) => obj.name).filter((name) => name !== ''))]
    prepareVariable({ namespaceList, podList, service: serviceName })
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
