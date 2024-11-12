import { Button, Modal, Table, Pagination, Tag } from 'antd'
import React, { useState } from 'react'
import { useSelector } from 'react-redux'
import { selectSecondsTimeRange } from 'src/core/store/reducers/timeRangeReducer'
import { getFailureReportApi } from 'src/pro/api/alertAnalyze'
import { convertTime } from 'src/core/utils/time'
import { useDebounce } from 'react-use'
import { color } from 'echarts'

export default function FaultReportModal({
  modalVisibility,
  hidenModal,
  service,
  endpoint,
  reason,
}) {
  const { startTime, endTime } = useSelector(selectSecondsTimeRange)
  const [reportList, setReportList] = useState([])
  const [currentPageSize, setCurrentPageSize] = useState(10)
  const [currentPage, setCurrentPage] = useState(1)
  const [total, setTotal] = useState(0)
  const [loadingStatus, setLoadingStatus] = useState(true)

  //表格列定义
  const columns = [
    {
      title: '本次请求耗时',
      dataIndex: 'costTime',
      key: 'costTime',
      align: 'center',
      render: (value) => (
        <div>
          {convertTime(value.duration, 'ms')}ms
          <Tag
            color={
              (value.duration - value.thresholdValue) / value.thresholdValue > 0 ? 'red' : 'green'
            }
            style={{ marginLeft: '5px' }}
          >
            偏离基线{((value.duration - value.thresholdValue) / value.thresholdValue).toFixed(2)}%
          </Tag>
        </div>
      ),
    },
    {
      title: '请求时间',
      dataIndex: 'requestTime',
      key: 'requestTime',
      align: 'center',
    },
    {
      title: '查看报告',
      dataIndex: 'action',
      render: (text, record, index) => {
        return (
          <Button
            href={`/#/cause/report/${encodeURIComponent(record.traceId)}/${encodeURIComponent(record.spanId)}?mutatedType=${encodeURIComponent(reason)}`}
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
        currentPage: currentPage,
        pageSize: currentPageSize,
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
      setReportList(ReportList)
      setTotal(result.pagination.total)
    } catch (error) {
      setLoadingStatus(false)
      throw new Error('请求列表错误')
    }
  }

  function pageChange(page, pageSize) {
    setCurrentPage(page)
    setCurrentPageSize(pageSize)
  }

  useDebounce(
    () => {
      if (modalVisibility) {
        getData()
      }
    },
    300,
    [currentPageSize, currentPage, service, endpoint, reason, modalVisibility],
  )

  return (
    <>
      <Modal
        open={modalVisibility}
        onCancel={hidenModal}
        width="60vw"
        height="50vh"
        destroyOnClose={true}
        footer={[]}
      >
        <>
          <Table
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
          />
          <Pagination
            align="center"
            current={currentPage}
            pageSize={currentPageSize}
            pageSizeOptions={[10, 20, 30]}
            onChange={pageChange}
            showQuickJumper={true}
            showSizeChanger={true}
            total={total}
          />
        </>
      </Modal>
    </>
  )
}
