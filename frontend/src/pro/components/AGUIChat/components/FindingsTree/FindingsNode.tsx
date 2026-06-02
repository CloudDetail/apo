import React from 'react'
import { Handle, Position } from 'reactflow'
import { useTranslation } from 'react-i18next'
import StatusIcon, { type Status } from '../StatusIcon'
import type { FindingsNodeData } from './findingsContract'
import {
  getFindingsNodePresentation,
  type FindingsNodeThemeKey,
} from './findingsNodePresentation'

/** 颜色与样式主题 */
const COLORS: Record<
  FindingsNodeThemeKey,
  {
    card: string
    ring: string
    title: string
    sub: string
    iconBg: string
    iconFg: string
    handle: string
    hoverRing: string
    shadow: string
    accent: string
    badgeBg: string
    badgeFg: string
  }
> = {
  normal: {
    card: 'bg-emerald-50',
    ring: 'ring-emerald-200',
    title: 'text-emerald-900',
    sub: 'text-emerald-700/90',
    iconBg: 'bg-emerald-500',
    iconFg: 'text-white',
    handle: 'bg-emerald-400',
    hoverRing: 'ring-emerald-300',
    shadow: 'shadow-[0_10px_30px_-12px_rgba(16,185,129,0.35)]',
    accent: 'from-emerald-400 via-emerald-500 to-emerald-400',
    badgeBg: 'bg-emerald-600/90',
    badgeFg: 'text-white',
  },
  warning: {
    card: 'bg-amber-50',
    ring: 'ring-amber-200',
    title: 'text-amber-900',
    sub: 'text-amber-700/90',
    iconBg: 'bg-amber-500',
    iconFg: 'text-white',
    handle: 'bg-amber-400',
    hoverRing: 'ring-amber-300',
    shadow: 'shadow-[0_10px_30px_-12px_rgba(245,158,11,0.35)]',
    accent: 'from-amber-400 via-amber-500 to-amber-400',
    badgeBg: 'bg-amber-700/95',
    badgeFg: 'text-white',
  },
  abnormal: {
    card: 'bg-rose-50',
    ring: 'ring-rose-200',
    title: 'text-rose-900',
    sub: 'text-rose-700/90',
    iconBg: 'bg-rose-500',
    iconFg: 'text-white',
    handle: 'bg-rose-400',
    hoverRing: 'ring-rose-300',
    shadow: 'shadow-[0_10px_30px_-12px_rgba(244,63,94,0.35)]',
    accent: 'from-rose-400 via-rose-500 to-rose-400',
    badgeBg: 'bg-rose-700/95',
    badgeFg: 'text-white',
  },
  loading: {
    card: 'bg-slate-100',
    ring: 'ring-slate-300',
    title: 'text-slate-700',
    sub: 'text-slate-500',
    iconBg: 'bg-slate-400',
    iconFg: 'text-white',
    handle: 'bg-slate-300',
    hoverRing: 'ring-slate-400',
    shadow: 'shadow-[0_10px_26px_-14px_rgba(100,116,139,0.35)]',
    accent: 'from-slate-300 via-slate-400 to-slate-300',
    badgeBg: 'bg-slate-500/95',
    badgeFg: 'text-white',
  },
}

