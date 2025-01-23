/**
 * Copyright 2025 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import { Tag } from 'antd'
import { DatasourceType } from 'src/core/types/dataGroup'

interface DatasourceTagProps {
  type: DatasourceType
  datasource: string
  closable?: boolean
  onClose?: any
}
const DatasourceTag = ({ type, datasource, closable = false, onClose }: DatasourceTagProps) => {
  return (
    <Tag
      color={type === 'service' ? 'cyan' : 'geekblue'}
      closable={closable}
      onClose={(e) => {
        onClose(e, datasource, type)
      }}
    >
      {datasource}
    </Tag>
  )
}
export default DatasourceTag
