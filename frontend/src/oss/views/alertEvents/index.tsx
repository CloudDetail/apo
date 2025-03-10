/**
 * Copyright 2025 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import { Button, Card, Modal, Pagination } from 'antd'
import { useEffect, useMemo, useState } from 'react'
import { useTranslation } from 'react-i18next'
import ReactJson from 'react-json-view'
import { useSelector } from 'react-redux'
import { useNavigate } from 'react-router-dom'
import { getAlertEventsApi } from 'src/core/api/alerts'
import BasicTable from 'src/core/components/Table/basicTable'
import { convertUTCToBeijing } from 'src/core/utils/time'
import WorkflowsIframe from '../workflows/workflowsIframe'
function isJSONString(str) {
  try {
    JSON.parse(str)
    return true
  } catch (e) {
    return false
  }
}
const AlertEventsPage = () => {
  const { t } = useTranslation('oss/alertEvents')
  const [pagination, setPagination] = useState({
    pageIndex: 1,
    pageSize: 10,
    total: 0,
  })
  const [alertEvents, setAlertEvents] = useState([])
  const { groupLabel } = useSelector((state) => state.groupLabelReducer)
  const { startTime, endTime } = useSelector((state) => state.timeRange)
  const [modalOpen, setModalOpen] = useState(false)
  const [workflowUrl, setWorkflowUrl] = useState(null)
  const [workflowId, setWorkflowId] = useState(null)
  const getAlertEvents = () => [
    getAlertEventsApi({
      startTime,
      endTime,
      pagination: {
        currentPage: pagination.pageIndex,
        pageSize: pagination.pageSize,
      },
    }).then((res) => {
      setAlertEvents(res?.events || [])
      setPagination({
        ...pagination,
        total: res?.pagination.total || 0,
      })
      setWorkflowId(res.alertEventAnalyzeWorkflowId)
    }),
  ]
  useEffect(() => {
    if (startTime && endTime) {
      getAlertEvents()
    }
  }, [pagination.pageIndex, startTime, endTime])
  function openWorkflowModal(workflowParams) {
    let result = '/dify/app/' + workflowId + '/run?'
    const params = Object.entries(workflowParams)
      .map(([key, value]) => `${key}=${encodeURIComponent(value)}`)
      .join('&')
    setWorkflowUrl(result + params)
    setModalOpen(true)
    // buildParams('workflowParams', workflowParams)
    // return paramsArray.join('&')
  }
  function openResultModal(workflowRunId) {
    let result = '/dify/app/' + workflowId + '/logs/' + workflowRunId
    setWorkflowUrl(result)
    setModalOpen(true)
  }
  const closeModal = () => {
    setWorkflowUrl(null)
    setModalOpen(false)
  }
  const columns = [
    {
      title: t('alertType'),
      accessor: 'group',
      customWidth: 150,
      Cell: ({ value }) => {
        return groupLabel[value]
      },
    },
    {
      title: t('alertName'),
      accessor: 'name',
      customWidth: 150,
    },
    {
      title: t('lastAlertTime'),
      accessor: 'receivedTime',
      customWidth: 180,
      Cell: ({ value }) => {
        return convertUTCToBeijing(value)
      },
    },
    {
      title: t('alertSource'),
      accessor: 'source',
      customWidth: 150,
    },
    {
      title: t('alertDetail'),
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
    {
      title: t('isValid'),
      accessor: 'isValid',
      customWidth: 200,
      Cell: (props) => {
        const { value, row } = props
        return value === 'unknown' ? (
          <span className="text-gray-400">{t(value)}</span>
        ) : (
          <Button
            type="link"
            onClick={() => {
              openResultModal(row.original.workflowRunId)
            }}
          >
            {t(value)}
          </Button>
        )
      },
    },
    {
      title: t('cause'),
      accessor: 'cause',
      customWidth: 200,
      Cell: (props) => {
        const { workflowParams } = props.row.original
        return (
          <Button
            type="link"
            onClick={() => {
              openWorkflowModal(workflowParams)
            }}
          >
            {t('viewWorkflow')}
          </Button>
        )
      },
    },
  ]
  const updatePagination = (pagination) => setPagination({ ...pagination, ...pagination })
  const changePagination = (page, pageSize) => {
    updatePagination({
      pageSize: pageSize,
      pageIndex: page,
      total: pagination.total,
    })
  }
  const tableProps = useMemo(() => {
    return {
      columns: columns,
      data: alertEvents,
      showBorder: false,
      loading: false,
    }
  }, [alertEvents])
  return (
    <>
      <Card
        style={{ height: 'calc(100vh - 60px)' }}
        styles={{
          body: {
            height: '100%',
            overflow: 'hidden',
            display: 'flex',
            flexDirection: 'column',
            padding: '12px 24px',
          },
        }}
      >
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
        <Modal
          open={modalOpen}
          title={t('workflowsModal')}
          onCancel={closeModal}
          destroyOnClose
          centered
          footer={() => <></>}
          maskClosable={false}
          width={1000}
          styles={{ body: { height: '80vh', overflowY: 'hidden', overflowX: 'hidden' } }}
        >
          {workflowUrl && <WorkflowsIframe src={workflowUrl} />}
        </Modal>
      </Card>
    </>
  )
}
export default AlertEventsPage
