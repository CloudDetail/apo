/**
 * Copyright 2025 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import { Button, Card, Modal, Tooltip } from 'antd'
import { useEffect, useMemo, useRef, useState } from 'react'
import { useTranslation } from 'react-i18next'
import ReactJson from 'react-json-view'
import { useSelector } from 'react-redux'
import { getAlertEventsApi } from 'src/core/api/alerts'
import BasicTable from 'src/core/components/Table/basicTable'
import { convertUTCToBeijing } from 'src/core/utils/time'
import WorkflowsIframe from '../workflows/workflowsIframe'
import Tag from 'src/core/components/Tag/Tag'
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
  const [alertCheckId, setAlertCheckId] = useState(null)
  const timerRef = useRef(null)
  const workflowMissToast = (type: 'alertCheckId' | 'workflowId') => {
    return (
      <Tooltip title={type === 'alertCheckId' ? t('missToast1') : t('missToast2')}>
        <span className="text-gray-400 text-xs">{t('workflowMiss')}</span>
      </Tooltip>
    )
  }

  const getAlertEvents = async () => {
    const res = await getAlertEventsApi({
      startTime,
      endTime,
      pagination: {
        currentPage: pagination.pageIndex,
        pageSize: pagination.pageSize,
      },
    })
    const totalPages = Math.ceil(res.pagination.total / pagination.pageSize)
    if (pagination.pageIndex > totalPages && totalPages > 0) {
      setPagination({ ...pagination, pageIndex: totalPages })
      return
    }
    setAlertEvents(res?.events || [])
    setPagination({
      ...pagination,
      total: res?.pagination.total || 0,
    })
    setWorkflowId(res.alertEventAnalyzeWorkflowId)
    setAlertCheckId(res.alertCheckId)
      if (timerRef.current) {
        clearTimeout(timerRef.current)
      }

      timerRef.current = setTimeout(
        () => {
          getAlertEvents()
        },
        5 * 60 * 1000,
      )
  }

  useEffect(() => {
    if (startTime && endTime) {
      getAlertEvents()
    }
    return () => {
      if (timerRef.current) {
        clearTimeout(timerRef.current)
      }
    }
  }, [pagination.pageIndex, pagination.pageSize, startTime, endTime])
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
    let result = '/dify/app/' + alertCheckId + '/logs/' + workflowRunId
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
      title: t('status'),
      accessor: 'status',
      customWidth: 120,
      Cell: ({ value }) => {
        return <Tag type={value === 'firing' ? 'error' : 'success'}>{t(value)}</Tag>
      },
    },
    {
      title: t('isValid'),
      accessor: 'isValid',
      customWidth: 160,
      Cell: (props) => {
        const { value, row } = props
        return !alertCheckId ? (
          workflowMissToast('alertCheckId')
        ) : ['unknown', 'skipped'].includes(value) ||
          (value === 'failed' && !row.original.workflowRunId) ? (
          <span className="text-gray-400 text-xs text-wrap [word-break:auto-phrase] text-center">
            {t(value)}
          </span>
        ) : (
          <Button
            type="link"
            className="text-xs text-wrap [word-break:auto-phrase] "
            size="small"
            onClick={() => {
              openResultModal(row.original.workflowRunId)
            }}
          >
            {t(value === 'failed' ? 'failedTo' : value)}
          </Button>
        )
      },
    },
    {
      title: <>{t('cause')}</>,
      accessor: 'cause',
      customWidth: 160,
      Cell: (props) => {
        const { workflowParams } = props.row.original
        return !workflowId ? (
          workflowMissToast('workflowId')
        ) : (
          <Button
            type="link"
            className="text-xs"
            size="small"
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
  const changePagination = (page, pageSize) => {
    setPagination({ ...pagination, pageIndex: page, pageSize })
  }
  const tableProps = useMemo(() => {
    return {
      columns: columns,
      data: alertEvents,
      showBorder: false,
      loading: false,
      onChange: changePagination,
      pagination: {
        pageSize: pagination.pageSize,
        pageIndex: pagination.pageIndex,
        total: pagination.total,
      },
    }
  }, [alertEvents, pagination.pageIndex, pagination.pageSize, pagination.total])
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
            fontSize: '12px',
          },
        }}
      >
        <BasicTable {...tableProps} />
        <Modal
          open={modalOpen}
          title={t('workflowsModal')}
          onCancel={closeModal}
          destroyOnClose
          centered
          footer={() => <></>}
          maskClosable={false}
          width={'80vw'}
          styles={{ body: { height: '80vh', overflowY: 'hidden', overflowX: 'hidden' } }}
        >
          {workflowUrl && <WorkflowsIframe src={workflowUrl} />}
        </Modal>
      </Card>
    </>
  )
}
export default AlertEventsPage
