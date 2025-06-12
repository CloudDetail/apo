/**
 * Copyright 2025 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */
import { Tag } from 'antd'
import { useTranslation } from 'react-i18next'
import { AlertLevel } from '../AlertInfoCom'

const ValueTag = ({
  itemKey,
  display,
  value,
}: {
  itemKey: string
  display?: string
  value?: string
}) => {
  const { t } = useTranslation('oss/alertEvents')
  return itemKey === 'status' ? (
    <Tag color={value === 'firing' ? 'error' : 'success'} className="mr-1">
      {t(value)}
    </Tag>
  ) : itemKey === 'severity' ? (
    <AlertLevel level={value} />
  ) : (
    <Tag bordered={false} className="mr-1  text-[var(--ant-color-text)]">
      {display || t(value)}
    </Tag>
  )
}
export default ValueTag
