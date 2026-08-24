import React, { useEffect, useMemo, useRef, useState } from 'react'
import { Button, Card, Descriptions, Empty, Input, Space, Tag } from 'antd'
import { SearchOutlined } from '@ant-design/icons'
import { useTranslation } from 'react-i18next'
import './index.scss'

interface SpanDetailSummary {
  hasDetail?: boolean
  detailLevel?: string
  totalNodes?: number
  maxDepth?: number
  typeCounts?: Record<string, number>
}

interface SpanDetailItem {
  startTime?: number
  duration?: number
  exclusiveTime?: number
  offset?: number
  depth?: number
  name?: string
  type?: string
  attributes?: Record<string, string>
}

interface SpanDetail {
  traceId?: string
  spanId?: string
  details?: SpanDetailItem[]
  summary?: SpanDetailSummary
}

interface TreeNode extends SpanDetailItem {
  key: string
  index: number
  children: TreeNode[]
  parentKey?: string
  compact?: boolean
  compactCount?: number
  compactKind?: 'siblings' | 'parent'
  compactStack?: StackFrame[]
  compactSourceKeys?: string[]
  subtreeExclusiveTime?: number
  descendantCount?: number
}

interface VisibleNode extends TreeNode {
  visibleDepth: number
}

interface StackFrame {
  key: string
  text: string
  type?: string
  current?: boolean
  compacted?: boolean
}

interface Props {
  data: any
}

const DEFAULT_COMPACT_PERCENT = 0.5
const ROW_HEIGHT = 46
const DETAIL_ROW_HEIGHT = 262
const DETAIL_STACK_ROW_HEIGHT = 430
const KEY_TYPES = new Set(['EXTERNAL', 'DATABASE', 'NOSQL', 'MQ', 'POOL'])
const TYPE_COLORS: Record<string, string> = {
  CODE: '#3479e8',
  EXTERNAL: '#9b7e7d',
  POOL: '#6dc8ec',
  DATABASE: '#5ad8a6',
}
const SPAN_TYPE_LABELS: Record<string, string> = {
  CODE: '代码',
  DATABASE: '数据库',
  EXTERNAL: '外部服务',
  POOL: '连接池',
  NOSQL: 'NoSQL 数据库',
  MQ: '消息队列',
}
const DETAIL_LEVEL_LABELS: Record<string, string> = {
  full: '完整明细',
  partial: '部分明细',
  none: '无明细',
}

const formatDuration = (value?: number) => {
  if (!value || value <= 0) {
    return '0 ms'
  }
  if (value >= 1000) {
    return `${(value / 1000).toFixed(2)} ms`
  }
  return `${Math.round(value)} us`
}

const getSpanDetail = (data: any): SpanDetail | undefined => data?.object?.spanDetail

const getSpanTypeLabel = (type: string) => SPAN_TYPE_LABELS[type] || type

const getDetailLevelLabel = (level?: string) => {
  if (!level) return '-'
  return DETAIL_LEVEL_LABELS[level] || level
}

const getShortName = (name?: string) => {
  if (!name) return '-'
  if (name.startsWith('http')) return name
  const parts = name.split('.')
  return parts.slice(-2).join('.') || name
}

const formatStackFrame = (node: TreeNode) => {
  const name = node.name || '-'
  const className = node.attributes?.class || node.attributes?.Class || ''
  const methodName = node.attributes?.method || node.attributes?.Method || ''
  const frameType = node.type && node.type !== 'CODE' ? node.type : ''
  let text = name
  if (className && methodName) {
    text = `${className}.${methodName}`
  } else if (className) {
    text = `${className}.${getShortName(name)}`
  }
  return {
    key: node.key,
    text,
    type: frameType,
  }
}

const getAttribute = (node: TreeNode | undefined, ...keys: string[]) => {
  if (!node?.attributes) return ''
  for (const key of keys) {
    const value = node.attributes[key]
    if (value) return value
  }
  return ''
}

