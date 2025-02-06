/**
 * Copyright 2025 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import { Image } from 'antd'
import { DatasourceItemData } from '../types'

const DatasourceItem = ({ src, name, description }: DatasourceItemData) => {
  return (
    <div className="relative h-[120px] bg-[#2a3042]  pt-3 rounded-lg cursor-pointer w-[140px]">
      <div className="flex flex-col items-center justify-between h-full transition-transform duration-300 hover:scale-110">
        <div className="h-[60px] w-[90px] flex items-center justify-center">
          <Image height={40} src={src} preview={false} />
        </div>
        <div className="h-[40px] font-bold text-sm text-gray-300 px-2 flex items-center flex-col">
          <span className="text-center">{name}</span>
          <div className="text-xs text-gray-400">{description}</div>
        </div>
      </div>
    </div>
  )
}

export default DatasourceItem
