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
import { useTranslation } from 'react-i18next'

const i18nKey = 'impactList'

const ImpactList = () => {
  const { t } = useTranslation('pro/alertAnalyze')
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
      Cell: (props) => {
        const { value } = props
        return <TempCell type="latency" data={value} timeRange={{ startTime, endTime }} />
      },
    },
    {
      title: t(`${i18nKey}.errorRate`),
      accessor: `errorRate`,
      minWidth: 140,
      Cell: (props) => {
        const { value } = props
        return <TempCell type="errorRate" data={value} timeRange={{ startTime, endTime }} />
      },
    },
    {
      title: t(`${i18nKey}.throughput`),
      accessor: `tps`,
      minWidth: 140,
      Cell: (props) => {
        const { value } = props
        return <TempCell type="tps" data={value} timeRange={{ startTime, endTime }} />
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
              variant="outlined"
              color="primary"
              onClick={() => {
                navigate(
                  `/alert-analyze/chatAI/${encodeURIComponent(row.serviceName)}/${encodeURIComponent(row.endpoint)}`,
                )
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
        <div className="text-base px-4 pt-2 font-bold">{t(`${i18nKey}.affectedBusinessEntry`)}</div>
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