const buildSpanInfoItems = (node: TreeNode | undefined) => {
  if (!node) return []
  const attrs = node.attributes || {}
  return [
    { label: 'Name', value: node.name },
    { label: 'callerApplication', value: getAttribute(node, 'callerApplication') },
    { label: 'callerInstance', value: getAttribute(node, 'callerInstance') },
    { label: 'threadName', value: getAttribute(node, 'threadName') },
    { label: 'Type', value: node.compact ? '低耗时合并' : node.type },
    { label: '调用耗时', value: formatDuration(node.duration) },
    { label: '自身耗时', value: formatDuration(node.exclusiveTime) },
    { label: 'Offset', value: formatDuration(node.offset) },
    { label: 'Class', value: getAttribute(node, 'class', 'Class') },
    { label: 'Method', value: getAttribute(node, 'method', 'Method') },
    { label: 'URL', value: getAttribute(node, 'url') },
    { label: 'Operation', value: getAttribute(node, 'operation') },
    { label: 'System', value: getAttribute(node, 'system') },
    { label: 'Span ID', value: attrs.spanId },
  ].filter((item) => item.value !== undefined && item.value !== '')
}

const getCompactStack = (nodes: TreeNode[], limit = 80): StackFrame[] => {
  const stack: StackFrame[] = []
  const visitedKeys = new Set<string>()
  const visit = (node: TreeNode) => {
    if (stack.length >= limit || visitedKeys.has(node.key)) return
    visitedKeys.add(node.key)
    stack.push({ ...formatStackFrame(node), compacted: true })
    node.children.forEach((child) => visit(child))
  }
  nodes.forEach((node) => visit(node))
  return stack
}

const getSubtreeStats = (node: TreeNode) => {
  if (node.compact) {
    return {
      exclusiveTime: node.subtreeExclusiveTime || node.exclusiveTime || 0,
      count: node.descendantCount || node.compactSourceKeys?.length || 1,
      keys: node.compactSourceKeys || [node.key],
    }
  }
  let exclusiveTime = node.exclusiveTime || 0
  let count = 1
  const keys = [node.key]
  node.children.forEach((child) => {
    const stats = getSubtreeStats(child)
    exclusiveTime += stats.exclusiveTime
    count += stats.count
    keys.push(...stats.keys)
  })
  return { exclusiveTime, count, keys }
}

const isLowSelfCostNode = (node: TreeNode, threshold: number): boolean => {
  if ((node.exclusiveTime || 0) >= threshold) return false
  return !KEY_TYPES.has(node.type || '')
}

const buildTree = (details: SpanDetailItem[]): TreeNode[] => {
  const roots: TreeNode[] = []
  const stack: TreeNode[] = []

  details.forEach((item, index) => {
    const node: TreeNode = {
      ...item,
      key: `${item.name || 'span'}-${index}`,
      index: index + 1,
      children: [],
    }
    const depth = item.depth || 0
    stack.length = depth
    const parent = stack[depth - 1]
    if (parent) {
      node.parentKey = parent.key
      parent.children.push(node)
    } else {
      roots.push(node)
    }
    stack[depth] = node
  })

  return roots
}

const hasImportantSpan = (node: TreeNode, threshold: number): boolean => {
  if ((node.exclusiveTime || 0) >= threshold) return true
  if (KEY_TYPES.has(node.type || '') && (node.duration || 0) > 0) return true
  return node.children.some((child) => hasImportantSpan(child, threshold))
}

const hasHotSpan = (node: TreeNode, threshold: number): boolean => {
  if ((node.exclusiveTime || 0) >= threshold) return true
  return node.children.some((child) => hasHotSpan(child, threshold))
}

const isLowCostSiblingCandidate = (node: TreeNode, threshold: number): boolean => {
  if ((node.depth || 0) < 2) return false
  const type = node.type || ''
  const durationLow = (node.duration || 0) < threshold
  if (!KEY_TYPES.has(type)) {
    return (node.exclusiveTime || 0) < threshold && durationLow
  }
  const stats = getSubtreeStats(node)
  const totalExclusiveTime = node.subtreeExclusiveTime ?? stats.exclusiveTime
  return totalExclusiveTime < threshold && durationLow && !hasHotSpan(node, threshold)
}

