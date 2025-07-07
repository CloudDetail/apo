/**
 * Copyright 2025 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import { Button, Tooltip } from 'antd'
import React from 'react'
import { LuChevronRight, LuX } from 'react-icons/lu'
import { useSelector } from 'react-redux'
import { DatasourceType } from 'src/core/types/dataGroup'
import DatasourceIcon from './DatasourceIcon'
import { useTranslation } from 'react-i18next'

interface DatasourceTagProps {
  type: DatasourceType
  name: string
  id: string
  cluster?: string
  namespace?: string
  closable?: boolean
  onRemoveSelection?: (id: string) => void
  block?: boolean
  path: string[]
}

const getSelectionColor = (type: string, theme: string) => {
  if (theme === 'dark') {
    switch (type) {
      case 'cluster':
        return 'bg-blue-900/30 text-blue-200 hover:bg-blue-900/40'
      case 'namespace':
        return 'bg-amber-900/30 text-amber-200 hover:bg-amber-900/40'
      case 'service':
        return 'bg-emerald-900/30 text-emerald-200 hover:bg-emerald-900/40'
      default:
        return 'bg-gray-800/30 text-gray-300 border-gray-600/50 hover:bg-gray-800/40'
    }
  } else {
    switch (type) {
      case 'cluster':
        return 'border bg-blue-100 text-blue-800 border-blue-200'
      case 'namespace':
        return 'border bg-amber-100 text-amber-800 border-amber-200'
      case 'service':
        return 'border bg-green-100 text-green-800 border-green-200'
      default:
        return 'border bg-gray-100 text-gray-800 border-gray-200'
    }
  }
}

const renderTooltipTitle = (type: DatasourceType, name: string) => {
  const { t } = useTranslation('core/dataGroup')
  return (
    <div className="flex">
      <span className="text-var([--ant-color-text-secondary]) mr-1">
        {t(`datasourceType.${type}`)} :{' '}
      </span>
      {name && <div className={`truncate font-semibold`}>{name}</div>}
    </div>
  )
}

const DatasourceTag: React.FC<DatasourceTagProps> = ({
  type,
  id,
  name,
  path,
  closable = false,
  onRemoveSelection,
  block = true,
}) => {
  const theme = useSelector((state: any) => state.settingReducer.theme)

  const handleRemove = () => {
    onRemoveSelection?.(id)
  }

  return (
    <Tooltip title={!block && renderTooltipTitle(type, name)}>
      <div
        className={`group flex items-center justify-between px-2 py-1 m-1 rounded-lg text-xs transition-all duration-200 hover:shadow-sm ${
          block ? 'w-full' : 'inline-flex'
        } ${getSelectionColor(type, theme)}`}
      >
        <div className="flex items-center gap-1 flex-1 min-w-0">
          <div className="flex items-center gap-1 truncate">
            {!block && <DatasourceIcon type={type} />}
            {path?.slice(1).map((item, index) => {
              return (
                <div key={index} className={`truncate opacity-75 flex items-center gap-1`}>
                  {item} <LuChevronRight />
                </div>
              )
            })}
            <div className={`truncate font-semibold`}>{name}</div>
          </div>
        </div>
        {closable && (
          <Button
            type="text"
            onClick={handleRemove}
            className="p-0.5 rounded flex-shrink-0 ml-2"
            title="Remove"
          >
            <LuX />
          </Button>
        )}
      </div>
    </Tooltip>
  )
}

export default DatasourceTag
