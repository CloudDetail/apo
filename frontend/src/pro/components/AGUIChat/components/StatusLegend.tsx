import React from 'react'
import { useTranslation } from 'react-i18next'
import StatusIcon, { type Status, useStatusLabel } from './StatusIcon'

export interface StatusLegendProps {
  /** 要显示的状态列表，如果不提供则显示所有状态 */
  statuses?: Status[]
  /** 是否垂直布局 */
  vertical?: boolean
  /** 自定义类名 */
  className?: string
  /** 是否显示标题 */
  showTitle?: boolean
}

const DEFAULT_STATUSES: Status[] = ['normal', 'warning', 'abnormal']

/**
 * 状态图例组件
 * 显示不同状态对应的图标和文字说明
 */
const StatusLegend: React.FC<StatusLegendProps> = ({
  statuses = DEFAULT_STATUSES,
  vertical = false,
  className = '',
  showTitle = true,
}) => {
  const { t } = useTranslation('pro/agui')
  const getStatusLabel = useStatusLabel()
  return (
    <div className={`${className}`}>
      {showTitle && (
        <div className="text-xs font-medium text-[var(--ant-color-text-secondary)] mb-2">
          {t('statusLegend')}
        </div>
      )}
      <div className={`flex items-center gap-3 ${vertical ? 'flex-col' : 'flex-wrap'}`}>
        {statuses.map((status) => (
          <div key={status} className="flex items-center gap-1.5 justify-center">
            <StatusIcon status={status} withContainer />
            <span className="text-xs text-[var(--ant-color-text-secondary)]">
              {getStatusLabel(status)}
            </span>
          </div>
        ))}
      </div>
    </div>
  )
}

export default StatusLegend