const createCompactNode = (
  nodes: TreeNode[],
  threshold: number,
  kind: 'siblings' | 'parent',
  parentKey?: string,
): TreeNode => {
  const first = nodes[0]
  const offset = Math.min(...nodes.map((item) => item.offset || 0))
  const end = Math.max(...nodes.map((item) => (item.offset || 0) + (item.duration || 0)))
  const stats = nodes.reduce(
    (result, item) => {
      const nodeStats = getSubtreeStats(item)
      result.exclusiveTime += nodeStats.exclusiveTime
      result.count += nodeStats.count
      result.keys.push(...nodeStats.keys)
      return result
    },
    { exclusiveTime: 0, count: 0, keys: [] as string[] },
  )
  const compactCount = kind === 'parent' ? Math.max(stats.count - 1, 0) : nodes.length
  return {
    key: `compact-${kind}-${first.key}-${compactCount}-${Math.round(threshold)}`,
    index: first.index,
    depth: first.depth,
    parentKey,
    name: kind === 'parent' ? first.name : `${compactCount} low-cost spans compacted`,
    type: kind === 'parent' ? first.type || 'CODE' : 'CODE',
    offset,
    duration: end - offset,
    exclusiveTime: stats.exclusiveTime,
    subtreeExclusiveTime: stats.exclusiveTime,
    descendantCount: stats.count,
    compact: true,
    compactKind: kind,
    compactCount,
    compactStack: getCompactStack(nodes),
    compactSourceKeys: stats.keys,
    children: nodes,
  }
}

const createParentChainCompactNode = (
  chainNodes: TreeNode[],
  children: TreeNode[],
  threshold: number,
  parentKey?: string,
): TreeNode => {
  const first = chainNodes[0]
  const last = chainNodes[chainNodes.length - 1]
  const compactCount = chainNodes.length - 1
  const exclusiveTime = chainNodes.reduce((sum, item) => sum + (item.exclusiveTime || 0), 0)
  return {
    ...first,
    key: `compact-parent-chain-${first.key}-${last.key}-${Math.round(threshold)}`,
    parentKey,
    name: first.name,
    exclusiveTime,
    subtreeExclusiveTime: exclusiveTime,
    descendantCount: chainNodes.length,
    compact: true,
    compactKind: 'parent',
    compactCount,
    compactStack: getCompactStack(chainNodes),
    compactSourceKeys: chainNodes.map((node) => node.key),
    children,
  }
}

const compactLowCostParentChain = (
  node: TreeNode,
  threshold: number,
  restoredNodeKeys: Set<string>,
  parentKey?: string,
): TreeNode | undefined => {
  if (
    restoredNodeKeys.has(node.key) ||
    !isLowSelfCostNode(node, threshold) ||
    node.children.length !== 1
  ) {
    return undefined
  }

  const chainNodes: TreeNode[] = [node]
  let cursor = node.children[0]

  while (
    cursor &&
    !restoredNodeKeys.has(cursor.key) &&
    isLowSelfCostNode(cursor, threshold) &&
    cursor.children.length === 1
  ) {
    chainNodes.push(cursor)
    cursor = cursor.children[0]
  }

  if (chainNodes.length < 2) return undefined

  let children: TreeNode[] = []
  if (cursor) {
    children = compactChildren([cursor], threshold, restoredNodeKeys, chainNodes[chainNodes.length - 1].key)
  }

  return createParentChainCompactNode(chainNodes, children, threshold, parentKey)
}

const compactChildren = (
  nodes: TreeNode[],
  threshold: number,
  restoredNodeKeys: Set<string>,
  parentKey?: string,
): TreeNode[] => {
  const result: TreeNode[] = []
  let compactBuffer: TreeNode[] = []

  const flushCompactBuffer = () => {
    if (compactBuffer.length === 0) return
    if (compactBuffer.length > 1) {
      result.push(createCompactNode(compactBuffer, threshold, 'siblings', parentKey))
    } else {
      result.push(compactBuffer[0])
    }
    compactBuffer = []
  }

  nodes.forEach((node) => {
    const parentChainCompact = compactLowCostParentChain(node, threshold, restoredNodeKeys, parentKey)
    if (parentChainCompact) {
      if (isLowCostSiblingCandidate(parentChainCompact, threshold)) {
        compactBuffer.push(parentChainCompact)
      } else {
        flushCompactBuffer()
        result.push(parentChainCompact)
      }
      return
    }

    const children = compactChildren(node.children || [], threshold, restoredNodeKeys, node.key)
    const nextNode = { ...node, children }
    if ((!hasImportantSpan(nextNode, threshold) || isLowCostSiblingCandidate(nextNode, threshold)) && (node.depth || 0) >= 2) {
      compactBuffer.push(nextNode)
      return
    }
    flushCompactBuffer()
    result.push(nextNode)
  })

  flushCompactBuffer()
  return result
}

const findPathByKey = (nodes: TreeNode[], key: string, path: TreeNode[] = []): TreeNode[] => {
  for (const node of nodes) {
    const nextPath = [...path, node]
    if (node.key === key || node.compactSourceKeys?.includes(key)) {
      return nextPath
    }
    const childPath = findPathByKey(node.children || [], key, nextPath)
    if (childPath.length > 0) return childPath
  }
  return []
}

