/**
 * Copyright 2024 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import React, { useCallback } from 'react'
import { Handle, Position } from 'reactflow'
import { BiCctv } from 'react-icons/bi'
import { theme } from 'antd'
const handleStyle = { left: 10 }
const ServiceNode = React.memo(({ data, isConnectable }) => {
  const onChange = useCallback((evt) => {}, [])
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
      <div
        className="w-[200px] min-h-[60px] p-2 rounded-md border-2 border-solid overflow-hidden"
        style={{
          backgroundColor: token.colorBgLayout,
          color: token.colorPrimaryText,
          borderColor: token.colorPrimaryText,
        }}
      >
        <div className="absolute top-0 left-0">
          {' '}
          <BiCctv size={35} color={data.isTraced ? '#80ce8dff' : '#a1a1a1'} />
        </div>
        <div className="text-center text-lg pt-2 px-2">
          {data.label}
          <div className="text-xs text-[#9e9e9e] text-left  break-all">{data.endpoint}</div>
        </div>
      </div>
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
