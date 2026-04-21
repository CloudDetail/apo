import React, { useCallback, useEffect, useRef, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { useSelector } from 'react-redux'
import ReactFlow, {
  Background,
  BackgroundVariant,
  MiniMap,
  ReactFlowProvider,
  MarkerType,
  useEdgesState,
  useNodesState,
  useReactFlow,
} from 'reactflow'
import 'reactflow/dist/style.css'
import FindingsNode from './FindingsNode'
import { buildFindingsLayout } from './findingsLayout'

const NODE_WIDTH = 220
const NODE_HEIGHT = 72
const EDGE_BADGE_PADDING = 12
// const _INITIAL_ZOOM = 1.2
const BADGE_Z = 1000

type VisualStatus = 'normal' | 'warning' | 'abnormal'

const getNodeVisualStatus = (data?: Record<string, any>): VisualStatus => {
  if (data?.visualStatus === 'abnormal' || data?.visualStatus === 'warning' || data?.visualStatus === 'normal') {
    return data.visualStatus
  }
  if (data?.nodeKind === 'unsupported') {
    return 'normal'
  }
  const status = data?.status
  if (status === 'abnormal' || status === 'warning' || status === 'normal') {
    return status
  }
  return 'normal'
}

const getStatusColor = (status: VisualStatus) =>
  status === 'abnormal' ? '#ff4d6d' : status === 'warning' ? '#f59e0b' : '#22c55e'

const getStatusStrokeColor = (status: VisualStatus) =>
  status === 'abnormal' ? '#dc2626' : status === 'warning' ? '#d97706' : '#059669'

// eslint-disable-next-line @typescript-eslint/no-unused-vars
function _BasicNode({ data }: any) {
  const color = getStatusColor(getNodeVisualStatus(data))
  return (
    <div
      style={{
        width: NODE_WIDTH,
        height: NODE_HEIGHT,
        borderRadius: 12,
        border: `1px solid ${color}55`,
        background: 'rgba(255,255,255,0.04)',
        color: '#e5e7eb',
        padding: 12,
        boxShadow: `0 4px 16px ${color}1f`,
      }}
    >
      <div style={{ fontSize: 14, marginBottom: 6 }}>
        <b>{data?.label || data?.id}</b>
      </div>
      <div style={{ fontSize: 12, opacity: 0.9 }}>
        status: <span style={{ color }}>{getNodeVisualStatus(data)}</span>
      </div>
    </div>
  )
}

const nodeTypes = { findingsNode: FindingsNode }

/** 反推当前 zoom & 世界坐标原点 */
function getZoomAndWorldOrigin(rf: any) {
  const w0 = rf.project({ x: 0, y: 0 })
  const wX = rf.project({ x: 100, y: 0 })
  const wY = rf.project({ x: 0, y: 100 })
  const zx = 100 / (wX.x - w0.x)
  const zy = 100 / (wY.y - w0.y)
  const zoom = (zx + zy) / 2
  return { zoom, worldAtScreen0: w0 }
}

function worldToScreenRel(wx: number, wy: number, zoom: number, w0: { x: number; y: number }) {
  return { sx: (wx - w0.x) * zoom, sy: (wy - w0.y) * zoom }
}

/** 工具：拿根节点(入度0) */
function getRootId(ns: any[], es: any[]) {
  const targets = new Set(es.map((e: any) => e.target))
  return ns.find((n: any) => !targets.has(n.id))?.id || ns[0]?.id
}

function buildChildrenMap(es: any[]) {
  const id2children = new Map<string, string[]>()
  es.forEach((e: any) => {
    if (!id2children.has(e.source)) id2children.set(e.source, [])
    id2children.get(e.source)!.push(e.target)
  })
  return id2children
}

export function FlowCore({
  data,
  findingsId,
  onNodeClick,
}: {
  data: any
  findingsId?: string
  onNodeClick: (node: any) => void
}) {
  const { t } = useTranslation('pro/agui')
  const { theme } = useSelector((state: any) => state.settingReducer)
  const isDark = theme === 'dark'
  const [nodes, setNodes, onNodesChange] = useNodesState([])
  const [edges, setEdges, onEdgesChange] = useEdgesState([])
  const rf = useReactFlow()
  const containerRef = useRef<HTMLDivElement | null>(null)
  const rafRef = useRef<number | null>(null)
  const [isReactFlowReady, setIsReactFlowReady] = useState(false)
  const hasInitializedView = useRef(false) // 标记是否已完成初始视图定位
  const [miniMapSize, setMiniMapSize] = useState({ width: 200, height: 150 })

  const [badges, setBadges] = useState<
    {
      id: string
      status: 'normal' | 'warning' | 'abnormal'
      left: number
      top: number
      dir: string
      count: number
    }[]
  >([])

  /** 初始化布局数据 */
  useEffect(() => {
    if (!data || !data.nodes || data.nodes.length === 0) {
      setNodes([])
      setEdges([])
      hasInitializedView.current = false // 重置视图初始化标记
      return
    }

    const { nodes: layoutNodes, edges: layoutEdges } = buildFindingsLayout(data, findingsId, onNodeClick)
    const styledEdges = layoutEdges.map((edge) => {
      const targetNode = layoutNodes.find((node) => node.id === edge.target)
      const stroke = getStatusColor(getNodeVisualStatus(targetNode?.data))
      return {
        ...edge,
        style: { stroke, strokeWidth: 2.5 },
        markerEnd: { type: MarkerType.ArrowClosed, width: 16, height: 16, color: stroke },
      }
    })

    setNodes(layoutNodes)
    setEdges(styledEdges)
    hasInitializedView.current = false // 新数据时重置视图初始化标记
  }, [data, findingsId, onNodeClick])

  /** 视图定位 - 只在初次加载时执行一次 */
  useEffect(() => {
    // 如果已经初始化过视图，则不再执行
    if (hasInitializedView.current) {
      return
    }

    if (!isReactFlowReady || nodes.length === 0 || !rf) {
      return
    }

    const initializeView = () => {
      if (!rf.getViewport || !rf.fitView || !rf.setCenter) {
        setTimeout(initializeView, 100)
        return
      }

      try {
        const sortByLayerAndOrder = (a: any, b: any) => {
          if (a.position.x !== b.position.x) return a.position.x - b.position.x
          return a.position.y - b.position.y
        }

        const abnormalNodes = nodes
          .filter((n: any) => getNodeVisualStatus(n.data) === 'abnormal')
          .sort(sortByLayerAndOrder)
        const warningNodes = nodes
          .filter((n: any) => getNodeVisualStatus(n.data) === 'warning')
          .sort(sortByLayerAndOrder)

        let targetNode: any = null
        let fallbackFitAll = false

        if (abnormalNodes.length > 0) targetNode = abnormalNodes[0]
        else if (warningNodes.length > 0) targetNode = warningNodes[0]
        else fallbackFitAll = true

        if (fallbackFitAll || !targetNode) {
          rf.fitView({ duration: 800, padding: 0.15 })
          hasInitializedView.current = true // 标记已完成初始化
          return
        }

        // 聚焦到第一个异常/警告节点
        const cx = targetNode.position.x + (targetNode.width || NODE_WIDTH) / 3
        const cy = targetNode.position.y + (targetNode.height || NODE_HEIGHT) / 3
        rf.setCenter(cx, cy, { zoom: 1, duration: 800 })
        hasInitializedView.current = true // 标记已完成初始化
      } catch {
        setTimeout(() => {
          try {
            rf.fitView({ duration: 500, padding: 0.1 })
            hasInitializedView.current = true // 标记已完成初始化
          } catch {
            // 忽略错误
          }
        }, 200)
      }
    }

    const timer = setTimeout(initializeView, 300)
    return () => clearTimeout(timer)
  }, [isReactFlowReady, nodes.length, rf])

  /** 计算边缘 Badge（离屏告警提示） */
  const computeBadges = useCallback(() => {
    if (!containerRef.current || nodes.length === 0) return
    const rect = containerRef.current.getBoundingClientRect()
    const { zoom, worldAtScreen0 } = getZoomAndWorldOrigin(rf)

    const rootId = getRootId(nodes, edges)
    const id2children = buildChildrenMap(edges)

    const nextBadges: any[] = []
    const offScreenNodes: { node: any; status: 'abnormal' | 'warning' }[] = []

    nodes.forEach((node) => {
      const status = getNodeVisualStatus(node.data)
      if (status !== 'abnormal' && status !== 'warning') return

      const worldCx = node.position.x + (node.width || NODE_WIDTH) / 2
      const worldCy = node.position.y + (node.height || NODE_HEIGHT) / 2
      const { sx, sy } = worldToScreenRel(worldCx, worldCy, zoom, worldAtScreen0)

      const pad = EDGE_BADGE_PADDING
      const insideX = sx >= pad && sx <= rect.width - pad
      const insideY = sy >= pad && sy <= rect.height - pad

      if (!insideX || !insideY) {
        offScreenNodes.push({ node, status })
      }
    })

    const directionGroups = new Map<string, { abnormal: any[]; warning: any[] }>()
    offScreenNodes.forEach(({ node, status }) => {
      const worldCx = node.position.x + (node.width || NODE_WIDTH) / 2
      const worldCy = node.position.y + (node.height || NODE_HEIGHT) / 2
      const { sx, sy } = worldToScreenRel(worldCx, worldCy, zoom, worldAtScreen0)

      const pad = EDGE_BADGE_PADDING
      let dir = 'left'
      if (sx < pad) dir = 'left'
      else if (sx > rect.width - pad) dir = 'right'
      else if (sy < pad) dir = 'top'
      else dir = 'bottom'

      if (!directionGroups.has(dir)) {
        directionGroups.set(dir, { abnormal: [], warning: [] })
      }
      directionGroups.get(dir)![status].push(node)
    })

    directionGroups.forEach((groups, dir) => {
      ;(['abnormal', 'warning'] as const)
        .filter((status) => groups[status].length > 0)
        .forEach((status, index) => {
          const statusNodes = groups[status]
          const representativeNode = statusNodes[0]
          const worldCx =
            representativeNode.position.x + (representativeNode.width || NODE_WIDTH) / 2
          const worldCy =
            representativeNode.position.y + (representativeNode.height || NODE_HEIGHT) / 2
          const { sx, sy } = worldToScreenRel(worldCx, worldCy, zoom, worldAtScreen0)

          const pad = EDGE_BADGE_PADDING
          const badgeSpacing = 50
          let left = sx
          let top = sy

          if (dir === 'left') {
            left = pad
            top = Math.min(Math.max(sy, pad), rect.height - pad) + index * badgeSpacing
            top = Math.min(top, rect.height - pad - 30)
          } else if (dir === 'right') {
            left = rect.width - pad
            top = Math.min(Math.max(sy, pad), rect.height - pad) + index * badgeSpacing
            top = Math.min(top, rect.height - pad - 30)
          } else if (dir === 'top') {
            top = pad
            left = Math.min(Math.max(sx, pad), rect.width - pad) + index * badgeSpacing
            left = Math.min(left, rect.width - pad - 80)
          } else {
            top = rect.height - pad
            left = Math.min(Math.max(sx, pad), rect.width - pad) + index * badgeSpacing
            left = Math.min(left, rect.width - pad - 80)
          }

          nextBadges.push({
            id: `${dir}-${status}`,
            status,
            left,
            top,
            dir,
            count: statusNodes.length,
            nodes: statusNodes,
            rootId,
            id2children,
          })
        })
    })

    setBadges(nextBadges)
  }, [nodes, edges, rf])

  const scheduleBadges = useCallback(() => {
    if (rafRef.current) return
    rafRef.current = requestAnimationFrame(() => {
      rafRef.current = null
      computeBadges()
    })
  }, [computeBadges])

  useEffect(() => {
    scheduleBadges()
  }, [nodes, edges, scheduleBadges])

  const onMove = () => scheduleBadges()
  const onMoveEnd = () => scheduleBadges()

  const onInit = useCallback(() => {
    console.log('ReactFlow 初始化完成')
    setIsReactFlowReady(true)
  }, [])

  // 根据屏幕大小调整 MiniMap 尺寸
  useEffect(() => {
    const updateMiniMapSize = () => {
      const width = window.innerWidth
      console.log(width)
      if (width < 640) {
        // 小屏幕 (移动设备) - 使用非常小的尺寸
        setMiniMapSize({ width: 80, height: 60 })
      } else if (width < 1024) {
        // 中等屏幕 (平板)
        setMiniMapSize({ width: 120, height: 90 })
      } else if (width < 1600) {
        // 中等屏幕 (平板)
        setMiniMapSize({ width: 120, height: 90 })
      } else {
        // 大屏幕 (桌面)
        setMiniMapSize({ width: 200, height: 150 })
      }
    }

    updateMiniMapSize()
    window.addEventListener('resize', updateMiniMapSize)
    return () => window.removeEventListener('resize', updateMiniMapSize)
  }, [])

  const focusBadge = (badge: any) => {
    const targetNodes = badge.nodes || []
    if (targetNodes.length === 0) return

    const abnormalNodes = targetNodes.filter((n: any) => getNodeVisualStatus(n.data) === 'abnormal')
    const warningNodes = targetNodes.filter((n: any) => getNodeVisualStatus(n.data) === 'warning')
    const target = abnormalNodes.length > 0 ? abnormalNodes[0] : warningNodes[0]
    if (!target) return

    const cx = target.position.x + (target.width || NODE_WIDTH) / 2
    const cy = target.position.y + (target.height || NODE_HEIGHT) / 2
    try {
      rf.setCenter(cx, cy, { duration: 400 })
    } catch {
      /* ignore */
    }
  }

  return (
    <div ref={containerRef} style={{ width: '100%', height: '100%', position: 'relative' }}>
      {/* MiniMap 自定义样式 */}
      <style>{`
        .react-flow__minimap {
          background: ${isDark ? 'rgba(17, 24, 39, 0.95)' : 'rgba(248, 250, 252, 0.95)'} !important;
          border: 1px solid ${isDark ? 'rgba(255, 255, 255, 0.1)' : 'rgba(0, 0, 0, 0.1)'} !important;
          border-radius: 12px !important;
          box-shadow: ${
            isDark
              ? '0 10px 30px -10px rgba(0, 0, 0, 0.5), 0 0 0 1px rgba(255, 255, 255, 0.05)'
              : '0 10px 30px -10px rgba(0, 0, 0, 0.1), 0 0 0 1px rgba(0, 0, 0, 0.05)'
          } !important;
          backdrop-filter: blur(12px);
        }
        .react-flow__minimap-mask {
          fill: ${isDark ? 'rgba(0, 0, 0, 0.6)' : 'rgba(255, 255, 255, 0.6)'} !important;
        }
        .react-flow__minimap-node {
          stroke-width: 2 !important;
        }
      `}</style>

      {/* 全览图控制按钮 */}
      <div
        style={{
          position: 'absolute',
          top: 16,
          right: 16,
          zIndex: BADGE_Z + 2, // 提高在 MiniMap 之上
          display: 'flex',
          gap: 8,
        }}
      ></div>

      <ReactFlow
        nodes={nodes}
        edges={edges}
        nodeTypes={nodeTypes}
        onNodesChange={onNodesChange}
        onEdgesChange={onEdgesChange}
        onInit={onInit}
        minZoom={0.1}
        maxZoom={2}
        zoomOnScroll
        zoomOnPinch
        panOnDrag
        fitView={false}
        proOptions={{ hideAttribution: true }}
        onMove={onMove}
        onMoveEnd={onMoveEnd}
        style={{
          background: isDark
            ? 'linear-gradient(180deg,#0b0b0d,#0f1724)'
            : 'linear-gradient(180deg,#f8fafc,#e2e8f0)',
        }}
      >
        <Background
          variant={BackgroundVariant.Dots}
          gap={14}
          size={1}
          color={isDark ? '#374151' : '#cbd5e1'}
        />

        <MiniMap
          nodeColor={(node) => {
            return getStatusColor(getNodeVisualStatus(node.data as any))
          }}
          nodeStrokeWidth={1.5}
          nodeStrokeColor={(node) => {
            return getStatusStrokeColor(getNodeVisualStatus(node.data as any))
          }}
          nodeBorderRadius={6}
          maskColor={isDark ? 'rgba(0, 0, 0, 0.7)' : 'rgba(255, 255, 255, 0.7)'}
          style={{
            height: miniMapSize.height,
            width: miniMapSize.width,
          }}
          position="bottom-right"
          pannable
          zoomable
          inversePan={false}
          zoomStep={0.5}
        />
      </ReactFlow>

      {badges.map((b) => {
        const isAbnormal = b.status === 'abnormal'
        const color = isAbnormal ? '#ff4d6d' : '#f59e0b'
        const bgColor = isDark
          ? isAbnormal
            ? 'rgba(255, 77, 109, 0.15)'
            : 'rgba(245, 158, 11, 0.15)'
          : isAbnormal
            ? 'rgba(255, 77, 109, 0.08)'
            : 'rgba(245, 158, 11, 0.08)'
        const borderColor = isDark
          ? isAbnormal
            ? 'rgba(255, 77, 109, 0.4)'
            : 'rgba(245, 158, 11, 0.4)'
          : isAbnormal
            ? 'rgba(255, 77, 109, 0.2)'
            : 'rgba(245, 158, 11, 0.2)'
        const shadowColor = isDark
          ? isAbnormal
            ? 'rgba(255, 77, 109, 0.3)'
            : 'rgba(245, 158, 11, 0.3)'
          : isAbnormal
            ? 'rgba(255, 77, 109, 0.15)'
            : 'rgba(245, 158, 11, 0.15)'
        const arrow = b.dir === 'left' ? '←' : b.dir === 'right' ? '→' : b.dir === 'top' ? '↑' : '↓'
        const statusText = isAbnormal ? t('findingsTree.abnormal') : t('findingsTree.warning')

        return (
          <div
            key={b.id}
            onClick={() => focusBadge(b)}
            title={t('findingsTree.directionContains', { count: b.count, statusText })}
            style={{
              position: 'absolute',
              left: b.left,
              top: b.top,
              transform: 'translate(-50%,-50%)',
              zIndex: BADGE_Z,
              display: 'flex',
              alignItems: 'center',
              gap: 6,
              cursor: 'pointer',
              userSelect: 'none',
              padding: '6px 10px',
              background: bgColor,
              backdropFilter: 'blur(8px)',
              borderRadius: 12,
              border: `1px solid ${borderColor}`,
              boxShadow: `0 4px 12px ${shadowColor}, 0 0 0 1px ${isDark ? 'rgba(255,255,255,0.05)' : 'rgba(0,0,0,0.05)'}`,
              transition: 'all 0.2s ease',
            }}
            onMouseEnter={(e) => {
              e.currentTarget.style.transform = 'translate(-50%,-50%) scale(1.05)'
              e.currentTarget.style.boxShadow = `0 6px 20px ${shadowColor}, 0 0 0 1px ${isDark ? 'rgba(255,255,255,0.1)' : 'rgba(0,0,0,0.1)'}`
            }}
            onMouseLeave={(e) => {
              e.currentTarget.style.transform = 'translate(-50%,-50%) scale(1)'
              e.currentTarget.style.boxShadow = `0 4px 12px ${shadowColor}, 0 0 0 1px ${isDark ? 'rgba(255,255,255,0.05)' : 'rgba(0,0,0,0.05)'}`
            }}
          >
            <div
              style={{
                width: 20,
                height: 20,
                borderRadius: '50%',
                background: `linear-gradient(135deg, ${color}, ${color}dd)`,
                color: '#fff',
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
                fontSize: 10,
                fontWeight: 'bold',
                boxShadow: `0 2px 4px ${color}40`,
              }}
            >
              {isAbnormal ? '!' : '⚠'}
            </div>

            <span
              style={{
                fontSize: 12,
                fontWeight: '600',
                color: color,
                minWidth: '16px',
                textAlign: 'center',
              }}
            >
              {b.count}
            </span>

            <div
              style={{
                fontSize: 14,
                color: color,
                fontWeight: 'bold',
                opacity: 0.8,
              }}
            >
              {arrow}
            </div>
          </div>
        )
      })}
    </div>
  )
}

export default function FindingsTopology({
  data,
  findingsId,
  onNodeClick,
}: {
  data: any
  findingsId?: string
  onNodeClick: (node: any) => void
}) {
  return (
    <div style={{ width: '100%', height: '100%' }}>
      <ReactFlowProvider>
        <FlowCore data={data} findingsId={findingsId} onNodeClick={onNodeClick} />
      </ReactFlowProvider>
    </div>
  )
}