const flattenTree = (
  nodes: TreeNode[],
  expandedKeys: Set<string>,
  searchValue: string,
  visibleDepth = 0,
): VisibleNode[] => {
  const result: VisibleNode[] = []
  const keyword = searchValue.trim().toLowerCase()

  nodes.forEach((node) => {
    const matched = !keyword || `${node.name || ''} ${node.type || ''}`.toLowerCase().includes(keyword)
    const childRows = flattenTree(node.children || [], expandedKeys, searchValue, visibleDepth + 1)
    if (matched || childRows.length > 0) {
      result.push({ ...node, visibleDepth })
      if (expandedKeys.has(node.key) || keyword) {
        result.push(...childRows)
      }
    }
  })

  return result
}

const findMaxExclusivePath = (nodes: TreeNode[], path: TreeNode[] = []): TreeNode[] => {
  let bestPath: TreeNode[] = []
  let bestExclusiveTime = -1

  nodes.forEach((node) => {
    const currentPath = [...path, node]
    if (!node.compact && (node.exclusiveTime || 0) > bestExclusiveTime) {
      bestExclusiveTime = node.exclusiveTime || 0
      bestPath = currentPath
    }
    const childPath = findMaxExclusivePath(node.children || [], currentPath)
    const child = childPath[childPath.length - 1]
    if (child && (child.exclusiveTime || 0) > bestExclusiveTime) {
      bestExclusiveTime = child.exclusiveTime || 0
      bestPath = childPath
    }
  })

  return bestPath
}

