import { Button, Modal, Pagination } from 'antd'
import React, { useEffect, useMemo, useState } from 'react'
import ReactJson from 'react-json-view'
import { useSelector } from 'react-redux'
import { getAlertDescendantEventApi } from 'src/pro/api/alertAnalyze'
import BasicTable from 'src/core/components/Table/basicTable'
import TempCell from 'src/core/components/Table/TempCell'
import { selectSecondsTimeRange } from 'src/core/store/reducers/timeRangeReducer'
import { convertUTCToBeijing } from 'src/core/utils/time'
function isJSONString(str) {
  try {
    JSON.parse(str)
    return true
  } catch (e) {
    return false
  }
}
const EventListModal = ({ endpointInfo, modalVisible, closeModal }) => {
  console.log(endpointInfo)
  const { startTime, endTime } = useSelector(selectSecondsTimeRange)
  const { groupLabel } = useSelector((state) => state.groupLabelReducer)
  const [pagination, setPagination] = useState({
    pageIndex: 1,
    pageSize: 10,
    total: 0,
  })
  const [data, setData] = useState([])
  const [loading, setLoading] = useState(false)
  useEffect(() => {
    if (endpointInfo && modalVisible && startTime && endTime) {
      setLoading(true)
      getAlertDescendantEventApi({
        startTime,
        endTime,
        service: endpointInfo.service,
        endpoint: endpointInfo.endpoint,
        currentPage: pagination.pageIndex,
        pageSize: pagination.pageSize,
      })
        .then((res) => {
          setData(res?.alertEvents || [])
          setPagination({
            ...pagination,
            total: res?.pagination?.total || 0,
          })
        })
        .catch(() => {
          setData([])
          setPagination({
            ...pagination,
            total: 0,
          })
        })
        .finally(() => {
          setLoading(false)
        })
    }
  }, [endpointInfo, modalVisible, pagination.pageIndex, pagination.pageSize])
  const columns = [
    {
      title: '告警类型',
      accessor: 'group',
      customWidth: 150,
      Cell: ({ value }) => {
        return groupLabel[value]
      },
    },
    {
      title: '告警事件名',
      accessor: 'name',
      customWidth: 150,
    },
    {
      title: '最近告警时间',
      accessor: 'receivedTime',
      customWidth: 180,
      Cell: ({ value }) => {
        return convertUTCToBeijing(value)
      },
    },
    {
      title: '告警来源',
      accessor: 'source',
      customWidth: 150,
    },
    {
      title: '告警详情',
      accessor: 'detail',
      justifyContent: 'left',
      Cell: ({ value }) =>
        isJSONString(value) ? (
          <ReactJson
            src={JSON.parse(value || '')}
            theme="brewer"
            collapsed={true}
            displayDataTypes={false}
            style={{ width: '100%' }}
            enableClipboard={false}
          />
        ) : (
          value
        ),
    },
  ]

  const changePagination = (page, pageSize) => {
    setPagination({
      pageSize: pageSize,
      pageIndex: page,
      total: pagination.total,
    })
  }
  const tableProps = useMemo(() => {
    return {
      columns: columns,
      data: data,
      showBorder: false,
      loading: loading,
    }
  }, [data, loading])
  return (
    <>
      <Modal
        title={
          <span>
            {endpointInfo?.service} ({endpointInfo?.endpoint})告警列表
          </span>
        }
        open={modalVisible}
        onCancel={closeModal}
        destroyOnClose
        centered
        cancelText="关闭"
        width={1000}
        footer={(_, { CancelBtn }) => (
          <>
            <CancelBtn />
          </>
        )}
      >
        <div className="h-[500px] pb-5">
          <BasicTable {...tableProps} />
          <Pagination
            defaultCurrent={1}
            total={pagination.total}
            current={pagination.pageIndex}
            pageSize={pagination.pageSize}
            className="flex-shrink-0 flex-grow-0 p-2"
            align="end"
            onChange={changePagination}
          />
        </div>
      </Modal>
    </>
  )
}
export default EventListModal
