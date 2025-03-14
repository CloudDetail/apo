import { Button, Card, Pagination, Tag, Tooltip, Tree } from 'antd'
import React, { useEffect, useMemo, useState, useRef } from 'react'
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
import { useTranslation } from 'react-i18next'
import { useChartsContext } from 'src/core/contexts/ChartsContext'
import ChartTempCell from 'src/core/components/Chart/ChartTempCell'
import TableFilter from 'src/core/components/Filter/TableFilter'
import { useDebounce } from 'react-use'

const i18nKey = 'impactList'

const ImpactList = () => {
  const { t } = useTranslation('pro/alertAnalyze')
  const { eventImpactList, getAlertEventImpact } = useAlertAnalyzeContext()
  const { startTime, endTime } = useSelector(selectSecondsTimeRange)
  const [pagination, setPagination] = useState({
    pageIndex: 1,
    pageSize: 10,
  })
  const [modalEndpointInfo, setModalEndpointInfo] = useState(null)
  const [modalVisible, setModalVisible] = useState(false)
  const navigate = useNavigate()
  const getChartsData = useChartsContext((ctx) => ctx.getChartsData)
  const [serviceName, setServiceName] = useState(null)
  const [endpoint, setEndpoint] = useState(null)
  const [namespace, setNamespace] = useState(null)

  const isFirstRender = useRef(true)
  const [sortBy, setSortBy] = useState()

  const latestController = useRef(new AbortController())
  const [pageData, setPageData] = useState([])

  useDebounce(
    () => {
      if (isFirstRender.current) {
        isFirstRender.current = false
        return
      }
      getAlertEventImpact({
        services: serviceName,
        namespaces: namespace,
        endpoints: endpoint,
      })
    },
    300, // 延迟 300ms
    [serviceName, endpoint, namespace], // 依赖项
  )

  const column = [
    {
      title: t(`${i18nKey}.entryApplicationName`),
      accessor: 'serviceName',
      customWidth: 150,
    },
    {
      title: t(`${i18nKey}.entryServiceEndpoint`),
      accessor: 'endpoint',
      canExpand: false,
    },
    {
      title: t(`${i18nKey}.averageResponseTime`),
      minWidth: 140,
      accessor: `latency`,
      disableSortBy: false,
      Cell: (props) => {
        const { value, row } = props
        const { serviceName, endpoint } = row.original
        return (
          <ChartTempCell
            type="latency"
            value={value}
            service={serviceName}
            endpoint={endpoint}
            timeRange={{ startTime, endTime }}
          />
        )
      },
    },
    {
      title: t(`${i18nKey}.errorRate`),
      accessor: `errorRate`,
      minWidth: 140,
      disableSortBy: false,
      Cell: (props) => {
        const { value, row } = props
        const { serviceName, endpoint } = row.original
        return (
          <ChartTempCell
            type="errorRate"
            value={value}
            service={serviceName}
            endpoint={endpoint}
            timeRange={{ startTime, endTime }}
          />
        )
      },
    },
    {
      title: t(`${i18nKey}.throughput`),
      accessor: `tps`,
      minWidth: 140,
      disableSortBy: false,
      Cell: (props) => {
        const { value, row } = props
        const { serviceName, endpoint } = row.original
        return (
          <ChartTempCell
            type="tps"
            value={value}
            service={serviceName}
            endpoint={endpoint}
            timeRange={{ startTime, endTime }}
          />
        )
      },
    },
    {
      title: t(`${i18nKey}.faultContribution`),
      accessor: 'info',
      hide: true,
      isNested: true,
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
            <Tooltip title={t(`${i18nKey}.faultContributionHint`)}>
              <div className="flex flex-row justify-center items-center">
                {t(`${i18nKey}.faultContribution`)}
                <AiOutlineInfoCircle size={16} className="ml-1" />
              </div>
            </Tooltip>
          ),
          accessor: `contribution`,
          Cell: (props) => {
            const { trs } = props
            return <ServiceEndpointTag service={trs.serviceName} endpoint={trs.endpoint} />
          },
        },
      ],
    },
    {
      title: t(`${i18nKey}.alertImpactPercentage`),
      accessor: `relatedAlertRate`,
      minWidth: 140,
      disableSortBy: false,
      Cell: ({ value }) => {
        if (typeof value !== 'number' || isNaN(value) || !isFinite(value)) {
          return '0%'
        }
        return value === 0 ? '0%' : parseFloat(value.toFixed(2)) + '%'
      },
    },
    {
      title: t(`${i18nKey}.action`),
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
              {t(`${i18nKey}.deepAlertAnalysis`)}
            </Button>
            <Button
              className="my-1"
              variant="outlined"
              color="primary"
              onClick={() => openEventModal(row.serviceName, row.endpoint)}
            >
              {t(`${i18nKey}.alertList`)}
            </Button>
            <Button
              className='my-1'
              variant="outlined"
              color="primary"
              onClick={() => {
                window.open(`#/alert-analyze/chatAI/${encodeURIComponent(row.serviceName)}/${encodeURIComponent(row.endpoint)}`)
              }}
            >
              {t(`${i18nKey}.rootCauseAnalysis`)}
            </Button>
          </div>
        )
      },
    },
  ]
  const openTopologyModal = (service, endpoint) => {
    // setSelectedEndpointInfo({ service, endpoint })
    window.open(`#/alert-analyze/info?service=${encodeURIComponent(service)}&endpoint=${encodeURIComponent(endpoint)}&breadcrumb-name=业务告警分析(${encodeURIComponent(service)} -> endpoint=${encodeURIComponent(endpoint)})`)
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

  const getServiceCharts = async () => {
    const data = sortData()
    const paginatedData = data.slice(
      (pagination.pageIndex - 1) * pagination.pageSize,
      pagination.pageIndex * pagination.pageSize,
    )
    const endpointList = paginatedData.map((item) => item.endpoint)
    const serviceList = paginatedData.map((item) => item.serviceName)
    await getChartsData(serviceList, endpointList)
  }
  const changePagination = (page, pageSize) => {
    setPagination({
      pageSize: pageSize,
      pageIndex: page,
    })
  }
  const getSortValueByColumnId = (id, row) => {
    switch (id) {
      case 'latency':
        return row.latency.value
      case 'errorRate':
        return row.errorRate.value
      case 'tps':
        return row.tps.value
      case 'relatedAlertRate':
        return row.relatedAlertRate
    }
  }
  const sortData = () => {
    const data = [...eventImpactList]
    if (sortBy?.length > 0 && sortBy[0].desc !== undefined) {
      const columnId = sortBy[0].id
      data.sort((a, b) => {
        const rateA = getSortValueByColumnId(columnId, a)
        const rateB = getSortValueByColumnId(columnId, b)

        if (sortBy[0].desc) {
          return rateB - rateA
        } else {
          return rateA - rateB
        }
      })
    }
    return data
  }

  useEffect(() => {
    if (pageData?.length > 0) {
      latestController.current.abort()
      latestController.current = new AbortController()
      console.log(latestController.current)
      getServiceCharts()
    }
  }, [pageData])
  useEffect(() => {
    const data = sortData()
    const paginatedData = data.slice(
      (pagination.pageIndex - 1) * pagination.pageSize,
      pagination.pageIndex * pagination.pageSize,
    )
    setPageData(paginatedData)
  }, [eventImpactList, pagination.pageSize, pagination.pageIndex, sortBy])
  const tableProps = useMemo(() => {
    return {
      columns: column,
      data: pageData,
      showBorder: false,
      loading: false,
      sortBy: sortBy,
      setSortBy: setSortBy,
    }
  }, [pageData])

  return (
    <>
      <div className="flex flex-col h-full">
        <div className="text-base px-2 pt-2 font-bold flex justify-center items-center">
          <span className="shrink-0">{t(`${i18nKey}.affectedBusinessEntry`)}</span>
          <TableFilter
            className="px-4 mb-0"
            setServiceName={setServiceName}
            setEndpoint={setEndpoint}
            setNamespace={setNamespace}
          />
        </div>
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
