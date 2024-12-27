import { CAccordionBody } from '@coreui/react'
import React, { useMemo, useState, useEffect, useRef } from 'react'
import BasicTable from 'src/core/components/Table/basicTable'
import DelayLineChart from 'src/core/components/Chart/DelayLineChart'
import { usePropsContext } from 'src/core/contexts/PropsContext'
import { useSelector } from 'react-redux'
import { selectSecondsTimeRange } from 'src/core/store/reducers/timeRangeReducer'
import { getStep } from 'src/core/utils/step'
import { getServiceErrorInstancesApi } from 'core/api/serviceInfo'
import { ErrorChain } from './ErrorChain'
import Timeline from '../TimeLine'
import { useDebounce } from 'react-use'
import ErrorCell from 'src/core/components/ErrorInstance/ErrorCell'
import { useTranslation } from 'react-i18next'

export default function ErrorInstanceInfo(props) {
  const { handlePanelStatus } = props
  const [status, setStatus] = useState()
  const [data, setData] = useState()
  const [loading, setLoading] = useState(false)
  const { serviceName, endpoint } = usePropsContext()
  const { startTime, endTime } = useSelector(selectSecondsTimeRange)
  const tableRef = useRef(null)
  const { t } = useTranslation('oss/serviceInfo')

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
      })
    })
    return { nodes, edges }
  }
  const column = useMemo(
    () => [
      {
        title: t('errorInstance.errorInstanceInfo.instanceName'),
        accessor: 'name',
        customWidth: 150,
      },
      {
        title: t('errorInstance.errorInstanceInfo.currentNodeException'),
        accessor: 'propations',
        Cell: (props) => {
          const { value, updateSelectedValue, column } = props
          const update = (item) => {
            updateSelectedValue(column.id, item)
          }

          return <ErrorCell data={value} update={update} />
        },
      },
      {
        title: t('errorInstance.errorInstanceInfo.errorPropagationChain'),
        accessor: 'chain',
        dependsOn: 'propations',
        Cell: (props) => {
          const { value, dependsValue, originalRow } = props
          return (
            <ErrorChain
              data={dependsValue?.customAbbreviation}
              instance={originalRow.name}
              chartId={originalRow.name}
            />
          )
        },
      },
      {
        title: t('errorInstance.errorInstanceInfo.logErrorCount'),
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
        title: t('errorInstance.errorInstanceInfo.logInfo'),
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
    ],
    [t, startTime, endTime],
  )

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
  useDebounce(
    () => {
      getData()
    },
    300,
    [serviceName, startTime, endTime, endpoint],
  )
  const tableProps = useMemo(() => {
    return {
      columns: column,
      data: data,
      showBorder: false,
      loading: false,
    }
  }, [column, data, serviceName])
  return (
    <>
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
