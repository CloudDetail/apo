/**
 * Copyright 2024 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import React, { useCallback } from 'react'
import { Handle, Position } from 'reactflow'
import { BiCctv } from 'react-icons/bi'
import { MdRemoveRedEye } from 'react-icons/md'
import { useTranslation } from 'react-i18next'
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
      <div
        className="px-3 py-2 rounded-full border-2 border-solid bg-[var(--ant-color-fill-secondary)] border-[var(--ant-color-primary-border)] text-[var(--ant-color-primary-text)] overflow-hidden flex flex-row items-center justify-center"
        // style={{ backgroundColor: 'rgba(19, 25, 32, 0.6)' }}
      >
        <MdRemoveRedEye className="mr-2" />
        {t('moreNode.seeMore')}
        {/* <div className="absolute top-0 left-0">
          {' '}
          <BiCctv size={35} color={data.isTraced ? '#80ce8dff' : '#a1a1a1'} />
        </div>
        <div className="text-center text-lg pt-2 px-2">
          {data.label}
          <div className="text-xs text-[#9e9e9e] text-left">{data.endpoint}</div>
        </div> */}
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

export default MoreNode
