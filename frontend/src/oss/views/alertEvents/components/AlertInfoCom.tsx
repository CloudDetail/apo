/**
 * Copyright 2025 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */
import { Button, Tag, Tooltip } from 'antd'
import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import ReactJson from 'react-json-view'
import { t } from 'i18next'
function isJSONString(str: string) {
  try {
    JSON.parse(str)
    return true
  } catch {
    return false
  }
}
interface Tag {
  [key: string]: string
}
interface AlertTagsProps {
  tags: Tag
  detail: string
  defaultVisible?: boolean
}
const AlertTags = ({ tags, detail, defaultVisible = false }: AlertTagsProps) => {
  const { t } = useTranslation('oss/alertEvents')
  const [visible, setVisible] = useState(false)

  return (
    <div className="overflow-hidden text-xs">
      {Object.entries(tags || {}).map(([key, tagValue]) => (
        <Tag className="text-pretty mb-1 break-all">
          <span>
            {key} = {tagValue}
          </span>
        </Tag>
      ))}

      {isJSONString(detail) && !defaultVisible && (
        <Button color="primary" variant="text" size="small" onClick={() => setVisible(!visible)}>
          {visible ? t('oss/alertEvents:collapse') : t('oss/alertEvents:expand')}
        </Button>
      )}

      {(visible || defaultVisible) && isJSONString(detail) && (
        <ReactJson
          src={JSON.parse(detail || '')}
          theme="brewer"
          collapsed={false}
          displayDataTypes={false}
          style={{ width: '100%' }}
          enableClipboard={false}
        />
      )}
    </div>
  )
}
const AlertDeration = ({
  duration,
  updateTime,
}: {
  duration: string
  updateTime?: string | null
}) => {
  const { t } = useTranslation('oss/alertEvents')
  return (
    <div>
      {duration}
      {updateTime && (
        <span className="text-[10px] block text-gray-400">
          {t('oss/alertEvents:updateTime')} {updateTime}
        </span>
      )}
    </div>
  )
}
const AlertStatus = ({
  status,
  resolvedTime,
}: {
  status: string
  resolvedTime?: string | null
}) => {
  if (!status) return
  return (
    <div className="text-center">
      <Tag color={status === 'firing' ? 'error' : 'success'}>{t(`oss/alertEvents:${status}`)}</Tag>
      {status === 'resolved' && resolvedTime && (
        <span className="text-[10px] block text-gray-400">
          {t('oss/alertEvents:resolvedOn')} {resolvedTime}
        </span>
      )}
    </div>
  )
}
const workflowMissToast = (type: 'alertCheckId' | 'workflowId') => {
  return (
    <Tooltip
      title={
        type === 'alertCheckId' ? t('oss/alertEvents:missToast1') : t('oss/alertEvents:missToast2')
      }
    >
      <div>
        <span className="text-gray-400 text-xs">{t('oss/alertEvents:workflowMiss')}</span>
      </div>
    </Tooltip>
  )
}
const ALertIsValid = ({
  isValid,
  alertCheckId,
  checkTime,
  openResultModal,
  workflowRunId,
}: {
  isValid: 'unknown' | 'skipped' | 'invalid' | 'valid' | 'failed'
  alertCheckId?: string | null
  workflowRunId?: string | null
  checkTime?: string | null
  openResultModal: any
}) => {
  return (
    <>
      {!alertCheckId ? (
        workflowMissToast('alertCheckId')
      ) : ['unknown', 'skipped'].includes(isValid) || (isValid === 'failed' && !workflowRunId) ? (
        <span className="text-gray-400 text-wrap [word-break:auto-phrase] text-center flex items-center">
          {t(`oss/alertEvents:${isValid}`)}
        </span>
      ) : (
        <div className="text-center">
          <Button
            type="link"
            className="text-xs text-wrap [word-break:auto-phrase] "
            size="small"
            onClick={() => {
              openResultModal()
            }}
          >
            {t(`oss/alertEvents:${isValid === 'failed' ? 'failedTo' : isValid}`)}
          </Button>
          {checkTime && (
            <span className="text-[10px] block text-gray-400">
              {t('oss/alertEvents:checkedOn')} {checkTime}
            </span>
          )}
        </div>
      )}
    </>
  )
}
export { AlertTags, AlertDeration, AlertStatus, ALertIsValid }
