/**
 * Copyright 2024 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import React from 'react'
import { CSpinner } from '@coreui/react'

const LoadingSpinner = ({ loading, size = null }) => {
  return (
    <>
      {loading ? (
        <div className=" absolute top-0 left-0 w-full h-full z-10 fade show backdrop-brightness-50 backdrop-opacity-40 flex items-center justify-center">
          <CSpinner size={size} />
        </div>
      ) : null}
    </>
  )
}
export default LoadingSpinner
