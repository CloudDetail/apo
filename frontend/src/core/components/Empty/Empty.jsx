/**
 * Copyright 2024 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import { CImage } from '@coreui/react'
import React, { useEffect, useState } from 'react'
import emptyImg from 'src/core/assets/images/empty.svg'
import { useTranslation } from 'react-i18next'
function Empty({ context = '', width = 100 }) {
  const [stateContext, setStateContext] = useState('No data')
  const { i18n } = useTranslation()
  useEffect(() => {
    if (context) {
      setStateContext(context)
      return
    }
    setStateContext(i18n.language === 'en' ? 'No data' : '暂无数据')
  }, [i18n.language])
  return (
    <div className="w-full h-full flex flex-col justify-center items-center py-4 select-none">
      <CImage src={emptyImg} width={width} />
      {stateContext}
    </div>
  )
}
export default Empty
