/**
 * Copyright 2025 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import React, { CSSProperties, ReactElement } from 'react'
import { SLOT_TYPES, CardTable, CardHeader } from './CardSlots'
import { Card } from 'antd'
import style from './index.module.scss'
type CardProps = {
  children: React.ReactNode
  bodyStyle?: CSSProperties
}

export const BasicCard: React.FC<CardProps> & {
  Header: typeof CardHeader
  Table: typeof CardTable
} = ({ children, bodyStyle }) => {
  let headerContent: ReactElement[] = []
  let tableContent: ReactElement | null = null
  const otherContent: ReactElement[] = []

  //@ts-ignore
  const contentHeight =
    import.meta.env.VITE_APP_CODE_VERSION === 'CE'
      ? 'var(--ce-app-content-height)'
      : 'var(--ee-app-content-height)'

  React.Children.forEach(children, (child) => {
    if (!React.isValidElement(child)) return

    const slotType = (child.type as any)?.slotType

    switch (slotType) {
      case SLOT_TYPES.HEADER:
        headerContent.push(child)
        break
      case SLOT_TYPES.TABLE:
        tableContent = child
        break
      default:
        otherContent.push(child)
        break
    }
  })

  return (
    <div className={style.basicCard}>
      <Card
        styles={{
          body: {
            // height: contentHeight,
            height: '100%',
            overflow: 'hidden',
            display: 'flex',
            flexDirection: 'column',
            padding: '12px 24px',
            ...bodyStyle,
          },
        }}
      >
        {/* Header Section */}
        {headerContent.length > 0 &&
          headerContent.map((header, index) => (
            <div
              className="w-full text-sm font-medium flex items-center justify-between"
              key={index}
            >
              {header}
            </div>
          ))}

        {/* Table Section */}
        {tableContent && (
          <div className="flex-1 overflow-auto">
            <div className="h-full text-xs justify-between">
              {tableContent && <>{tableContent}</>}
            </div>
          </div>
        )}

        {/* Other Content */}
        {otherContent.length > 0 &&
          otherContent.map((content, index) => (
            <React.Fragment key={index}>{content}</React.Fragment>
          ))}
      </Card>
    </div>
  )
}

BasicCard.Header = CardHeader
BasicCard.Table = CardTable
