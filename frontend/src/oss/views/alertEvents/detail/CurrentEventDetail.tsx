/**
 * Copyright 2025 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */
import { Button, Descriptions, DescriptionsProps, Modal, Result, theme } from 'antd'
import { useTranslation } from 'react-i18next'
import {
  AlertDeration,
  ALertIsValid,
  AlertLevel,
  AlertStatus,
  AlertTags,
} from '../components/AlertInfoCom'
import { convertUTCToLocal } from 'src/core/utils/time'
import WorkflowsIframe from '../../workflows/workflowsIframe'
import { useState } from 'react'
import { FaEye } from 'react-icons/fa'
import { getAlertWorkflowIdApi } from 'src/core/api/alerts'
import LoadingSpinner from 'src/core/components/Spinner'
const CurrentEventDetail = ({
  detail,
  alertCheckId,
}: {
  detail: any
  alertCheckId?: string | null
}) => {
  const { t } = useTranslation('oss/alertEvents')
  const [modalOpen, setModalOpen] = useState(false)
  const [workflowUrl, setWorkflowUrl] = useState(null)
  const [loading, setLoading] = useState(true)
  const { useToken } = theme
  const { token } = useToken()
  const closeModal = () => {
    setModalOpen(false)
  }
  const items: DescriptionsProps['items'] = [
    {
      key: '1',
      span: 2,
      label: t('alertName'),
      children: detail?.name,
    },
    {
      key: '3',
      label: t('createTime'),
      children: detail?.createTime && convertUTCToLocal(detail?.createTime),
    },
    {
      key: 'recordTime',
      label: t('recordTime'),
      children:
        detail?.status &&
        convertUTCToLocal(detail?.status === 'firing' ? detail?.updateTime : detail?.endTime),
    },
    {
      key: '4',
      label: t('duration'),
      children: <AlertDeration duration={detail?.duration} />,
    },

    {
      key: '5',
      label: t('currentStatus'),
      span: 2,
      children: (
        <AlertStatus
          status={detail?.status}
          // resolvedTime={convertUTCToLocal(detail?.endTime)}
        />
      ),
    },
    {
      key: '2',
      label: t('severity'),
      children: <AlertLevel level={detail?.severity} />,
    },
    {
      key: '5',
      label: t('isValid'),
      span: 2,
      children: detail && (
        <ALertIsValid
          alertCheckId={alertCheckId}
          isValid={detail?.validity}
          // checkTime={convertUTCToLocal(detail?.lastCheckAt)}
          openResultModal={() => openResultModal(detail.workflowRunId)}
          workflowRunId={detail.workflowRunId}
        />
      ),
    },
    {
      key: 'detail',
      label: t('alertDetail'),
      span: 4,
      children: <AlertTags tags={detail?.tagsDisplay} detail={detail?.detail} defaultVisible />,
    },
  ]
  async function getWorkflowId(alertGroup, alertName) {
    try {
      const res = await getAlertWorkflowIdApi({ alertGroup, alertName })
      return res?.workflowId
    } catch (error) {
      return null
    }
  }
  function openResultModal(workflowRunId) {
    let result = '/dify/app/' + alertCheckId + '/logs/' + workflowRunId
    setWorkflowUrl(result)
    setModalOpen(true)
  }
  async function openWorkflowModal() {
    try {
      setLoading(true)
      setModalOpen(true)

      const workflowId = await getWorkflowId(detail.group, detail.name)
      if (!workflowId) {
        throw new Error()
      }
      let result = '/dify/app/' + workflowId + '/run-once?'
      const params = Object.entries(detail.workflowParams)
        .map(([key, value]) => `${key}=${encodeURIComponent(value)}`)
        .join('&')
      setWorkflowUrl(result + params)
    } catch {
      setLoading(false)
      return
    } finally {
      setLoading(false)
    }
  }
  return (
    <div
      className="w-full rounded-xl  h-full text-sm p-2"
      style={{ backgroundColor: token.colorBgContainer }}
    >
      <div className="flex flex-col h-full justify-between">
        <div className="flex-1 h-0 flex flex-col">
          <div className="text-base font-bold ">{t('alertEventDetail')}</div>
          <Descriptions
            items={items}
            size="small"
            className="overflow-auto items-center"
            column={5}
          />
        </div>
        <div className="w-full text-right grow-0 flex items-center justify-end overflow-auto">
          <Button
            color="primary"
            variant="outlined"
            className="ml-2"
            classNames={{ icon: 'flex items-center' }}
            icon={<FaEye />}
            onClick={async () => {
              await openWorkflowModal()
            }}
          >
            {t('viewWorkflow')}
          </Button>
        </div>
      </div>

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
        {!loading && !workflowUrl && (
          <Result
            status="error"
            title={t('missToast2')}
            className="h-full flex flex-col items-center justify-center w-full"
          />
        )}
        <LoadingSpinner loading={loading} />
        {workflowUrl && <WorkflowsIframe src={workflowUrl} />}
      </Modal>
    </div>
  )
}
export default CurrentEventDetail
