/**
 * Copyright 2024 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import React, { useCallback } from 'react'
import { Handle, Position } from 'reactflow'
import { BiCctv } from 'react-icons/bi'
import { theme, Tooltip } from 'antd'
import { LuShieldOff } from 'react-icons/lu'
import { useTranslation } from 'react-i18next'
const handleStyle = { left: 10 }
const ServiceNode = React.memo(({ data, isConnectable }) => {
  const onChange = useCallback((evt) => {}, [])
  const { t } = useTranslation('common')
  const { useToken } = theme
  const { token } = useToken()
  return (
    <div className="text-updater-node">
      <Handle
        type="target"
        position={Position.Left}
        isConnectable={isConnectable}
        className="invisible"
      />
      <Tooltip title={data.outOfGroup ? t('outOfGroup') : ''}>
        <div
          className="w-[200px] min-h-[60px] p-2 rounded-md border-2 border-solid overflow-hidden"
          style={{
            backgroundColor: token.colorBgLayout,
            color: data?.outOfGroup ? token.colorTextDisabled : token.colorPrimaryText,
            borderColor: data?.outOfGroup ? token.colorTextDisabled : token.colorPrimaryText,
          }}
        >
          <div className="absolute top-0 left-0">
            {' '}
            <BiCctv size={35} color={data.isTraced ? '#80ce8dff' : '#a1a1a1'} />
          </div>
          {data.outOfGroup && (
            <div className="absolute top-0 right-0">
              <LuShieldOff size={35} />
            </div>
          )}
          <div className="text-center text-lg pt-2 px-2">
            {data.label}
            <div className="text-xs text-[#9e9e9e] text-left  break-all">{data.endpoint}</div>
          </div>
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

export default ServiceNode