const FindingsNode = ({ data }: { data: FindingsNodeData }) => {
  const { t } = useTranslation('pro/agui')
  const {
    title,
    subtitle,
    status = 'warning',
    visualStatus,
    badge,
    active = false,
    disabled = false,
    onClick,
    content,
    nodeKind,
    hydrationStatus,
    errorType,
    errorMessage,
  } = data || {}
  const {
    resolvedStatus,
    themeKey,
    isRunning,
    isPending,
    isFailed,
    isLoading,
    badgeLabel,
    badgeLabelKey,
  } = getFindingsNodePresentation({
    status,
    visualStatus,
    nodeKind,
    hydrationStatus,
    badge,
  })
  const c = COLORS[themeKey]
  const displayBadgeLabel = badgeLabelKey ? t(badgeLabelKey) : badgeLabel

  const handleClick = () => {
    if (onClick) {
      onClick(buildFindingsNodeClickPayload(data, resolvedStatus))
    }
  }

  const handleKeyDown = (e: React.KeyboardEvent<HTMLDivElement>) => {
    if (e.key === 'Enter' || e.key === ' ') {
      e.preventDefault()
      handleClick()
    }
  }

  return (
    <div
      className={[
        'group relative select-none overflow-hidden',
        'w-[220px]  h-[80px]',
        'rounded-2xl px-3.5 py-3 pt-4', // 留出顶部细条空间
        'ring-1 ring-inset transition-all duration-200',
        c?.card,
        c?.ring,
        active ? `ring-2 ${c.hoverRing}` : 'hover:ring-2',
        disabled || isPending ? 'opacity-85 grayscale-[8%]' : c.shadow,
        'cursor-pointer',
      ].join(' ')}
      role={'button'}
      tabIndex={0}
      aria-disabled={disabled || !onClick}
      title={title || content || 'FindingsNode'}
      onClick={handleClick}
      onKeyDown={handleKeyDown}
      data-active={active ? 'true' : 'false'}
      data-disabled={disabled ? 'true' : 'false'}
    >
      {/* 顶部渐变细条（呼应“只要有顶部条就行”的简洁视觉） */}
      <div
        className={[
          'absolute left-0 right-0 top-0 h-2 rounded-t-2xl',
          'bg-gradient-to-r',
          c.accent,
          active ? 'opacity-100' : 'opacity-90 group-hover:opacity-100',
        ].join(' ')}
      />

      {/* 右上角徽标 */}
      {displayBadgeLabel && (
        <div
          className={[
            'absolute top-2 right-2 rounded-full px-2 py-1 text-[10px] leading-none',
            'backdrop-blur-sm shadow ring-1 ring-white/25',
            isFailed ? 'bg-rose-700/95' : c.badgeBg,
            c.badgeFg,
          ].join(' ')}
        >
          {displayBadgeLabel}
        </div>
      )}

      <div className="flex items-center gap-3">
        {/* 状态图标块 */}
        {isLoading ? (
          <div
            className={[
              'inline-flex items-center justify-center shrink-0',
              'w-7 h-7 rounded-xl',
              c.iconBg,
              c.iconFg,
              'shadow-sm ring-1 ring-white/60',
            ].join(' ')}
          >
            <span className="inline-block h-3.5 w-3.5 rounded-full border-2 border-current border-t-transparent animate-spin" />
          </div>
        ) : (
          <StatusIcon status={resolvedStatus as Status} withContainer />
        )}

        {/* 文本区 */}
        <div className="flex-1 min-w-0">
          <div
            className={[
              'text-[15px] font-semibold leading-snug truncate',
              displayBadgeLabel ? 'pr-16' : '',
              c.title,
            ].join(' ')}
          >
            {title || content || 'FindingsNode'}
          </div>

          {/* 副标题 */}
          {subtitle && (
            <div className={['mt-0.5 text-[12px] leading-snug truncate', c.sub].join(' ')}>
              {subtitle}
            </div>
          )}
          {isRunning && (
            <div className={['mt-1 flex items-center gap-1 text-[11px]', c.sub].join(' ')}>
              <span className="inline-block h-2 w-2 rounded-full border border-current border-t-transparent animate-spin" />
              <span>{t('findingsTree.hydrating')}</span>
            </div>
          )}
          {isFailed && (
            <div className="mt-1 text-[11px] leading-snug truncate text-rose-700">
              {errorMessage || errorType || t('findingsTree.hydrationFailedDetail')}
            </div>
          )}

          {/* 结果/结论小条（可选） */}
        </div>
      </div>

      {/* hover 高亮柔光边 */}
      <div
        className={[
          'pointer-events-none absolute inset-0 rounded-2xl',
          'ring-0 group-hover:ring-2 group-hover:ring-white/40',
          'transition-all',
        ].join(' ')}
      />

      {/* 顶/底连接点（ReactFlow Handle）—默认透明，避免喧宾夺主 */}
      <Handle
        type="target"
        position={Position.Left}
        className="!w-2.5 !h-2.5"
        style={{ left: 0, background: 'transparent', border: 'none', opacity: 0 }}
      />
      <Handle
        type="source"
        position={Position.Right}
        className="!w-2.5 !h-2.5"
        style={{ background: 'transparent', border: 'none', opacity: 0 }}
      />
    </div>
  )
}

export const buildFindingsNodeClickPayload = (
  data: FindingsNodeData,
  resolvedStatus: FindingsNodeData['visualStatus'],
): FindingsNodeData => {
  const { active: _active, disabled: _disabled, onClick: _onClick, ...payload } = data

  return {
    ...payload,
    visualStatus: resolvedStatus,
  }
}

export default React.memo(FindingsNode)
