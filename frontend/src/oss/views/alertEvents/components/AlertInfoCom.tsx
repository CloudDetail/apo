/**
 * Copyright 2025 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */
import { Button, Tag, theme, Tooltip } from 'antd'
import React, { FC, useState } from 'react'
import { useTranslation } from 'react-i18next'
import ReactJson from 'react-json-view'
import { t } from 'i18next'
import { useSelector } from 'react-redux'
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

const AlertDetail = ({ detail }) => {
  // Check if detail is an object
  const isObject = typeof detail === 'object' && detail !== null;

  return (
    <div className="bg-[var(--ant-color-fill-tertiary)] p-2 rounded text-xs overflow-auto break-all space-y-3">
      {isObject ? (
        Object.entries(detail).map(([key, value], index) => {
          // Handle the description field
          if (key === 'description') {
            const descriptionLines = value.split('\n');

            return (
              <div key={index}>
                <p className="font-semibold">{key.toUpperCase()}:</p>
                {descriptionLines
                  .filter(line => line.trim() !== '')
                  .map((line, idx) => {
                    // Detect and parse LABELS lines
                    if (line.trim().startsWith('LABELS = map[')) {
                      const blankSpace = line.match(/^\s+/);
                      const prefixBlankSpace = blankSpace ? blankSpace[0] : '';
                      const regex = /(\w+):(.*?)(?=\s+\w+:|$)/g;

                      // remove 'LABELS = map[' and ']'
                      const lineReady = line.slice('LABELS = map['.length, -1)

                      let match;
                      const labels = [];

                      while ((match = regex.exec(lineReady)) !== null) {
                        labels.push({ key: match[1], value: match[2] });
                      }

                      return (
                        <div key={idx}>
                          <p>{prefixBlankSpace}LABELS =</p>
                          <ul className={`list-disc pl-4 mb-0 ml-${prefixBlankSpace.length}`}>
                            {labels.map((label, labelIndex) => (
                              <li key={labelIndex}>
                                <span>{label.key}:</span> {label.value}
                              </li>
                            ))}
                          </ul>
                        </div>
                      );
                    }

                    // Render other lines
                    return (
                      <p key={idx} className="mb-1">
                        {line}
                      </p>
                    );
                  })}
              </div>
            );
          }

          // Render other fields
          return (
            <div key={index}>
              <span className="font-semibold">{key.toUpperCase()}:</span>{' '}
              {typeof value === 'object' ? (
                <pre className="whitespace-pre-wrap break-all">{JSON.stringify(value, null, 2)}</pre>
              ) : (
                <span>{value}</span>
              )}
            </div>
          );
        })
      ) : (
        // If detail is not an object, render it as plain text
        <p className="whitespace-pre-wrap break-all">{detail}</p>
      )}
    </div>
  );
};


const AlertTags = ({ tags, detail, defaultVisible = false }: AlertTagsProps) => {
  const { t } = useTranslation('oss/alertEvents')
  const { reactJsonTheme } = useSelector((state) => state.settingReducer)
  const [visible, setVisible] = useState(false)

  return (
    <div className="overflow-hidden text-xs">
      {Object.entries(tags || {}).map(([key, tagValue]) => (
        <Tag className="text-pretty mb-1 break-all" key={key}>
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
        <AlertDetail detail={JSON.parse(detail || '')} />
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
  const { useToken } = theme
  const { token } = useToken()

  return (
    <div>
      {duration}
      {updateTime && (
        <span className="text-[10px] block" style={{ color: token.colorTextSecondary }}>
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
        <span className="text-[10px] block text-[var(--ant-color-text-secondary)]">
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
        <span className="text-[var(--ant-color-text-secondary)] text-xs">
          {t('oss/alertEvents:workflowMiss')}
        </span>
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
  const { useToken } = theme
  const { token } = useToken()

  const statusColors = {
    valid: {
      color: token.colorSuccess,
      backgroundColor: token.colorSuccessBg,
    },
    invalid: {
      color: token.colorWarning,
      backgroundColor: token.colorWarningBg,
    },
    failed: {
      color: token.colorError,
      backgroundColor: token.colorErrorBg,
    },
    unknown: {
      color: token.colorTextSecondary,
      backgroundColor: 'transparent',
    },
    skipped: {
      color: token.colorTextSecondary,
      backgroundColor: 'transparent',
    },
  }

  const currentColors = statusColors[isValid]

  return (
    <>
      {!alertCheckId ? (
        workflowMissToast('alertCheckId')
      ) : ['unknown', 'skipped'].includes(isValid) || (isValid === 'failed' && !workflowRunId) ? (
        <span
          className="text-wrap [word-break:auto-phrase] text-center flex items-center text-xs"
          style={{ color: currentColors?.color }}
        >
          {t(`oss/alertEvents:${isValid}`)}
        </span>
      ) : (
        <div className="text-center">
          <Button
            type="link"
            className="text-xs text-wrap [word-break:auto-phrase]"
            size="small"
            onClick={() => openResultModal()}
            style={{
              color: currentColors?.color,
              backgroundColor: currentColors?.backgroundColor,
            }}
          >
            {t(`oss/alertEvents:${isValid === 'failed' ? 'failedTo' : isValid}`)}
          </Button>
          {checkTime && (
            <span className="text-[10px] block" style={{ color: token.colorTextSecondary }}>
              {t('oss/alertEvents:checkedOn')} {checkTime}
            </span>
          )}
        </div>
      )}
    </>
  )
}

const levelColorMap: Record<string, string> = {
  critical: 'red',
  error: 'volcano',
  warning: 'orange',
  info: 'blue',
  unknown: 'default'
};

const AlertLevel = ({ level }: { level: string }) => {
  const { t } = useTranslation('oss/alertEvents')

  const normalizedLevel = level.toLowerCase();
  const color = levelColorMap[normalizedLevel] || levelColorMap['unknown'];
  const label = normalizedLevel in levelColorMap ? normalizedLevel : 'unknown';

  return <Tag className='font-normal' bordered={false} color={color}>{t(label)}</Tag>;
};

export { AlertTags, AlertDeration, AlertStatus, ALertIsValid, AlertLevel }
