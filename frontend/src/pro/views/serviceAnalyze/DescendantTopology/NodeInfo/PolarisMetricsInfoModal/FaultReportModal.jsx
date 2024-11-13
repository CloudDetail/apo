import { Button, Modal, Table, Pagination, Tag } from 'antd'
import React, { useMemo, useState } from 'react'
import { useSelector } from 'react-redux'
import { selectSecondsTimeRange } from 'src/core/store/reducers/timeRangeReducer'
import { getFailureReportApi } from 'src/pro/api/alertAnalyze'
import { convertTime } from 'src/core/utils/time'
import { useDebounce, useUpdateEffect } from 'react-use'
import { color } from 'echarts'
import BasicTable from 'src/core/components/Table/basicTable'
import { getEntryEndpointApi } from 'src/pro/api/cause'
import ServiceEndpointTag from 'src/pro/components/ServiceEndpointTag'
import LoadingSpinner from 'src/core/components/Spinner'

export default function FaultReportModal({
  modalVisibility,
  hidenModal,
  service,
  endpoint,
  reason,
}) {
  const { startTime, endTime } = useSelector(selectSecondsTimeRange)
  const [reportList, setReportList] = useState([])
  const [pageIndex, setPageIndex] = useState(1)
  const [pageSize, setPageSize] = useState(10)
  const [total, setTotal] = useState(0)
  const [loadingStatus, setLoadingStatus] = useState(true)

  //表格列定义
  const columns = [
    {
      title: '本次请求耗时',
      accessor: 'costTime',
      key: 'costTime',
      align: 'center',
      customWidth: 200,
      Cell: (cell) => {
        const { original } = cell.cell.row
        const { duration = 0, thresholdValue = 0 } = original
        return (
          <div>
            {convertTime(duration, 'ms')}ms
            <Tag
              color={(duration - thresholdValue) / thresholdValue > 0 ? 'red' : 'green'}
              style={{ marginLeft: '5px' }}
            >
              偏离基线
              {((duration - thresholdValue) / thresholdValue).toFixed(2)}%
            </Tag>
          </div>
        )
      },
    },
    {
      title: '请求时间',
      accessor: 'timestamp',
      key: 'requestTime',
      align: 'center',
      customWidth: 200,
      Cell: ({ value }) => {
        return convertTime(value, 'yyyy-mm-dd hh:mm:ss')
      },
    },
    {
      title: '入口信息',
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
          title: '入口应用名称',
          accessor: `service`,
        },
        {
          title: '入口服务端点',
          accessor: `endpoint`,
        },
      ],
    },
    {
      title: '查看报告',
      accessor: 'action',
      Cell: (cell) => {
        const { value } = cell
        const { original } = cell.cell.row
        return (
          <Button
            href={`/#/cause/report/${encodeURIComponent(original.traceId)}/${encodeURIComponent(original.spanId)}?mutatedType=${encodeURIComponent(reason)}`}
            target="_blank"
          >
            查看
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
      throw new Error('请求列表错误')
    }
  }

  useDebounce(
    () => {
      if (modalVisibility) {
        if (pageIndex === 1) {
          getData()
        } else {
          setPageIndex(1)
        }
      }
    },
    300, // 延迟时间 300ms
    [startTime, endTime, service, endpoint, reason, modalVisibility],
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
        title={' 故障类型为 ' + reason + ' 的延时报告列表'}
      >
        <div className="h-[650px] pb-5 overflow-hidden">
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
