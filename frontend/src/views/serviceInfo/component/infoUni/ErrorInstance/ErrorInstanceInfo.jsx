import { CAccordionBody } from '@coreui/react'
import React, { useMemo, useState, useEffect, useRef } from 'react'
import BasicTable from 'src/components/Table/basicTable'
import DelayLineChart from 'components/Chart/DelayLineChart'
import { usePropsContext } from 'src/contexts/PropsContext'
import { useSelector } from 'react-redux'
import { selectSecondsTimeRange } from 'src/store/reducers/timeRangeReducer'
import { getStep } from 'src/utils/step'
import { getServiceErrorInstancesApi } from 'src/api/serviceInfo'
import ErrorCell from './ErrorCell'
import { ErrotChain } from './ErrorChain'
import Timeline from '../TimeLine'
import { useDebounce } from 'react-use'

export default function ErrorInstanceInfo(props) {
  const { handlePanelStatus } = props
  const [status, setStatus] = useState()
  const [data, setData] = useState()
  const [loading, setLoading] = useState(false)
  const { serviceName, endpoint } = usePropsContext()
  const { startTime, endTime } = useSelector(selectSecondsTimeRange)
  const tableRef = useRef(null)

  const prepareTopologyData = (data) => {
    if (!data) {
      return { nodes: [], edges: [] }
    }
    const current = data.current

    const nodes = [
      {
        id: 'current-' + current.instance,
        data: {
          label: current.instance,
          isTraced: current.isTraced,
          endpoint: current.endpoint,
        },
        position: { x: 0, y: 0 },
        type: 'serviceNode',
      },
    ]
    const edges = []
    data.children?.forEach((child) => {
      nodes.push({
        id: 'child-' + child.instance,
        data: {
          label: child.instance,
          isTraced: child.isTraced,
          endpoint: child.endpoint,
        },
        position: { x: 0, y: 0 },
        type: 'serviceNode',
      })
      edges.push({
        id: current.instance + '-' + child.instance,
        source: 'current-' + current.instance,
        target: 'child-' + child.instance,
      })
    })
    data.parents?.forEach((parent) => {
      nodes.push({
        id: 'parent-' + parent.instance,
        data: {
          label: parent.instance,
          isTraced: parent.isTraced,
          endpoint: parent.endpoint,
        },
        position: { x: 0, y: 0 },
        type: 'serviceNode',
      })
      edges.push({
        id: parent.instance + '-' + current.instance,
        source: 'parent-' + parent.instance,
        target: 'current-' + current.instance,
        // markerEnd: markerEnd,
        // style:{
        //   stroke: '#6293FF'
        // }
      })
    })
    return { nodes, edges }
  }
  const column = [
    {
      title: '实例名',
      accessor: 'name',
      customWidth: 150,
    },
    {
      title: '当前节点异常',
      accessor: 'propations',
      Cell: (props) => {
        const { value, updateSelectedValue, column } = props
        // updateSelectedValue(column.id, value[0]);
        const update = (item) => {
          updateSelectedValue(column.id, item)
        }

        return <ErrorCell data={value} update={update} />
      },
    },
    {
      title: '错误故障传播链',
      accessor: 'chain',
      dependsOn: 'propations', // 该列依赖于第一列的值
      Cell: (props) => {
        const { value, dependsValue, originalRow } = props
        return (
          <ErrotChain
            data={dependsValue?.customAbbreviation}
            instance={originalRow.name}
            chartId={originalRow.name}
          />
        )
        // return  nodes?.length > 0 ? <Topology canZoom={false} data={{nodes, edges}}  /> : <Empty />
      },
    },

    {
      title: '日志错误数量',
      accessor: 'logs',
      Cell: (props) => {
        const { value } = props
        return (
          <div className="h-52 w-80">
            <DelayLineChart data={value} timeRange={{ startTime, endTime }} type="logs" />
          </div>
        )
      },
    },
    {
      title: '日志信息',
      accessor: 'detail',
      customWidth: 320,
      Cell: (props) => {
        const { value, row } = props
        return (
          <Timeline
            instance={row.original.name}
            nodeName={row.original.nodeName}
            pid={row.original.pid}
            containerId={row.original.containerId}
            type="errorLogs"
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
      getServiceErrorInstancesApi({
        startTime: startTime,
        endTime: endTime,
        service: serviceName,
        endpoint: endpoint,
        step: getStep(startTime, endTime),
      })
        .then((res) => {
          setStatus(res.status)
          handlePanelStatus(res.status)
          setData(res?.instances)
          setLoading(false)
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
    [serviceName, startTime, endTime, endpoint],
  )
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
      {/* <CAccordionHeader onClick={() => handleToggle('error')}>
        {status && <StatusInfo status={status} />}
        <span className="ml-2">{serviceName}的错误实例</span>
      </CAccordionHeader> */}

      <CAccordionBody className="text-xs">
        {data && (
          <div ref={tableRef}>
            <BasicTable {...tableProps} />
          </div>
        )}
      </CAccordionBody>
    </>
  )
}