function SpanDetailStep({ data }: Props) {
  const { t } = useTranslation('pro/causeReport')
  const spanDetail = getSpanDetail(data)
  const details = spanDetail?.details || []
  const summary = spanDetail?.summary
  const [searchValue, setSearchValue] = useState('')
  const [expandedKeys, setExpandedKeys] = useState<Set<string>>(new Set())
  const [selectedKey, setSelectedKey] = useState<string>('')
  const [detailOpenKey, setDetailOpenKey] = useState<string>('')
  const [stackPreviewKey, setStackPreviewKey] = useState<string>('')
  const [restoredNodeKeys, setRestoredNodeKeys] = useState<Set<string>>(new Set())
  const [pendingScrollKey, setPendingScrollKey] = useState<string>('')
  const rowRefs = useRef<Map<string, HTMLDivElement>>(new Map())

  const maxEnd = useMemo(() => {
    return details.reduce((max, item) => {
      const end = (item.offset || 0) + (item.duration || 0)
      return Math.max(max, end)
    }, 0)
  }, [details])

  const compactThreshold = useMemo(() => maxEnd * (DEFAULT_COMPACT_PERCENT / 100), [maxEnd])

  const rawTree = useMemo(() => buildTree(details), [details])

  const tree = useMemo(() => {
    return compactChildren(rawTree, compactThreshold, restoredNodeKeys)
  }, [rawTree, compactThreshold, restoredNodeKeys])

  const topHotspots = useMemo(() => {
    const grouped = new Map<
      string,
      {
        key: string
        name?: string
        type?: string
        count: number
        totalExclusiveTime: number
        maxExclusiveItem: TreeNode
      }
    >()

    details.forEach((item, index) => {
      const node = {
        ...item,
        key: `${item.name || 'span'}-${index}`,
        index: index + 1,
        children: [],
      }
      const groupKey = `${item.type || '-'}:${item.name || '-'}`
      const current = grouped.get(groupKey)
      if (!current) {
        grouped.set(groupKey, {
          key: groupKey,
          name: item.name,
          type: item.type,
          count: 1,
          totalExclusiveTime: item.exclusiveTime || 0,
          maxExclusiveItem: node,
        })
        return
      }
      current.count += 1
      current.totalExclusiveTime += item.exclusiveTime || 0
      if ((item.exclusiveTime || 0) > (current.maxExclusiveItem.exclusiveTime || 0)) {
        current.maxExclusiveItem = node
      }
    })

    return Array.from(grouped.values())
      .filter((item) => item.totalExclusiveTime > 0)
      .sort((a, b) => b.totalExclusiveTime - a.totalExclusiveTime)
      .slice(0, 5)
  }, [details])

  const initialExpandedKeys = useMemo(() => {
    const keys = new Set<string>()
    const visit = (nodes: TreeNode[], level = 0) => {
      nodes.forEach((node) => {
        const isWideParent = (node.duration || 0) > maxEnd * 0.25
        if (level < 1 || (!node.compact && isWideParent && level < 3)) {
          keys.add(node.key)
          visit(node.children || [], level + 1)
        }
      })
    }
    visit(tree)
    findMaxExclusivePath(tree)
      .slice(0, -1)
      .forEach((node) => keys.add(node.key))
    return keys
  }, [maxEnd, tree])

  const mergedExpandedKeys = useMemo(() => {
    return expandedKeys.size > 0 ? expandedKeys : initialExpandedKeys
  }, [expandedKeys, initialExpandedKeys])

  const maxExclusiveKey = useMemo(() => {
    const path = findMaxExclusivePath(tree)
    return path[path.length - 1]?.key || ''
  }, [tree])

  useEffect(() => {
    setSelectedKey(maxExclusiveKey)
    setDetailOpenKey('')
    setStackPreviewKey('')
  }, [maxExclusiveKey])

  const visibleRows = useMemo(() => {
    return flattenTree(tree, mergedExpandedKeys, searchValue)
  }, [tree, mergedExpandedKeys, searchValue])

  useEffect(() => {
    if (!pendingScrollKey) return
    const row = rowRefs.current.get(pendingScrollKey)
    if (!row) return

    window.requestAnimationFrame(() => {
      row.scrollIntoView({ behavior: 'smooth', block: 'center', inline: 'nearest' })
      setPendingScrollKey('')
    })
  }, [pendingScrollKey, visibleRows])

  const selectedNode = useMemo(() => {
    const selectedPath = selectedKey ? findPathByKey(tree, selectedKey) : []
    const selectedInTree = selectedPath[selectedPath.length - 1]
    return (
      selectedInTree ||
      visibleRows.find((row) => row.key === selectedKey) ||
      visibleRows.find((row) => row.key === maxExclusiveKey) ||
      visibleRows[0]
    )
  }, [maxExclusiveKey, selectedKey, tree, visibleRows])

  const selectedStackFrames = useMemo(() => {
    if (!selectedNode) return []
    if (stackPreviewKey !== selectedNode.key) return []
    const sourceKey = selectedNode.compact ? selectedNode.compactSourceKeys?.[0] : selectedNode.key
    const path = sourceKey ? findPathByKey(rawTree, sourceKey) : findPathByKey(tree, selectedNode.key)
    const frames = selectedNode.compact
      ? [...path.slice(0, -1).map((node) => formatStackFrame(node)), ...(selectedNode.compactStack || [])]
      : path.map((node) => formatStackFrame(node))
    if (selectedNode.compact) {
      const compactSourceKeys = new Set(selectedNode.compactSourceKeys || [])
      return frames.map((frame, index) => ({
        ...frame,
        current: frame.key === sourceKey || (index === frames.length - 1 && compactSourceKeys.has(frame.key)),
      }))
    }
    return frames.map((frame) => ({ ...frame, current: frame.key === selectedNode.key }))
  }, [rawTree, selectedNode, stackPreviewKey, tree])

  const selectedSpanInfoItems = useMemo(() => buildSpanInfoItems(selectedNode), [selectedNode])

  const ticks = useMemo(() => {
    return [0, 0.5, 1].map((ratio) => ({
      ratio,
      label: formatDuration(maxEnd * ratio),
    }))
  }, [maxEnd])

  const toggleExpanded = (key: string) => {
    setExpandedKeys((prev) => {
      const next = new Set(prev.size > 0 ? prev : initialExpandedKeys)
      if (next.has(key)) {
        next.delete(key)
      } else {
        next.add(key)
      }
      return next
    })
  }

  const focusSpan = (key: string) => {
    const path = findPathByKey(tree, key)
    if (path.length === 0) {
      setSelectedKey(key)
      return
    }
    const targetKey = path[path.length - 1].key
    setExpandedKeys((prev) => {
      const next = new Set(prev.size > 0 ? prev : initialExpandedKeys)
      path.slice(0, -1).forEach((node) => next.add(node.key))
      return next
    })
    setSelectedKey(targetKey)
    setDetailOpenKey(targetKey)
    setPendingScrollKey(targetKey)
    setStackPreviewKey('')
  }

  const openSpanDetail = (key: string) => {
    const path = findPathByKey(tree, key)
    const targetKey = path[path.length - 1]?.key || key
    setSelectedKey(targetKey)
    setDetailOpenKey(targetKey)
    setStackPreviewKey('')
  }

  const expandCompactParent = (key: string) => {
    const currentRow = visibleRows.find((row) => row.key === key)
    if (!currentRow?.compact || currentRow.compactKind !== 'parent') return
    const path = findPathByKey(tree, key)
    const compactedSourceKeys = currentRow.compactSourceKeys || []
    if (compactedSourceKeys.length > 0) {
      setRestoredNodeKeys((prev) => new Set([...Array.from(prev), ...compactedSourceKeys]))
    }
    setExpandedKeys((prev) => {
      const next = new Set(prev.size > 0 ? prev : initialExpandedKeys)
      path.slice(0, -1).forEach((node) => next.add(node.key))
      compactedSourceKeys.forEach((sourceKey) => next.add(sourceKey))
      return next
    })
    setSelectedKey(compactedSourceKeys[compactedSourceKeys.length - 1] || currentRow.key)
    setDetailOpenKey('')
    setStackPreviewKey('')
  }

  const viewStack = (key: string) => {
    const path = findPathByKey(tree, key)
    const target = path[path.length - 1]
    const targetKey = target?.key || key
    setSelectedKey(targetKey)
    setDetailOpenKey(targetKey)
    setStackPreviewKey((prev) => (prev === (target?.key || key) ? '' : target?.key || key))
  }

  const resetAutoCompact = () => {
    setSearchValue('')
    setExpandedKeys(new Set())
    setRestoredNodeKeys(new Set())
    setStackPreviewKey('')
    setDetailOpenKey('')
    setSelectedKey('')
  }

  if (!spanDetail || details.length === 0) {
    return (
      <Card title={data.report_name}>
        <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} />
      </Card>
    )
  }

  return (
    <Card title={data.report_name} className="span_detail_step">
      <Descriptions size="small" column={2} className="span_detail_summary">
        <Descriptions.Item label={t('spanDetail.traceId', { defaultValue: '调用链 ID' })}>
          {spanDetail.traceId || '-'}
        </Descriptions.Item>
        <Descriptions.Item label={t('spanDetail.spanId', { defaultValue: 'Span ID' })}>
          {spanDetail.spanId || '-'}
        </Descriptions.Item>
        <Descriptions.Item label={t('spanDetail.totalNodes', { defaultValue: 'Span 数量' })}>
          {summary?.totalNodes ?? details.length}
        </Descriptions.Item>
        <Descriptions.Item label={t('spanDetail.maxDepth', { defaultValue: '最大深度' })}>
          {summary?.maxDepth ?? '-'}
        </Descriptions.Item>
        <Descriptions.Item label={t('spanDetail.detailLevel', { defaultValue: '明细级别' })}>
          {getDetailLevelLabel(summary?.detailLevel)}
        </Descriptions.Item>
        <Descriptions.Item label={t('spanDetail.typeCounts', { defaultValue: '类型分布' })}>
          <Space size={[4, 4]} wrap>
            {Object.entries(summary?.typeCounts || {}).map(([key, value]) => (
              <Tag key={key}>
                {getSpanTypeLabel(key)}: {value}
              </Tag>
            ))}
          </Space>
        </Descriptions.Item>
      </Descriptions>

      <div className="span_detail_calltree">
        <div className="span_detail_toolbar">
          <div className="span_detail_filter_group">
            <span className="span_detail_filter_label">
              {t('spanDetail.searchLabel', { defaultValue: '搜索 Span' })}
            </span>
            <Input
              size="small"
              allowClear
              prefix={<SearchOutlined />}
              placeholder={t('spanDetail.searchPlaceholder', { defaultValue: '名称 / 类型' })}
              value={searchValue}
              onChange={(event) => setSearchValue(event.target.value)}
              className="span_detail_search"
            />
          </div>
          <div className="span_detail_toolbar_actions">
            <Button size="small" onClick={resetAutoCompact}>
              {t('spanDetail.autoCompact', { defaultValue: '自动折叠' })}
            </Button>
            <span className="span_detail_count">
              {t('spanDetail.visibleCount', {
                visible: visibleRows.length,
                total: details.length,
                defaultValue: `Span个数：${visibleRows.length}/${details.length}`,
              })}
            </span>
          </div>
        </div>

        <div className="span_detail_hotspots">
          <span className="span_detail_hotspots_title">
            {t('spanDetail.hotspots', { defaultValue: '热点 Span' })}
          </span>
          {topHotspots.map((item, index) => (
            <button
              key={item.key}
              type="button"
              className={`span_detail_hotspot ${
                selectedNode?.name === item.name && selectedNode?.type === item.type ? 'active' : ''
              }`}
              onClick={() => focusSpan(item.maxExclusiveItem.key)}
            >
              <span>{index + 1}</span>
              <strong title={item.name}>{getShortName(item.name)}</strong>
              <small>{item.type || '-'}</small>
              <small>{item.count} 次调用</small>
              <em>{formatDuration(item.totalExclusiveTime)}</em>
            </button>
          ))}
        </div>

        <div className="span_detail_body">
          <div className="span_detail_tree_panel">
            <div className="span_detail_tree_header">
              <span>{t('spanDetail.callTree', { defaultValue: '调用树' })}</span>
              <span>{t('spanDetail.timeAxis', { defaultValue: '调用时间轴' })}</span>
            </div>
            <div className="span_detail_tree_content">
              <div className="span_detail_tree_scroll">
                <div className="span_detail_tree_ruler_spacer" />
                <div className="span_detail_tree_rows">
                  {visibleRows.map((row) => {
                    const hasChildren = row.children && row.children.length > 0
                    const expanded = mergedExpandedKeys.has(row.key) || !!searchValue.trim()
                    const active = selectedNode?.key === row.key
                    const detailOpen = detailOpenKey === row.key && active
                    const stackOpen = detailOpen && stackPreviewKey === row.key && selectedStackFrames.length > 0
                    const detailHeight = stackOpen ? DETAIL_STACK_ROW_HEIGHT : DETAIL_ROW_HEIGHT
                    return (
                      <React.Fragment key={row.key}>
                        <div
                          ref={(element) => {
                            if (element) {
                              rowRefs.current.set(row.key, element)
                            } else {
                              rowRefs.current.delete(row.key)
                            }
                          }}
                          className={`span_detail_tree_row ${active ? 'active' : ''} ${
                            row.compact ? 'compact' : ''
                          } ${detailOpen ? 'detail-open' : ''}`}
                          style={{ height: ROW_HEIGHT }}
                          onClick={() => openSpanDetail(row.key)}
                        >
                          <div className="span_detail_node" style={{ paddingLeft: row.visibleDepth * 20 + 8 }}>
                            <button
                              type="button"
                              className={`span_detail_toggle ${hasChildren ? '' : 'leaf'}`}
                              onClick={(event) => {
                                event.stopPropagation()
                                if (hasChildren) toggleExpanded(row.key)
                              }}
                            >
                              {hasChildren ? (expanded ? '⌄' : '›') : ''}
                            </button>
                            <span className="span_detail_node_main">
                              <span className="span_detail_node_name_line">
                                <button
                                  type="button"
                                  className="span_detail_node_name_button"
                                  onClick={(event) => {
                                    event.stopPropagation()
                                    focusSpan(row.key)
                                  }}
                                >
                                  <span className="span_detail_node_name">
                                    {row.compact
                                      ? row.compactKind === 'parent'
                                        ? t('spanDetail.collapsedParent', {
                                            name: getShortName(row.name),
                                            count: row.compactCount,
                                            defaultValue: `${getShortName(row.name)} 和其他 ${row.compactCount} 个 Span`,
                                          })
                                        : t('spanDetail.collapsedSpans', {
                                            count: row.compactCount,
                                            defaultValue: `${row.compactCount} 个低耗时 Span`,
                                          })
                                      : getShortName(row.name)}
                                  </span>
                                  {active && (
                                    <span className="span_detail_current_badge">
                                      {t('spanDetail.currentSpan', { defaultValue: '当前' })}
                                    </span>
                                  )}
                                </button>
                                <span className="span_detail_node_actions">
                                  {row.compact && row.compactKind === 'parent' && (
                                    <>
                                      <button
                                        type="button"
                                        onClick={(event) => {
                                          event.stopPropagation()
                                          expandCompactParent(row.key)
                                        }}
                                      >
                                        {t('spanDetail.expandCallTree', { defaultValue: '展开' })}
                                      </button>
                                    </>
                                  )}
                                </span>
                              </span>
                              <span className="span_detail_node_meta">
                                {row.compact ? t('spanDetail.lowCostGroup', { defaultValue: '低耗时合并' }) : row.type || '-'} ·{' '}
                                {formatDuration(row.duration)} · {t('spanDetail.exclusiveTime')}{' '}
                                {formatDuration(row.exclusiveTime)}
                              </span>
                            </span>
                          </div>
                        </div>
                        {detailOpen && (
                          <div className="span_detail_inline_detail_row" style={{ minHeight: detailHeight }}>
                            <div
                              className="span_detail_inline_detail"
                              style={{ marginLeft: row.visibleDepth * 20 + 40 }}
                            >
                              <div className="span_detail_inline_detail_header">
                                <span>{t('spanDetail.spanInfo', { defaultValue: 'Span 详情' })}</span>
                                <button type="button" onClick={() => setDetailOpenKey('')}>
                                  {t('spanDetail.closeDetail', { defaultValue: '关闭' })}
                                </button>
                              </div>
                              <div className="span_detail_info_grid">
                                {selectedSpanInfoItems.map((item) => (
                                  <div key={item.label} className="span_detail_info_item">
                                    <span>{item.label}</span>
                                    <strong title={`${item.value}`}>{item.value}</strong>
                                  </div>
                                ))}
                              </div>
                              <div className="span_detail_stack_section">
                                <button type="button" onClick={() => viewStack(row.key)}>
                                  {stackOpen
                                    ? t('spanDetail.closeStack', { defaultValue: '收起堆栈' })
                                    : t('spanDetail.viewStack', { defaultValue: '显示堆栈' })}
                                </button>
                                {stackOpen && (
                                  <ol className="span_detail_stack_frames">
                                    {selectedStackFrames.map((frame, frameIndex) => (
                                      <li
                                        key={`${frame.key}-${frameIndex}`}
                                        className={`${frame.current ? 'current' : ''} ${frame.compacted ? 'compacted' : ''}`}
                                      >
                                        <span className="span_detail_stack_index">{frameIndex + 1}</span>
                                        <span className="span_detail_stack_text" title={frame.text}>
                                          {frame.text}
                                        </span>
                                        {frame.type && <span className="span_detail_stack_type">({frame.type})</span>}
                                        {frame.current && (
                                          <span className="span_detail_stack_current">
                                            {t('spanDetail.currentSpan', { defaultValue: '当前' })}
                                          </span>
                                        )}
                                      </li>
                                    ))}
                                  </ol>
                                )}
                              </div>
                            </div>
                          </div>
                        )}
                      </React.Fragment>
                    )
                  })}
                </div>
              </div>

              <div className="span_detail_timeline_fixed">
                <div className="span_detail_ruler">
                  {ticks.map((tick) => (
                    <span key={tick.ratio} style={{ left: `${tick.ratio * 100}%` }}>
                      {tick.label}
                    </span>
                  ))}
                </div>
                <div className="span_detail_timeline_rows">
                  {visibleRows.map((row) => {
                    const left = maxEnd > 0 ? ((row.offset || 0) / maxEnd) * 100 : 0
                    const width = maxEnd > 0 ? Math.max(((row.duration || 0) / maxEnd) * 100, 0.7) : 0
                    const exclusiveWidth =
                      maxEnd > 0 ? Math.max(((row.exclusiveTime || 0) / maxEnd) * 100, 0.7) : 0
                    const color = TYPE_COLORS[row.type || ''] || '#8c8c8c'
                    const active = selectedNode?.key === row.key
                    const detailOpen = detailOpenKey === row.key && active
                    const stackOpen = detailOpen && stackPreviewKey === row.key && selectedStackFrames.length > 0
                    const detailHeight = stackOpen ? DETAIL_STACK_ROW_HEIGHT : DETAIL_ROW_HEIGHT
                    return (
                      <React.Fragment key={row.key}>
                        <div
                          className={`span_detail_timeline_row ${active ? 'active' : ''} ${
                            detailOpen ? 'detail-open' : ''
                          }`}
                          style={{ height: ROW_HEIGHT }}
                          onClick={() => openSpanDetail(row.key)}
                        >
                          <div className="span_detail_chart">
                            <div
                              className="span_detail_projected_bar"
                              style={{ left: `${left}%`, width: `${width}%`, background: color }}
                            />
                            <div
                              className="span_detail_exclusive_bar"
                              style={{ left: `${left}%`, width: `${exclusiveWidth}%` }}
                            />
                          </div>
                        </div>
                        {detailOpen && (
                          <div className="span_detail_timeline_detail_gap" style={{ minHeight: detailHeight }} />
                        )}
                      </React.Fragment>
                    )
                  })}
                </div>
              </div>
            </div>
          </div>

        </div>
      </div>
    </Card>
  )
}

export default SpanDetailStep
