/**
 * Copyright 2025 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import { Image } from 'antd'
import { DatasourceItemData } from '../types'

interface DatasourceItemProps extends DatasourceItemData {
  size: 'small' | 'normal'
}
const DatasourceItem = ({ src, name, description, size = 'normal' }: DatasourceItemProps) => {
  return (
    <div
      className="relative  bg-[#2a3042]  pt-3 rounded-lg cursor-pointer "
      style={{ height: size === 'normal' ? 120 : 60, width: size === 'normal' ? 140 : 100 }}
    >
      <div className="flex flex-col items-center justify-between h-full transition-transform duration-300 hover:scale-110">
        <div
          className=" flex items-center justify-center"
          style={{ height: size === 'normal' ? 60 : 30, width: size === 'normal' ? 90 : 30 }}
        >
          <Image height={size === 'normal' ? 40 : 20} src={src} preview={false} />
        </div>
        <div
          className="font-bold text-sm text-gray-300  flex items-center flex-col"
          style={{ height: size === 'normal' ? 40 : 40, fontSize: size === 'normal' ? 14 : 10 }}
        >
          <span className="text-center">{name}</span>
          <div className="text-xs text-gray-400">{description}</div>
        </div>
      </div>
    </div>
  )
}

export default DatasourceItem
