import React from 'react'
import { BaseEdge, getMarkerEnd } from 'reactflow'

/**
 * Manhattan 正交折线
 * - 顶层布局为自上而下：优先“竖-横-竖”走线，避免压节点
 * - 支持虚线等样式：通过 style 传入 strokeDasharray 等
 */
export default function ManhattanEdge(props) {
  const { id, sourceX, sourceY, targetX, targetY, markerEnd, style } = props

  // 中间水平线 Y（取中点更美观）
  const midY = sourceY + (targetY - sourceY) / 2

  // M sx sy -> V midY -> H tx -> V ty
  const d = `M ${sourceX},${sourceY} L ${sourceX},${midY} L ${targetX},${midY} L ${targetX},${targetY}`
  const markerEndId = getMarkerEnd(markerEnd, id)

  return <BaseEdge id={id} path={d} markerEnd={markerEndId} style={style} interactionWidth={16} />
}
