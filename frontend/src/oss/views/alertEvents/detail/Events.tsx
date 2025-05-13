/**
 * Copyright 2025 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */
import { useTranslation } from 'react-i18next'
import { Modal } from 'antd'
import { useMemo, useState } from 'react'
import { convertUTCToLocal } from 'src/core/utils/time'
import BasicTable from 'src/core/components/Table/basicTable'
import { AlertTags, ALertIsValid, AlertStatus } from '../components/AlertInfoCom'
import WorkflowsIframe from '../../workflows/workflowsIframe'
import { FaLocationDot } from 'react-icons/fa6'

interface EventsProps {
  alertCheckId: string | null
  eventId?: string | null
  workflowId?: string | null
  alertEvents: any[]
  pageSize: number
  pageIndex: number
  total: number
  changePagination: any
}
const Events = ({
  eventId,
  alertCheckId,
  alertEvents = [],
  pageSize,
  pageIndex,
  total,
  changePagination,
}: EventsProps) => {
  const [modalOpen, setModalOpen] = useState(false)
  const [workflowUrl, setWorkflowUrl] = useState(null)
  const { t } = useTranslation('oss/alertEvents')
  function openResultModal(workflowRunId) {
    let result = '/dify/app/' + alertCheckId + '/logs/' + workflowRunId
    setWorkflowUrl(result)
    setModalOpen(true)
  }
  const closeModal = () => {
    setWorkflowUrl(null)
    setModalOpen(false)
  }
  const tdStyle = {
    padding: 8,
  }
  const columns = [
    {
      title: t('recordTime'),
      accessor: 'updateTime',
      style: {
        borderBottom: 0,
      },
      customWidth: 250,
      Cell: ({ value, row }) => {
        const { id, status, endTime } = row.original
        const result = convertUTCToLocal(status === 'firing' ? value : endTime)
        const style = {
          width: 1,
          height: '100%',
          borderLeft: '2px solid rgba(253, 253, 253, 0.12)',
          flexGrow: 1,
        }

        return (
          <div
            id={id}
            className=" flex h-full items-center w-full text-xs "
            style={
              eventId === id
                ? {
                    background: 'rgba(93, 135, 255, 0.2)',
                    color: '#5d87ff',
                    borderLeft: '5px solid #5d87ff',
                    borderRadius: '4px 20px 20px 4px',
                  }
                : {
                    borderLeft: '5px solid transparent',
                  }
            }
          >
            <div className="flex-1 text-right font-bold">{result}</div>
            <div className="flex flex-col items-center h-full">
              <div className="grow ">
                <div style={style}></div>
              </div>
              {eventId === id ? (
                <FaLocationDot className="w-4 h-4 m-2" color="#1668dc" size={24} />
              ) : (
                <div
                  style={{
                    border: '3px solid #1668dc',
                    borderColor: status === 'firing' ? '#ff4d4f' : '#52c41a',
                    borderRadius: '50%',
                  }}
                  className="w-3 h-3 m-2"
                ></div>
              )}

              <div className="grow">
                <div style={style}></div>
              </div>
            </div>

            <div className="flex-1">
              <AlertStatus
                status={row.original.status}
                // resolvedTime={convertUTCToLocal(row.original.endTime)}
              />
            </div>
          </div>
        )
      },
    },
    {
      title: t('alertDetail'),
      accessor: 'tags',
      style: tdStyle,

      justifyContent: 'left',
      Cell: ({ value, row }) => {
        return <AlertTags tags={value} detail={row.original.detail} />
      },
    },
    {
      title: t('notifyStatus'),
      accessor: 'notifyAt',
      style: tdStyle,
      customWidth: 200,
      Cell: ({ value, row }) => {
        const { notifyFailed, notifySuccess } = row.original
        const notifyAt = convertUTCToLocal(value)
        return (
          <>
            {notifyFailed ? (
              <span className="text-[#E84749]">{t('notifyFailed')}</span>
            ) : !notifySuccess ? (
              <span className="text-gray-400">{t('unNotify')}</span>
            ) : (
              <>
                <span className="">
                  {' '}
                  {t('notifyAt')} {notifyAt}
                </span>
              </>
            )}
          </>
        )
      },
    },
    {
      title: t('isValid'),
      accessor: 'isValid',
      style: tdStyle,

      customWidth: 160,
      Cell: (props) => {
        const { value, row } = props
        const checkTime = convertUTCToLocal(row.original.lastCheckAt)

        return (
          <ALertIsValid
            isValid={value}
            alertCheckId={alertCheckId}
            checkTime={checkTime}
            openResultModal={() => openResultModal(row.original.workflowRunId)}
          />
        )
      },
    },
  ]
  const tableProps = useMemo(() => {
    return {
      columns: columns,
      data: alertEvents,
      showBorder: false,
      loading: false,
      pagination: {
        pageSize: pageSize,
        pageIndex: pageIndex,
        total: total,
      },
      tdPadding: 0,
      onChange: changePagination,
    }
  }, [alertEvents, pageIndex, pageSize, total])
  return (
    <div className="no-tr-bottom h-0 grow">
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
    </div>
  )
}
export default Events
