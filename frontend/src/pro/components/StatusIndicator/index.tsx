import React from 'react'
import { IoWarning } from 'react-icons/io5'
import styles from './index.module.scss'

export type StatusType = 'loading' | 'error' | 'warning' | 'success'

export interface StatusItem {
  /** 状态类型 */
  status: StatusType
  /** 显示的文本标签 */
  label: string
  /** 唯一标识 */
  key?: string | number
}

interface StatusIndicatorProps {
  /** 状态项列表 */
  items: StatusItem[]
  /** 容器宽度 */
  width?: string | number
  /** 容器类名 */
  className?: string
  /** 容器样式 */
  style?: React.CSSProperties
}

/**
 * 状态指示器组件
 * 用于显示多个带有加载动画或警告图标的状态项
 */
const StatusIndicator: React.FC<StatusIndicatorProps> = ({
  items,
  width = '300px',
  className = '',
  style = {},
}) => {
  const renderStatusIcon = (status: StatusType) => {
    switch (status) {
      case 'loading':
      case 'success':
        return <div className={styles.loader}></div>
      case 'error':
        return <div className={styles.errorLoader}></div>
      case 'warning':
        return <IoWarning className="text-yellow-500 mr-1" />
      default:
        return null
    }
  }

  return (
    <div className={`${styles.loaderContainer} ${className}`} style={{ width, ...style }}>
      {items?.map((item, index) => (
        <div key={item.key ?? index} className="flex items-center">
          {renderStatusIcon(item.status)}
          <span className="text-xs text-[var(--ant-color-text-tertiary)]">{item.label}</span>
        </div>
      ))}
    </div>
  )
}

export default StatusIndicator
