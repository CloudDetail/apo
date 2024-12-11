import _ from 'lodash'
import React, { createContext, useContext, useState, ReactNode, useMemo, useEffect } from 'react'
import { useLocation, useParams, useSearchParams } from 'react-router-dom'
import {
  getErrorFlowsApi,
  getErrortracesApi,
  getFlowsApi,
  getMutatedtracesApi,
} from 'src/pro/api/cause'

// 定义上下文的数据类型
interface CauseReportType {
  traceId: string
  spanId: string
  category: string
  isTimeOut: string
  entryTrace: any
  topologyData: any
  topologyTreeData: any
  reportData: any
  modalVisible: string
  loading: string
  relation: string
  reports: string
  uploadTraceId: string
  node: any
  causeNode: any
  dataMissing: boolean
  // 方法
  setNode: any
}

// 创建上下文
const CauseReportContext = createContext<CauseReportType | undefined>(undefined)

// 创建自定义 Hook
export const useCauseReportContext = () => useContext(CauseReportContext)
// 创建 Provider 组件
export const CauseReportProvider: React.FC<{ children: ReactNode }> = ({ children }) => {
  const [searchParams] = useSearchParams()
  const { traceId, spanId } = useParams()
  const [category, setCategory] = useState<string>(null)
  const [isTimeOut, setIsTimeOut] = useState<boolean>(false)
  const [entryTrace, setEntryTrace] = useState<any>({})
  //故障根因节点
  const [causeNode, setCauseNode] = useState<any>({})
  const [topologyData, setTopologyData] = useState<any>({})
  const [topologyTreeData, setTopologyTreeData] = useState({})
  const [reportData, setReportData] = useState<any>({})
  const [modalVisible, setModalVisible] = useState<boolean>(false)
  const [loading, setLoading] = useState<boolean>(true)
  const [dataMissing, setDataMissing] = useState<boolean>(false)

  const [node, setNode] = useState<any>()
  // 导入用
  const [relation, setRelation] = useState(null)
  const [reports, setReports] = useState(null)
  const [uploadTraceId, setUploadTraceId] = useState(null)
  // 从url中获取数据
  const location = useLocation()

  const handleTopologyData = (nodes: any[], edges: any[], node: any, depth = 0) => {
    if (node.isPath) {
      let newNode: any = { ...node, depth, chartId: node.id + '-' + depth }
      delete newNode.children
      nodes.push(newNode)
    }
    // 缺省肯定不会是第一个或最后一个节点
    if (node.missVNode) {
      edges.push({
        source: `missingUpstream-${node.id}` + '-' + depth,
        target: node.id + '-' + (depth + 1),
        missVNode: true,
      })
    }
    if (node.children && node.children.length > 0) {
      if (_.some(node.children, (node) => node.isPath)) {
        let subNode = _.find(node.children, { isPath: true })

        edges.push({ source: node.id + '-' + depth, target: subNode.id + '-' + (depth + 1) })
        handleTopologyData(nodes, edges, subNode, depth + 1)
      }
    }
  }
  const prepareTopologyData = (data: any) => {
    const nodes: any[] = [],
      edges: any[] = []
    handleTopologyData(nodes, edges, data)
    let spanId = ''
    const nodesWithUpstream: any[] = []
    _.forEach(nodes, (node: any) => {
      if (node.isMutated) {
        node.is_warn = true
        spanId = node.spanId
      }
      if (node.missVNode) {
        const missingUpstreamNode = {
          id: `missingUpstream-${node.id}`,
          label: 'Missing Upstream',
          isMissing: true,
        }
        nodesWithUpstream.push(missingUpstreamNode)
      }

      // 添加当前节点到nodesWithUpstream中
      nodesWithUpstream.push(node)
    })
    setTopologyTreeData(_.cloneDeep(data))
    setEntryTrace(nodes[0])
    setCauseNode(nodes[nodes?.length - 1])
    setTopologyData({ nodes: nodesWithUpstream, edges })
    return spanId
  }
  // const prepareTopologyDataForError = (data: any) => {
  //   const nodes: any[] = [],
  //     edges: any[] = []
  //   handleTopologyData(nodes, edges, data)
  //   _.forEach(nodes, (node: any) => {
  //     if (node.isMutated) {
  //       node.is_warn = true
  //     }
  //   })
  //   setTopologyTreeData(_.cloneDeep(data))
  //   setEntryTrace(nodes[0])
  //   setTopologyData({ nodes: nodesWithUpstream, edges })
  // }
  function fetchSlowData(mutatedType) {
    setLoading(true)
    setDataMissing(false)
    // 定义各个请求的 Promise
    const mutatedTracesPromise = getMutatedtracesApi(traceId, spanId)
      .then((res) => {
        if (res.success && res.data) {
          prepareTopologyData(res.data)
        } else if (!res.success && !res.data) {
          setDataMissing(true)
        }
      })
      .catch((err) => {
        console.error('Mutated traces error:', err)
      })

    const flowsPromise = getFlowsApi(traceId, spanId, mutatedType)
      .then((res) => {
        if (res.status === 'failure' && res.message.includes('timeout')) {
          setIsTimeOut(true)
        } else if (res.status !== 'failure') {
          setReportData(res)
          setIsTimeOut(false)
        }
      })
      .catch((err) => {
        console.error('Flows error:', err)
      })

    // 使用 Promise.allSettled 等待所有请求完成后更新 loading
    Promise.allSettled([mutatedTracesPromise, flowsPromise]).finally(() => {
      setLoading(false)
    })
  }
  function getErrorTraces() {
    getErrortracesApi(traceId)
      .then((res) => {
        const spanId = prepareTopologyData(res.data)
        return getErrorFlowsApi(traceId, spanId)
      })
      .then(async (res) => {
        if (res.status === 'failure' && res.message.includes('timeout')) {
          setIsTimeOut(true)
        } else if (res.status !== 'failure') {
          setReportData(res)
          setIsTimeOut(false)
        }
      })
      .catch((err) => {
        console.error('Error loading data:', err)
      })
      .finally(() => {
        setLoading(false)
      })
  }
  useEffect(() => {
    let category = 'error'
    if (location.pathname.startsWith('/cause/report')) {
      category = 'slow'
    }
    setCategory(category)
    let mutatedType = searchParams.get('mutatedType') || ''
    if (traceId) {
      if (category === 'slow') {
        if (spanId) {
          fetchSlowData(mutatedType)
        }
      } else {
        getErrorTraces()
      }
    }
  }, [spanId, traceId, searchParams, location])
  useEffect(() => { }, [location])
  const memoizedValue = useMemo(
    () => ({
      traceId,
      category,
      isTimeOut,
      entryTrace,
      topologyData,
      topologyTreeData,
      reportData,
      modalVisible,
      loading,
      relation,
      reports,
      uploadTraceId,
      spanId,
      node,
      causeNode,
      dataMissing,
      setNode,
    }),
    [
      traceId,
      category,
      isTimeOut,
      entryTrace,
      topologyData,
      topologyTreeData,
      reportData,
      modalVisible,
      loading,
      relation,
      reports,
      uploadTraceId,
      spanId,
      node,
      causeNode,
    ],
  )
  return <CauseReportContext.Provider value={memoizedValue}>{children}</CauseReportContext.Provider>
}
