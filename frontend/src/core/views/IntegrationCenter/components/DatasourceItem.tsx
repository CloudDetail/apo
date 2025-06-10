/**
 * Copyright 2025 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import { DatasourceItemData } from '../types'
interface DatasourceItemProps extends DatasourceItemData {
  size: 'small' | 'normal'
}
import { ReactSVG } from 'react-svg'
const DatasourceIcon = ({ src, height }: { src: string; height?: string }) => {
  if (src.includes('.svg')) {
    return (
      <ReactSVG
        src={src}
        width={'100%'}
        beforeInjection={(svg) => {
          svg.setAttribute('width', '100%')
          svg.setAttribute('height', height)
        }}
      />
    )
  }

  // fallback: PNG, JPG, etc.
  return <img src={src} style={{ height: height }} alt="icon" />
}
const DatasourceItem = ({ src, name, description, size = 'normal' }: DatasourceItemProps) => {
  return (
    <div
      className="relative  bg-[var(--ant-color-bg-layout)]  pt-3 rounded-lg cursor-pointer "
      style={{ height: size === 'normal' ? 120 : 60, width: size === 'normal' ? 140 : 100 }}
    >
      <div className="flex flex-col items-center justify-between h-full transition-transform duration-300 hover:scale-110">
        <div
          className=" flex items-center justify-center"
          style={{
            height: size === 'normal' ? 60 : 30,
            width: size === 'normal' ? 90 : 30,
          }}
        >
          <DatasourceIcon src={src} height={size === 'normal' ? 40 : 20} />
        </div>
        <div
          className="font-bold text-sm text-[var(--ant-color-text-secondary)] flex items-center flex-col"
          style={{ height: size === 'normal' ? 40 : 40, fontSize: size === 'normal' ? 14 : 10 }}
        >
          <span className="text-center">{name}</span>
          <div className="text-xs text-[var(--ant-color-text-secondary)]">{description}</div>
        </div>
      </div>
    </div>
  )
}

export default DatasourceItem
