/**
 * Copyright 2024 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import { useEffect, useMemo, useState } from 'react'
import BasicTable from 'src/core/components/Table/basicTable'
import { CButton } from '@coreui/react'
import { useNavigate } from 'react-router-dom'
import TempCell from 'src/core/components/Table/TempCell'
import StatusInfo from 'src/core/components/StatusInfo'
import { getServicesAlertApi, getServicesEndpointsApi } from 'core/api/service'
import { useSelector } from 'react-redux'
import { selectSecondsTimeRange } from 'src/core/store/reducers/timeRangeReducer'
import { getStep } from 'src/core/utils/step'
import { DelaySourceTimeUnit } from 'src/constants'
import { convertTime } from 'src/core/utils/time'
import EndpointTableModal from './component/EndpointTableModal'
import LoadingSpinner from 'src/core/components/Spinner'
import { Tooltip } from 'antd'
import { AiOutlineInfoCircle } from 'react-icons/ai'
import { useDebounce } from 'react-use'
import { useTranslation } from 'react-i18next'
import React from 'react'
import { ChartsProvider, useChartsContext } from 'src/core/contexts/ChartsContext'
import ChartTempCell from 'src/core/components/Chart/ChartTempCell'
import { BasicCard } from 'src/core/components/Card/BasicCard'
import DataSourceFilter from 'src/core/components/Filter/DataSourceFilter'
const ServiceTable = React.memo(() => {
  const { t, i18n } = useTranslation('oss/service')
  const navigate = useNavigate()
  const [data, setData] = useState([])
  const [loading, setLoading] = useState(true)
  const [modalVisible, setModalVisible] = useState(false)
  const [modalServiceName, setModalServiceName] = useState()
  const [requestTimeRange, setRequestTimeRange] = useState({
    startTime: null,
    endTime: null,
  })
  const [cluster, setCluster] = useState(null)
  const [serviceName, setServiceName] = useState(null)
  const [endpoint, setEndpoint] = useState(null)
  const [namespace, setNamespace] = useState(null)
  const [pageIndex, setPageIndex] = useState(1)
  const [pageSize, setPageSize] = useState(10)
  const { startTime, endTime } = useSelector(selectSecondsTimeRange)
  const { dataGroupId } = useSelector((state) => state.dataGroupReducer)
  const getChartsData = useChartsContext((ctx) => ctx.getChartsData)
  const [sortBy, setSortBy] = useState()
  const column = [
    {
      title: t('index.serviceTableColumns.serviceName'),
      accessor: 'serviceName',
      customWidth: 150,
    },
    {
      title: t('index.serviceTableColumns.serviceDetails.title'),
      accessor: 'serviceDetails',
      hide: true,
      isNested: true,
      customWidth: '55%',
      clickCell: (props) => {
        // const navigate = useNavigate()
        // toServiceInfo()
        const serviceName = props.cell.row.values.serviceName
        const endpoint = props.trs.endpoint
        const namespace = props.cell.row.values.namespaces
        // cluster 变量就是 clusterIds
        const clusterIdsParam = cluster
          ? `&clusterIds=${encodeURIComponent(Array.isArray(cluster) ? cluster.join(',') : cluster)}`
          : ''
        navigate(
          `/service/info?service-name=${encodeURIComponent(serviceName)}&endpoint=${encodeURIComponent(endpoint)}&breadcrumb-name=${encodeURIComponent(serviceName)}&namespace=${encodeURIComponent(namespace)}&groupId=${dataGroupId}${clusterIdsParam}`,
        )
      },
      showMore: (original) => {
        const clickMore = () => {
          setModalVisible(true)
          setModalServiceName(original.serviceName)
        }
        return (
          original.endpointCount > 3 && (
            <CButton color="info" variant="ghost" size="sm" onClick={clickMore}>
              {t('index.serviceTableColumns.serviceDetails.showMore')}
            </CButton>
          )
        )
        // return
      },

      children: [
        {
          title: t('index.serviceTableColumns.serviceDetails.endpoint'),
          accessor: 'endpoint',
          canExpand: false,
        },
        {
          title: (
            <Tooltip
              title={
                <div>
                  <p className="text-[#D3D3D3]">
                    {t('index.serviceTableColumns.serviceDetails.delaySource.title1')}
                  </p>
                  <p className="text-[#D3D3D3] mt-2">
                    {t('index.serviceTableColumns.serviceDetails.delaySource.title2')}
                  </p>
                  <p className="text-[#D3D3D3] mt-2">
                    {t('index.serviceTableColumns.serviceDetails.delaySource.title3')}
                  </p>
                </div>
              }
            >
              <div className="flex flex-row justify-center items-center">
                <div>
                  <div className="text-center flex flex-row">
                    {t('index.serviceTableColumns.serviceDetails.delaySource.title4')}
                  </div>
                  <div className="block text-[10px]">
                    {t('index.serviceTableColumns.serviceDetails.delaySource.title5')}
                  </div>
                </div>
                <AiOutlineInfoCircle size={16} className="ml-1" />
              </div>
            </Tooltip>
          ),
          accessor: 'delaySource',
          canExpand: false,
          customWidth: 150,
          Cell: (props) => {
            const { value } = props
            return <>{DelaySourceTimeUnit[value]}</>
          },
        },
        {
          title: t('index.serviceTableColumns.serviceDetails.latency'),
          minWidth: 140,
          accessor: (idx) => `latency`,
          sortType: ['desc'],

          Cell: (props) => {
            const { value, cell, trs } = props
            const { serviceName } = cell.row.original
            const { endpoint } = trs
            return (
              <ChartTempCell
                type="latency"
                value={value}
                service={serviceName}
                endpoint={endpoint}
                timeRange={requestTimeRange}
              />
            )
          },
        },
        {
          title: t('index.serviceTableColumns.serviceDetails.errorRate'),
          accessor: (idx) => `errorRate`,
          sortType: ['desc'],
          minWidth: 140,
          Cell: (props) => {
            const { value, cell, trs } = props
            const { serviceName } = cell.row.original
            const { endpoint } = trs
            return (
              <ChartTempCell
                type="errorRate"
                value={value}
                service={serviceName}
                endpoint={endpoint}
                timeRange={requestTimeRange}
              />
            )
          },
        },
        {
          title: t('index.serviceTableColumns.serviceDetails.tps'),
          accessor: (idx) => `tps`,
          minWidth: 140,
          sortType: ['desc'],
          Cell: (props) => {
            const { value, cell, trs } = props
            const { serviceName } = cell.row.original
            const { endpoint } = trs
            return (
              <ChartTempCell
                type="tps"
                value={value}
                service={serviceName}
                endpoint={endpoint}
                timeRange={requestTimeRange}
              />
            )
          },
        },
      ],
    },
    {
      title: t('index.serviceTableColumns.serviceInfo.title'),
      accessor: 'serviceInfo',
      hide: true,
      isNested: true,
      // customWidth: '55%',
      api: async (props) => {
        const { serviceName } = props
        try {
          const result = await getServicesAlert(serviceName)
          return { data: result, error: null }
        } catch (error) {
          console.error('Error calling getServicesAlert:', error)
          return { data: [], error: error }
        }
      },
      children: [
        {
          title: t('index.serviceTableColumns.serviceInfo.logs'),
          accessor: `logs`,
          sortType: ['desc'],
          minWidth: 160,
          Cell: (props) => {
            // eslint-disable-next-line react/prop-types
            const { value } = props
            return <TempCell type="logs" data={value} timeRange={requestTimeRange} />
          },
        },
        {
          title: t('index.serviceTableColumns.serviceInfo.infrastructureStatus'),
          accessor: `infrastructureStatus`,
          Cell: (props) => {
            // eslint-disable-next-line react/prop-types
            const { value, trs, column } = props
            const alertReason = trs?.alertReason?.[column.accessor]
            return (
              <>
                <StatusInfo status={value} alertReason={alertReason} title={column.title} />
              </>
            )
          },
        },
        {
          title: t('index.serviceTableColumns.serviceInfo.netStatus'),
          accessor: `netStatus`,
          Cell: (/** @type {{ value: any; }} */ props) => {
            // eslint-disable-next-line react/prop-types
            const { value, trs, column } = props
            const alertReason = trs?.alertReason?.[column.accessor]
            return (
              <>
                <StatusInfo status={value} alertReason={alertReason} title={column.title} />
              </>
            )
          },
        },
        {
          title: t('index.serviceTableColumns.serviceInfo.k8sStatus'),
          accessor: `k8sStatus`,
          Cell: (props) => {
            // eslint-disable-next-line react/prop-types
            const { value, trs, column } = props
            const alertReason = trs?.alertReason?.[column.accessor]
            return (
              <>
                <StatusInfo status={value} alertReason={alertReason} title={column.title} />
              </>
            )
          },
        },
        {
          title: t('index.serviceTableColumns.serviceInfo.timestamp'),
          accessor: `timestamp`,
          minWidth: 90,
          Cell: (props) => {
            const { value } = props
            return (
              <>
                {value !== null ? (
                  convertTime(value, 'yyyy-mm-dd hh:mm:ss')
                ) : (
                  <span className="text-slate-400">N/A</span>
                )}{' '}
              </>
            )
          },
        },
      ],
    },
  ]
  const getTableData = () => {
    if (startTime && endTime) {
      setLoading(true)
      // 记录请求的时间范围，以便后续趋势图补0
      setRequestTimeRange({
        startTime: startTime,
        endTime: endTime,
      })
      const sortRule = getSortRule(sortBy?.length > 0 ? sortBy[0].id : null)

      getServicesEndpointsApi({
        startTime: startTime,
        endTime: endTime,
        step: getStep(startTime, endTime),
        serviceName: serviceName ?? undefined,
        endpointName: endpoint ?? undefined,
        namespace: namespace ?? undefined,
        sortRule: sortRule,
        groupId: dataGroupId,
        clusterIds: cluster ?? undefined,
      })
        .then((res) => {
          if (res && res?.length > 0) {
            setData(res)
          } else {
            setData([])
          }
          setPageIndex(1)
          setLoading(false)
        })
        .catch(() => {
          setPageIndex(1)
          setData([])
          setLoading(false)
        })
    }
  }
  //防抖避免跳转使用旧时间
  useDebounce(
    () => {
      if (startTime && endTime && dataGroupId !== null) {
        getTableData()
      }
    },
    300, // 延迟时间 300ms
    [startTime, endTime, serviceName, endpoint, namespace, dataGroupId, sortBy, cluster],
  )
  const getServicesAlert = (serviceName, returnData) => {
    return getServicesAlertApi({
      startTime: startTime,
      endTime: endTime,
      step: getStep(startTime, endTime),
      serviceName: serviceName,
      returnData: returnData,
      groupId: dataGroupId,
      clusterIds: cluster ?? undefined,
    })
      .then((res) => {
        if (res && res?.length > 0) {
          return res
        } else {
          return []
        }
        // setLoading(false)
      })
      .catch(() => {
        return []
        // setLoading(false)
      })
  }
  useEffect(() => {
    if (data) getServiceCharts()
  }, [pageIndex, pageSize, data])
  const getServiceCharts = async () => {
    const paginatedData = data.slice((pageIndex - 1) * pageSize, pageIndex * pageSize)
    const serviceList = []
    const endpointList = []
    paginatedData.forEach((item) => {
      serviceList.push(item.serviceName)
      item.serviceDetails.forEach((endpoint) => {
        endpointList.push(endpoint.endpoint)
      })
    })
    getChartsData(serviceList, endpointList, dataGroupId, cluster)
  }
  const handleTableChange = (pageIndex, pageSize) => {
    if (pageSize && pageIndex) {
      setPageSize(pageSize), setPageIndex(pageIndex)
    }
  }
  const getSortRule = (sortById) => {
    switch (sortById) {
      case 'latency':
        return 3
      case 'errorRate':
        return 4
      case 'tps':
        return 5
      case 'logs':
        return 6
      default:
        return 1
    }
  }
  const tableProps = useMemo(() => {
    const paginatedData = data.slice((pageIndex - 1) * pageSize, pageIndex * pageSize)

    return {
      columns: column,
      data: paginatedData,
      onChange: handleTableChange,
      pagination: {
        pageSize: pageSize,
        pageIndex: pageIndex,
        total: data.length,
      },
      emptyContent: (
        <div className="text-center">
          {t('index.tableProps.emptyText')}
          <div className="text-left p-2">
            <div className="py-2">
              {t('index.tableProps.unmonitoredText')}
              <a
                className="underline text-sky-500"
                target="_blank"
                href={
                  i18n.language === 'zh'
                    ? 'https://kindlingx.com/docs/APO%20向导式可观测性中心/安装手册/安装%20APO-OneAgent/'
                    : 'https://docs.autopilotobservability.com/Installation/APO%20OneAgent'
                }
              >
                {t('index.tableProps.unmonitoredLinkText')}
              </a>
            </div>
            <div>
              {t('index.tableProps.monitoredText')}
              <a
                className="underline text-sky-500"
                target="_blank"
                href={
                  i18n.language === 'zh'
                    ? 'https://kindlingx.com/docs/APO%20向导式可观测性中心/常见问题/运维与故障排除/APO%20服务概览无数据排查文档'
                    : 'https://docs.autopilotobservability.com/Troubleshooting/Service%20Overview%20No%20Data'
                }
              >
                {t('index.tableProps.monitoredLinkText')}
              </a>
            </div>
          </div>
        </div>
      ),
      showLoading: false,
      sortBy: sortBy,
      setSortBy: setSortBy,
    }
  }, [data, pageIndex, pageSize, i18n.language])
  return (
    <BasicCard>
      <LoadingSpinner loading={loading} />

      <BasicCard.Header>
        <DataSourceFilter
          setServiceName={setServiceName}
          setEndpoint={setEndpoint}
          setNamespace={setNamespace}
          setCluster={setCluster}
          startTime={startTime}
          endTime={endTime}
          className="mb-2"
          category="apm"
          extra="endpoint"
        />
        {/* <TableFilter
          dataGroupId={dataGroupId}
          setServiceName={setServiceName}
          setEndpoint={setEndpoint}
          setNamespace={setNamespace}
          className="mb-2"
        /> */}
      </BasicCard.Header>

      <BasicCard.Table>
        <BasicTable {...tableProps} />
      </BasicCard.Table>

      <ChartsProvider>
        <EndpointTableModal
          visible={modalVisible}
          serviceName={modalServiceName}
          timeRange={requestTimeRange}
          clusterIds={cluster}
          closeModal={() => setModalVisible(false)}
        />
      </ChartsProvider>
    </BasicCard>
  )
})
export default ServiceTable
