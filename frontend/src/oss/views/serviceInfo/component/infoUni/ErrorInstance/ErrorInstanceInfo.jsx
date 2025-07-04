/**
 * Copyright 2024 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import React, { useMemo, useState, useRef } from 'react'
import BasicTable from 'src/core/components/Table/basicTable'
import DelayLineChart from 'src/core/components/Chart/DelayLineChart'
import { usePropsContext } from 'src/core/contexts/PropsContext'
import { useSelector } from 'react-redux'
import { selectSecondsTimeRange } from 'src/core/store/reducers/timeRangeReducer'
import { getStep } from 'src/core/utils/step'
import { getServiceErrorInstancesApi } from 'core/api/serviceInfo'
import Timeline from '../TimeLine'
import { useDebounce } from 'react-use'
import ErrorCell from 'src/core/components/ErrorInstance/ErrorCell'
import { useTranslation } from 'react-i18next'
import { useServiceInfoContext } from 'src/oss/contexts/ServiceInfoContext'

export default function ErrorInstanceInfo() {
  const setPanelsStatus = useServiceInfoContext((ctx) => ctx.setPanelsStatus)
  const openTab = useServiceInfoContext((ctx) => ctx.openTab)
  const { dataGroupId } = useSelector((state) => state.dataGroupReducer)
  const [status, setStatus] = useState()
  const [data, setData] = useState()
  const [loading, setLoading] = useState(false)
  const { serviceName, endpoint, clusterIds } = usePropsContext()
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
          const { value, row } = props
          return <ErrorCell data={value} instance={row.values.name} />
        },
      },
      {
        title: t('errorInstance.errorInstanceInfo.logErrorCount'),
        accessor: 'logs',
        customWidth: 320,
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
        groupId: dataGroupId,
        clusterIds,
      })
        .then((res) => {
          setStatus(res.status)
          if (res?.status === 'critical') openTab('error')
          setPanelsStatus('error', res.status)
          setData(res?.instances)
          setLoading(false)
        })
        .catch((error) => {
          setData([])
          setPanelsStatus('error', 'unknown')
          setLoading(false)
        })
    }
  }
  useDebounce(
    () => {
      if (startTime && endTime && dataGroupId !== null) {
        getData()
      }
    },
    300,
    [serviceName, startTime, endTime, endpoint, dataGroupId, clusterIds],
  )
  const tableProps = useMemo(() => {
    return {
      columns: column,
      data: data,
      showBorder: false,
      loading: false,
      pagination: {
        pageSize: 5,
        pageIndex: 1,
        total: data?.length || 0,
      },
      scrollY: 500,
    }
  }, [column, data, serviceName])
  return (
    <>
      <div className="text-xs">
        {data && (
          <div ref={tableRef}>
            <BasicTable {...tableProps} />
          </div>
        )}
      </div>
    </>
  )
}
