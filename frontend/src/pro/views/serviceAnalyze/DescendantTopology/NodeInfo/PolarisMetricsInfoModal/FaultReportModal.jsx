import { Button, Modal, Pagination, Tag } from 'antd'
import React, { useMemo, useState } from 'react'
import { useSelector } from 'react-redux'
import { selectSecondsTimeRange } from 'src/core/store/reducers/timeRangeReducer'
import { getFailureReportApi } from 'src/pro/api/alertAnalyze'
import { convertTime } from 'src/core/utils/time'
import { useDebounce, useUpdateEffect } from 'react-use'
import BasicTable from 'src/core/components/Table/basicTable'
import { getEntryEndpointApi } from 'src/pro/api/cause'
import LoadingSpinner from 'src/core/components/Spinner'
import { useTranslation } from 'react-i18next'
import { useServiceAnalyzeContext } from 'src/pro/contexts/ServiceAnalyzeContext'
const i18nKey = 'node'

export default function FaultReportModal({
  modalVisibility,
  hidenModal,
  service,
  endpoint,
  reason,
}) {
  const { t } = useTranslation('pro/serviceAnalyze')
  const { dataGroupId } = useSelector((state) => state.dataGroupReducer)
  const { clusterIds } = useServiceAnalyzeContext()
  const { startTime, endTime } = useSelector(selectSecondsTimeRange)
  const [reportList, setReportList] = useState([])
  const [pageIndex, setPageIndex] = useState(1)
  const [pageSize, setPageSize] = useState(10)
  const [total, setTotal] = useState(0)
  const [loadingStatus, setLoadingStatus] = useState(true)

  //表格列定义
  const columns = [
    {
      title: t(`${i18nKey}.costTime`),
      accessor: 'costTime',
      key: 'costTime',
      align: 'center',
      customWidth: 200,
      Cell: (cell) => {
        const { original } = cell.cell.row
        const { duration = 0, thresholdValue = 0, thresholdMultiple = 1 } = original
        const baseline = thresholdValue / thresholdMultiple
        return (
          <div>
            {convertTime(duration, 'ms')}ms
            <Tag
              color={(duration - baseline) / baseline > 0 ? 'red' : 'green'}
              style={{ marginLeft: '5px' }}
            >
              {t(`${i18nKey}.baselineDeviation`)}
              {(((duration - baseline) / baseline) * 100).toFixed(2)}%
            </Tag>
          </div>
        )
      },
    },
    {
      title: t(`${i18nKey}.requestTime`),
      accessor: 'timestamp',
      key: 'requestTime',
      align: 'center',
      customWidth: 200,
      Cell: ({ value }) => {
        return convertTime(value, 'yyyy-mm-dd hh:mm:ss')
      },
    },
    {
      title: t(`${i18nKey}.entryInfo`),
      accessor: 'serviceInfo',
      hide: true,
      isNested: true,
      customWidth: 400,
      api: async (props) => {
        console.log(props)
        const { traceId, apmType, spanId, timestamp } = props
        try {
          const result = await getEntryEndpointApi({
            traceId,
            apmType,
            spanId,
            timestamp,
          })
          return { data: [result.data], error: null }
        } catch (error) {
          console.error('Error calling getServicesAlert:', error)
          return { data: [], error: error }
        }
      },
      children: [
        {
          title: t(`${i18nKey}.entryApp`),
          accessor: `service`,
        },
        {
          title: t(`${i18nKey}.entryEndpoint`),
          accessor: `endpoint`,
        },
      ],
    },
    {
      title: t(`${i18nKey}.viewReport`),
      accessor: 'action',
      Cell: (cell) => {
        const { value } = cell
        const { original } = cell.cell.row
        return (
          <Button
            type="primary"
            href={`/#/cause/report/${encodeURIComponent(original.traceId)}/${encodeURIComponent(original.spanId)}?mutatedType=${encodeURIComponent(reason)}`}
            target="_blank"
          >
            {t(`${i18nKey}.view`)}
          </Button>
        )
      },
      align: 'center',
    },
  ]

  //获取数据
  async function getData() {
    try {
      setLoadingStatus(true)
      const result = await getFailureReportApi({
        startTime: startTime,
        endTime: endTime,
        service: service,
        contentKey: endpoint,
        reason: reason,
        currentPage: pageIndex,
        pageSize: pageSize,
        groupId: dataGroupId,
        clusterIds: clusterIds,
      })
      const ReportList = result.list.map((item) => {
        return {
          costTime: {
            duration: item.duration,
            thresholdValue: item.thresholdValue,
          },
          requestTime: `${convertTime(item.timestamp, 'yyyy-mm-dd hh:mm:ss')}`,
          traceId: item.traceId,
          spanId: item.spanId,
          reason: item.reason,
        }
      })
      setLoadingStatus(false)
      setReportList(result.list)
      setTotal(result.pagination.total)
    } catch (error) {
      setLoadingStatus(false)
      throw new Error('error')
    }
  }

  useDebounce(
    () => {
      if (modalVisibility && dataGroupId !== null) {
        if (pageIndex === 1) {
          getData()
        } else {
          setPageIndex(1)
        }
      }
    },
    300, // 延迟时间 300ms
    [startTime, endTime, service, endpoint, reason, modalVisibility, dataGroupId, clusterIds],
  )
  useUpdateEffect(() => {
    getData()
  }, [pageIndex, pageSize])

  const handleTableChange = (page, pageSize) => {
    if (page && pageSize) {
      setPageSize(pageSize), setPageIndex(page)
    }
  }
  const tableProps = useMemo(() => {
    // 分页处理
    return {
      columns: columns,
      data: reportList,
      showBorder: false,
      // pagination: {
      //   pageSize: pageSize,
      //   pageIndex: pageIndex,
      //   pageCount: Math.ceil(total / pageSize),
      // },
      loading: false,
    }
  }, [reportList])
  const closeModal = () => {
    hidenModal()
    setReportList([])
  }
  return (
    <>
      <Modal
        open={modalVisibility}
        onCancel={closeModal}
        width="80vw"
        destroyOnClose={true}
        footer={[]}
        bodyStyle={{ height: '70vh', overflowY: 'auto', overflowX: 'hidden' }}
        title={t(`${i18nKey}.faultReportModalTitle`, { reason })}
      >
        <div className="h-full pb-5 overflow-hidden">
          <LoadingSpinner loading={loadingStatus} />
          <BasicTable {...tableProps} />
          {/* <Table
            loading={loadingStatus}
            columns={columns}
            dataSource={reportList}
            style={{
              marginTop: '5%',
              width: '100%',
              paddingRight: '2px',
              marginBottom: '10px',
            }}
            scroll={{
              y: 400,
            }}
            pagination={false}
          /> */}
          <Pagination
            align="end"
            current={pageIndex}
            pageSize={pageSize}
            pageSizeOptions={[10, 20, 30]}
            onChange={handleTableChange}
            showQuickJumper={true}
            showSizeChanger={true}
            total={total}
          />
        </div>
      </Modal>
    </>
  )
}
