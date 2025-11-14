import React from 'react'
import { useTranslation } from 'react-i18next'

export type Status = 'normal' | 'info' | 'muted' | 'warning' | 'abnormal'

export interface StatusIconProps {
  status: Status
  className?: string
  size?: number | string
  /** 是否显示带样式的容器 */
  withContainer?: boolean
  /** 自定义容器类名 */
  containerClassName?: string
}

/** 状态对应的颜色样式 */
const STATUS_COLORS: Record<
  Status,
  {
    iconBg: string
    iconFg: string
  }
> = {
  normal: {
    iconBg: 'bg-emerald-500',
    iconFg: 'text-white',
  },
  info: {
    iconBg: 'bg-teal-500',
    iconFg: 'text-white',
  },
  muted: {
    iconBg: 'bg-gray-300',
    iconFg: 'text-white',
  },
  warning: {
    iconBg: 'bg-amber-500',
    iconFg: 'text-white',
  },
  abnormal: {
    iconBg: 'bg-rose-500',
    iconFg: 'text-white',
  },
}

/**
 * 纯图标渲染（内部使用）
 */
const IconSvg: React.FC<{ status: Status; className?: string; style?: React.CSSProperties }> = ({
  status,
  className,
  style,
}) => {
  const baseClass = className || 'w-5 h-5'

  switch (status) {
    case 'normal':
      return (
        <svg viewBox="0 0 20 20" className={baseClass} style={style}>
          <path
            fill="currentColor"
            d="M8.5 13.3 4.8 9.6l1.4-1.4 2.3 2.3 5.3-5.3 1.4 1.4-6.7 6.7z"
          />
        </svg>
      )
    case 'warning':
      return (
        <svg viewBox="0 0 20 20" className={baseClass} style={style}>
          <path
            fill="currentColor"
            d="M10 1L2 17h16L10 1zm0 3l5.5 11H4.5L10 4zm-1 6h2v3H9v-3zm0 4h2v2H9v-2z"
          />
        </svg>
      )
    case 'abnormal':
      return (
        <svg viewBox="0 0 20 20" className={baseClass} style={style}>
          <path d="M9 6h2v6H9V6zm0 8h2v2H9v-2z" fill="currentColor" />
        </svg>
      )
    case 'info':
      return (
        <svg viewBox="0 0 20 20" className={baseClass} style={style}>
          <path
            fill="currentColor"
            d="M10 0a10 10 0 1 0 0 20A10 10 0 0 0 10 0zM9 4h2v2H9V4zm0 4h2v8H9V8z"
          />
        </svg>
      )
    case 'muted':
    default:
      return (
        <svg viewBox="0 0 20 20" className={baseClass} style={style}>
          <circle cx="10" cy="10" r="5" fill="currentColor" />
        </svg>
      )
  }
}

/**
 * 状态图标组件
 * 根据不同的状态显示对应的图标
 * 支持两种模式：纯图标模式和带容器模式
 */
const StatusIcon: React.FC<StatusIconProps> = ({
  status,
  className,
  size,
  withContainer = false,
  containerClassName,
}) => {
  const style = size ? { width: size, height: size } : undefined
  const colors = STATUS_COLORS[status]

  // 纯图标模式
  if (!withContainer) {
    return <IconSvg status={status} className={className} style={style} />
  }

  // 带容器模式
  const defaultContainerClasses = [
    'inline-flex items-center justify-center shrink-0',
    'w-7 h-7 rounded-xl',
    colors.iconBg,
    colors.iconFg,
    'shadow-sm ring-1 ring-white/60',
    'transition-transform duration-200 group-hover:scale-[1.04]',
  ].join(' ')

  return (
    <div className={containerClassName || defaultContainerClasses}>
      <IconSvg status={status} className={className} style={style} />
    </div>
  )
}

/**
 * 获取状态对应的文本标签
 */
export const useStatusLabel = () => {
  const { t } = useTranslation('pro/agui')

  return (status: Status): string => {
    return t(`status.${status}`)
  }
}

export default StatusIcon
