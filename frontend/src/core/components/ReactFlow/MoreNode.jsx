/**
 * Copyright 2024 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import React, { useCallback } from 'react'
import { Handle, Position } from 'reactflow'
import { MdRemoveRedEye } from 'react-icons/md'
import { useTranslation } from 'react-i18next'
import { LuShieldOff } from 'react-icons/lu'
import { Tooltip } from 'antd'
const handleStyle = { left: 10 }
const MoreNode = React.memo((prop) => {
  const { data, isConnectable } = prop
  const onChange = useCallback((evt) => {}, [])
  const { t } = useTranslation('oss/serviceInfo')
  return (
    <div className="text-updater-node cursor-pointer">
      <Handle
        type="target"
        position={Position.Left}
        isConnectable={isConnectable}
        className="invisible"
      />
      <Tooltip title={data.disabled ? t('moreNode.outOfGroup') : ''}>
        <div
          className={`px-3 py-2 rounded-full border-1 border-solid bg-[var(--ant-color-fill-secondary)]  text-[var(--ant-color-primary-text)] overflow-hidden
             flex flex-row items-center justify-center ${data.disabled ? 'text-[var(--ant-color-text-secondary)] border-[var(--ant-color-text-secondary)]' : 'text-[var(--ant-color-primary-text)] border-[var(--ant-color-primary-border)]'}`}
          style={{
            cursor: data.disabled ? 'not-allowed' : 'pointer',
          }}
        >
          {data.disabled ? (
            <>
              <LuShieldOff className="mr-2" />
              {t('moreNode.more')}
            </>
          ) : (
            <>
              <MdRemoveRedEye className="mr-2" />
              {t('moreNode.seeMore')}
            </>
          )}
        </div>
      </Tooltip>

      <Handle
        type="source"
        position={Position.Right}
        id="b"
        isConnectable={isConnectable}
        className="invisible"
      />
    </div>
  )
})

export default MoreNode
