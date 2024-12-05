import { Button, Card, Pagination, Tag, Tooltip, Tree } from 'antd'
import React, { useEffect, useMemo, useState } from 'react'
import { useAlertAnalyzeContext } from 'src/pro/contexts/AlertAnalyzeContext'
import TempCell from 'src/core/components/Table/TempCell'
import BasicTable from 'src/core/components/Table/basicTable'
import { getAlertDescendantContributionApi } from 'src/pro/api/alertAnalyze'
import { selectSecondsTimeRange } from 'src/core/store/reducers/timeRangeReducer'
import { useSelector } from 'react-redux'
import { getStep } from 'src/core/utils/step'
import { AiOutlineInfoCircle } from 'react-icons/ai'
import EventListModal from './EventListModal'
import { showToast } from 'src/core/utils/toast'
import ServiceEndpointTag from 'src/pro/components/ServiceEndpointTag'
import { useNavigate } from 'react-router-dom'

const ImpactList = () => {
  const { eventImpactList, setSelectedEndpointInfo } = useAlertAnalyzeContext()
  const { startTime, endTime } = useSelector(selectSecondsTimeRange)
  const [pagination, setPagination] = useState({
    pageIndex: 1,
    pageSize: 10,
  })
  const [modalEndpointInfo, setModalEndpointInfo] = useState(null)
  const [modalVisible, setModalVisible] = useState(false)
  const navigate = useNavigate()
  const column = [
    {
      title: '入口应用名称',
      accessor: 'serviceName',
      customWidth: 150,
    },
    {
      title: '入口服务端点',
      accessor: 'endpoint',
      canExpand: false,
    },

    {
      title: '平均响应时间',
      minWidth: 140,
      accessor: `latency`,

      Cell: (props) => {
        // eslint-disable-next-line react/prop-types
        const { value } = props

        return <TempCell type="latency" data={value} timeRange={{ startTime, endTime }} />
      },
    },
    {
      title: '错误率',
      accessor: `errorRate`,

      minWidth: 140,
      Cell: (props) => {
        // eslint-disable-next-line react/prop-types
        const { value } = props
        return <TempCell type="errorRate" data={value} timeRange={{ startTime, endTime }} />
      },
    },
    {
      title: '吞吐量',
      accessor: `tps`,
      minWidth: 140,

      Cell: (props) => {
        // eslint-disable-next-line react/prop-types
        const { value } = props
        return <TempCell type="tps" data={value} timeRange={{ startTime, endTime }} />
      },
    },
    {
      title: '故障贡献度',
      accessor: 'info',
      hide: true,
      isNested: true,
      // customWidth: '55%',
      api: async (props) => {
        const { serviceName, endpoint } = props
        const step = getStep(startTime, endTime)
        try {
          const result = await getAlertDescendantContributionApi({
            service: serviceName,
            endpoint: endpoint,
            startTime: startTime,
            endTime: endTime,
            step: step,
          })
          return { data: result?.latencyContributationList || [], error: null }
        } catch (error) {
          console.error('Error calling getServicesAlert:', error)
          return { data: [], error: error }
        }
      },
      children: [
        {
          title: (
            <Tooltip title="展示故障贡献度前三的下游">
              <div className="flex flex-row justify-center items-center">
                故障贡献度
                <AiOutlineInfoCircle size={16} className="ml-1" />
              </div>
            </Tooltip>
          ),
          accessor: `contribution`,
          Cell: (props) => {
            // eslint-disable-next-line react/prop-types
            const { trs } = props
            return <ServiceEndpointTag service={trs.serviceName} endpoint={trs.endpoint} />
            // return <TempCell type="logs" data={value} timeRange={requestTimeRange} />
          },
        },
      ],
    },
    {
      title: '告警影响百分比',
      accessor: `relatedAlertRate`,
      minWidth: 140,

      Cell: ({ value }) => {
        if (typeof value !== 'number' || isNaN(value) || !isFinite(value)) {
          return '0%'
        }

        // 处理 0 或空值的情况
        if (value === 0) {
          return '0%'
        }

        // 如果是有效数值则格式化保留两位小数
        return parseFloat(value.toFixed(2)) + '%'
      },
    },
    {
      title: '操作',
      accessor: 'action',
      minWidth: 140,
      Cell: (props) => {
        const row = props.row.original

        return (
          <div className="flex flex-col">
            <Button
              onClick={() => openTopologyModal(row.serviceName, row.endpoint)}
              className="my-1"
              variant="outlined"
              color="primary"
            >
              告警深入分析
            </Button>
            <Button
              className="my-1"
              variant="outlined"
              color="primary"
              onClick={() => openEventModal(row.serviceName, row.endpoint)}
            >
              告警列表
            </Button>
            <Button
              variant="outlined"
              color="primary"
              onClick={() => {
                navigate(
                  `/alert-analyze/chartAI/${encodeURIComponent(row.serviceName)}/${encodeURIComponent(row.endpoint)}`,
                )
                // showToast({
                //   title: '功能开发中，敬请期待',
                //   color: 'info',
                // })
              }}
            >
              大模型根因分析
            </Button>
          </div>
        )
      },
    },
  ]
  const openTopologyModal = (service, endpoint) => {
    // setSelectedEndpointInfo({ service, endpoint })
    navigate(
      `/alert-analyze/info?service=${encodeURIComponent(service)}&endpoint=${encodeURIComponent(endpoint)}&breadcrumb-name=业务告警分析(${encodeURIComponent(service)} -> endpoint=${encodeURIComponent(endpoint)})`,
    )
  }
  const closeModal = () => {
    setModalEndpointInfo(null)
    setModalVisible(false)
  }
  const openEventModal = (service, endpoint) => {
    setModalEndpointInfo({
      service,
      endpoint,
    })
    setModalVisible(true)
  }
  const changePagination = (page, pageSize) => {
    setPagination({
      pageSize: pageSize,
      pageIndex: page,
    })
  }
  const tableProps = useMemo(() => {
    const paginatedData = eventImpactList.slice(
      (pagination.pageIndex - 1) * pagination.pageSize,
      pagination.pageIndex * pagination.pageSize,
    )
    return {
      columns: column,
      data: paginatedData,
      showBorder: false,
      loading: false,
    }
  }, [eventImpactList, pagination.pageSize, pagination.pageIndex])
  return (
    <>
      <div className="flex flex-col h-full">
        <div className="text-base px-4 pt-2 font-bold">受告警影响的业务入口</div>
        <BasicTable {...tableProps} />
        <Pagination
          defaultCurrent={1}
          total={eventImpactList?.length || 0}
          current={pagination.pageIndex}
          pageSize={pagination.pageSize}
          className="flex-shrink-0 flex-grow-0 p-2"
          align="end"
          onChange={changePagination}
        />
        <EventListModal
          endpointInfo={modalEndpointInfo}
          modalVisible={modalVisible}
          closeModal={closeModal}
        />
      </div>
    </>
  )
}
export default ImpactList
