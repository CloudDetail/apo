/**
 * Copyright 2024 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import React from 'react'
import { Spin } from 'antd'
import { LoadingOutlined } from '@ant-design/icons'

const LoadingSpinner = ({ loading, size = 'large' }) => {
  return (
    <>
      {loading ? (
        <div className=" absolute top-0 left-0 w-full h-full z-50 bg-[var(--mask-bg)] flex items-center justify-center">
          <Spin indicator={<LoadingOutlined spin />} size={size} />
        </div>
      ) : null}
    </>
  )
}
export default LoadingSpinner
