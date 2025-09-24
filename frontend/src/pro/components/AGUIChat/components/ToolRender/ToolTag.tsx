import React, { memo, useMemo, useState, useCallback } from 'react'
import { useSelector } from 'react-redux'
import { FaDatabase, FaTools } from 'react-icons/fa'
export type ToolKind = 'toolCall' | 'agentTool'
type ToolPointer = { id: string; type: ToolKind; index: number }

interface ToolTagProps {
  /** 父组件传入：已含稳定 id 的“指针描述” */
  items: ToolPointer[]
  /** 父组件传入：仅用于显示名称，不渲染详情 */
  toolMap: Map<string, any>
  /** 父组件状态：哪些 id 已展开 */
  expandedIds: Set<string>
  /** 父组件回调：切换展开 */
  onToggle: (id: string) => void
  /** 最大显示数量，超过则省略为 +N */
  maxVisible?: number
}

/** 仅负责“标签行 + 交互回调”，不渲染任何工具详情 */
const ToolTag: React.FC<ToolTagProps> = memo(
  ({ items, toolMap, expandedIds, onToggle, maxVisible = 3 }) => {
    const { theme } = useSelector((state: any) => state.settingReducer)
    const [showAll, setShowAll] = useState(false)
    const visible = useMemo(
      () => (showAll ? items : items.slice(0, maxVisible)),
      [items, showAll, maxVisible],
    )
    const hiddenCount = Math.max(0, items.length - maxVisible)
    const showMore = useCallback(() => setShowAll(true), [])

    if (items.length === 0) return null

    return (
      <div className="inline-flex flex-wrap items-center gap-1">
        {visible.map((desc) => {
          const obj = toolMap.get(desc.id)
          const name = obj?.name || obj?.toolName || 'Unknown'
          const isOpen = expandedIds.has(desc.id)
          const icon = desc.type === 'toolCall' ? <FaTools /> : <FaDatabase />

          return (
            <button
              key={desc.id}
              type="button"
              onClick={() => onToggle(desc.id)}
              title={name}
              className={`bg-[var(--ant-color-bg-container-disabled)] text-[var(--ant-color-text-secondary)] px-2 py-[2px] rounded-lg text-[10px] 
                hover:bg-[var(--ant-control-item-bg-active-hover)] transition-colors inline-flex items-center ${
                  isOpen ? 'ring-1 ring-gray-500' : ''
                }`}
              aria-pressed={isOpen}
            >
              <span className="mr-1 inline-flex">{icon}</span>
              {name}
            </button>
          )
        })}

        {!showAll && hiddenCount > 0 && (
          <button
            type="button"
            onClick={showMore}
            title={`open ${hiddenCount} more tools`}
            className="bg-[var(--ant-color-bg-container)] text-[var(--ant-color-text-secondary)] px-2 py-[2px] rounded-lg text-[10px] 
            hover:bg-[var(--ant-control-item-bg-active-hover)] transition-colors inline-flex items-center"
          >
            +{hiddenCount}
          </button>
        )}
      </div>
    )
  },
)

export default ToolTag
