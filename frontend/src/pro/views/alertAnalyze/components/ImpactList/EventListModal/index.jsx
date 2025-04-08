import { Button, Modal, Pagination } from 'antd'
import React, { useEffect, useMemo, useState } from 'react'
import ReactJson from 'react-json-view'
import { useSelector } from 'react-redux'
import { getAlertDescendantEventApi } from 'src/pro/api/alertAnalyze'
import BasicTable from 'src/core/components/Table/basicTable'
import TempCell from 'src/core/components/Table/TempCell'
import { selectSecondsTimeRange } from 'src/core/store/reducers/timeRangeReducer'
import { convertUTCToLocal } from 'src/core/utils/time'
import { useTranslation } from 'react-i18next'
function isJSONString(str) {
  try {
    JSON.parse(str)
    return true
  } catch (e) {
    return false
  }
}
const i18nKey = 'eventList'
const EventListModal = ({ endpointInfo, modalVisible, closeModal }) => {
  const { t } = useTranslation('pro/alertAnalyze')
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
      title: t(`${i18nKey}.alertType`),
      accessor: 'group',
      customWidth: 150,
      Cell: ({ value }) => {
        return groupLabel[value]
      },
    },
    {
      title: t(`${i18nKey}.alertName`),
      accessor: 'name',
      customWidth: 150,
    },
    {
      title: t(`${i18nKey}.lastAlertTime`),
      accessor: 'receivedTime',
      customWidth: 180,
      Cell: ({ value }) => {
        return convertUTCToLocal(value)
      },
    },
    {
      title: t(`${i18nKey}.alertSource`),
      accessor: 'source',
      customWidth: 150,
    },
    {
      title: t(`${i18nKey}.alertDetail`),
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
            {t(`${i18nKey}.alertListFor`, {
              service: endpointInfo?.service,
              endpoint: endpointInfo?.endpoint,
            })}
          </span>
        }
        open={modalVisible}
        onCancel={closeModal}
        destroyOnClose
        centered
        cancelText={t(`${i18nKey}.close`)}
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
