/**
 * Copyright 2024 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import { useState, useEffect } from 'react'

import { useTranslation } from 'react-i18next'
import FilterSelector from './FilterSelector'
import React from 'react'
import { useSelector } from 'react-redux'
import { getDatasourceFilterApiV2 } from 'src/core/api/dataGroup'
import { useDebounce } from 'react-use'
type NodeType = 'cluster' | 'namespace' | 'service' | 'endpoint' | 'pod' | 'instance'
interface InstanceInfo {
  name: string
  node: string
  pid: string
  pod: string
  id?: string
  containerId: string
}

interface TreeNode extends Partial<InstanceInfo> {
  id?: string
  name: string
  type: NodeType
  children?: TreeNode[]
  extraChildren?: TreeNode[]
}

interface OptionItem {
  label: string
  value: string
  path: string[] // 父节点名称路径
  idPath: string[] // 父节点 ID 路径
  info?: InstanceInfo
}

interface GroupedList {
  parentId: string
  idPath: string[]
  label: string // 路径字符串：例如 dev-6 -> train-ticket
  options: OptionItem[]
}

interface TreeExtractedResult {
  clusters: OptionItem[]
  namespaces: GroupedList[]
  services: GroupedList[]
  endpoints: GroupedList[]
  pods: GroupedList[]
  instances: GroupedList[]
}

export function extractTreeData(nodes: TreeNode[]): TreeExtractedResult {
  const clusters: OptionItem[] = []
  const namespacesMap = new Map<string, GroupedList>()
  const servicesMap = new Map<string, GroupedList>()
  const endpointsMap = new Map<string, GroupedList>()
  const podsMap = new Map<string, GroupedList>()
  const instancesMap = new Map<string, GroupedList>()
  function makeLabel(path: string[]): string {
    return path.join(' / ')
  }

  function visit(node: TreeNode, path: string[], idPath: string[], parent?: TreeNode) {
    const currentPath = [...path, node.name]
    const currentIdPath = [...idPath, node.id || node.name]
    const value = node.id || path.join('-') + node.name

    const item: OptionItem = {
      label: node.name,
      value,
      path: currentPath,
      idPath: currentIdPath,
    }
    const parentId = parent?.id || ''
    const groupLabel = makeLabel(path)

    switch (node.type) {
      case 'cluster':
        clusters.push(item)
        break
      case 'namespace':
        if (!namespacesMap.has(parentId)) {
          namespacesMap.set(parentId, {
            parentId: parentId,
            idPath: idPath,
            label: groupLabel,
            options: [],
          })
        }
        namespacesMap.get(parentId)!.options.push(item)
        break
      case 'service':
        if (!servicesMap.has(parentId)) {
          servicesMap.set(parentId, {
            parentId: parentId,
            idPath: idPath,
            label: groupLabel,
            options: [],
          })
        }
        servicesMap.get(parentId)!.options.push(item)
        break
      case 'endpoint':
        if (!endpointsMap.has(parentId)) {
          endpointsMap.set(parentId, {
            parentId: parentId,
            idPath: idPath,
            label: groupLabel,
            options: [],
          })
        }
        endpointsMap.get(parentId)!.options.push(item)
        break
      case 'pod':
        if (!podsMap.has(parentId)) {
          podsMap.set(parentId, {
            parentId: parentId,
            idPath: idPath,
            label: groupLabel,
            options: [],
          })
        }
        podsMap.get(parentId)!.options.push(item)
        break
      case 'instance':
        if (!instancesMap.has(parentId)) {
          instancesMap.set(parentId, {
            parentId: parentId,
            idPath: idPath,
            label: groupLabel,
            options: [],
          })
        }
        console.log(node)
        item.info = {
          name: node.name,
          node: node.node,
          pid: node.pid,
          pod: node.pod,
          containerId: node.containerId,
          id: node.id,
        }
        instancesMap.get(parentId)!.options.push(item)
        break
    }

    ;(node.children || node.extraChildren)?.forEach((child) =>
      visit(child, currentPath, currentIdPath, node),
    )
  }

  nodes.forEach((root) => visit(root, [], []))

  return {
    clusters,
    namespaces: Array.from(namespacesMap.values()),
    services: Array.from(servicesMap.values()),
    endpoints: Array.from(endpointsMap.values()),
    pods: Array.from(podsMap.values()),
    instances: Array.from(instancesMap.values()),
  }
}
// cluster
// —-namespace
// ----service
// ------endpoint
// ------pod //暂无用
// ------instance

interface DataSourceFilterProps {
  setCluster?: (cluster: string[]) => void
  setServiceName?: (serviceName: string[]) => void
  setPod?: (pod: string[]) => void
  setEndpoint?: (endpoint: string[]) => void
  setNamespace?: (namespace: string[]) => void
  category: 'log' | 'apm'
  extra: 'endpoint' | 'instance'
  className?: string
  initCluster?: string[]
  initNamespace?: string[]
  initServiceName?: string[]
  initEndpointName?: string[]
  initPod?: string[]
  initInstance?: string[]
  startTime: number | null
  endTime: number | null
  setIsFilterDone?: (isFilterDone: boolean) => void
  setInstance?: (instance: object[]) => void
}
const DataSourceFilter = (props: DataSourceFilterProps) => {
  const { t } = useTranslation('oss/service')
  const {
    setCluster,
    setServiceName,
    setPod,
    setEndpoint,
    setNamespace,
    category,
    extra,
    className = '',
    initCluster,
    initNamespace,
    initServiceName,
    initEndpointName,
    initPod,
    startTime,
    endTime,
    setIsFilterDone,
    setInstance,
  } = props
  const { dataGroupId } = useSelector((state: any) => state.dataGroupReducer)
  const [serviceNameOptions, setServiceNameOptions] = useState([])
  const [endpointNameOptions, setEndpointNameOptions] = useState([])
  const [namespaceOptions, setNamespaceOptions] = useState([])
  const [searchServiceName, setSearchServiceName] = useState([])
  const [searchEndpointName, setSearchEndpointName] = useState([])
  const [searchNamespace, setSearchNamespace] = useState([])
  const [searchCluster, setSearchCluster] = useState([])
  const [clusterOptions, setClusterOptions] = useState([])
  const [podOptions, setPodOptions] = useState([])
  const [searchPods, setSearchPods] = useState([])
  const [searchInstances, setSearchInstances] = useState([])
  const [instanceOptions, setInstanceOptions] = useState([])
  const [treeData, setTreeData] = useState<TreeExtractedResult>({
    clusters: [],
    namespaces: [],
    services: [],
    endpoints: [],
    pods: [],
    instances: [],
  })
  const getDatasourceByGroup = () => {
    getDatasourceFilterApiV2({
      groupId: dataGroupId,
      category: category,
      startTime,
      endTime,
      extra,
    }).then((res) => {
      console.log(res)
      const treeData = extractTreeData(res.view?.children || [])
      console.log(treeData)
      setTreeData(treeData)
      setClusterOptions(treeData.clusters)
      setNamespaceOptions(treeData.namespaces)
      setServiceNameOptions(treeData.services)
      setEndpointNameOptions(treeData.endpoints)
      setPodOptions(treeData.pods)
      setInstanceOptions(treeData.instances)
    })
  }
  useDebounce(
    () => {
      if (dataGroupId != null && startTime != null && endTime != null) {
        setIsFilterDone?.(false)
        getDatasourceByGroup()
      }
    },
    300,
    [dataGroupId, startTime, endTime],
  )
  const filterAndSyncSearch = (options, searchValues, getValue = (item) => item.value) => {
    const allValues = options.flatMap((group) => group.options.map(getValue))
    const filtered = (searchValues || []).filter((val) => allValues.includes(val))
    const newSearch =
      filtered.length !== (searchValues?.length || 0) ? [...filtered] : searchValues || []
    return { allValues, newSearch }
  }
  useEffect(() => {
    updateFilteredOptions(
      searchCluster,
      searchNamespace,
      searchServiceName,
      searchEndpointName,
      searchPods,
      searchInstances,
    )
  }, [treeData])
  const updateFilteredOptions = (
    searchCluster,
    searchNamespace,
    searchServiceName,
    searchEndpointName,
    searchPods,
    searchInstances,
  ) => {
    // 1. 过滤 namespace
    const namespaces = treeData.namespaces.filter(
      (item) => !searchCluster?.length || searchCluster.includes(item.idPath[0]),
    )
    const { newSearch: newSearchNamespace } = filterAndSyncSearch(namespaces, searchNamespace)

    // 2. 过滤 service
    const namespaceIds = namespaces.flatMap((group) => group.options).map((item) => item.value)
    let services = treeData.services.filter((item) => namespaceIds.includes(item.idPath[1]))
    if (newSearchNamespace.length > 0) {
      services = services.filter((item) => newSearchNamespace.includes(item.idPath[1]))
    }
    const { newSearch: newSearchServiceName } = filterAndSyncSearch(services, searchServiceName)
    const servicesIds = services.flatMap((group) => group.options).map((item) => item.value)

    // 3. 过滤 endpoint
    let endpoints = treeData.endpoints.filter((item) => servicesIds.includes(item.idPath[2]))
    if (newSearchServiceName.length > 0) {
      endpoints = endpoints.filter((item) => newSearchServiceName.includes(item.idPath[2]))
    }
    const { newSearch: newSearchEndpointName } = filterAndSyncSearch(endpoints, searchEndpointName)
    // 4. 过滤 pod
    let pods = treeData.pods.filter((item) => servicesIds.includes(item.idPath[2]))
    if (newSearchServiceName.length > 0) {
      pods = pods.filter((item) => newSearchServiceName.includes(item.idPath[2]))
    }
    const { newSearch: newSearchPods } = filterAndSyncSearch(pods, searchPods)
    // 5. 过滤 instance
    let instances = treeData.instances.filter((item) => servicesIds.includes(item.idPath[2]))
    if (newSearchServiceName.length > 0) {
      instances = instances.filter((item) => newSearchServiceName.includes(item.idPath[2]))
    }
    const { newSearch: newSearchInstances } = filterAndSyncSearch(instances, searchInstances)
    // 6. set
    setNamespaceOptions(namespaces)
    setServiceNameOptions(services)
    setEndpointNameOptions(endpoints)
    setPodOptions(pods)
    setSearchCluster(searchCluster)
    setSearchNamespace(newSearchNamespace)
    setSearchServiceName(newSearchServiceName)
    setSearchEndpointName(newSearchEndpointName)
    setSearchPods(newSearchPods)
    setSearchInstances(newSearchInstances)
  }
  function getUniqueLabelsByValues(
    options: GroupedList[],
    values: string[],
  ): { labels: string[]; infos: string[] } {
    const labelSet = new Set<string>()
    const infoSet = new Set<string>()
    const valueSet = new Set(values)

    for (const group of options) {
      for (const opt of group.options) {
        if (valueSet.has(opt.value)) {
          labelSet.add(opt.label)
          if (opt.info) {
            // 处理 undefined
            infoSet.add(opt.info)
          }
        }
      }
    }

    return {
      labels: Array.from(labelSet),
      infos: Array.from(infoSet),
    }
  }
  useEffect(() => {
    setCluster?.(searchCluster)
  }, [searchCluster, clusterOptions])

  useEffect(() => {
    const { labels } = getUniqueLabelsByValues(namespaceOptions, searchNamespace)
    setNamespace?.(labels)
  }, [searchNamespace])

  useEffect(() => {
    const { labels } = getUniqueLabelsByValues(serviceNameOptions, searchServiceName)
    setServiceName?.(labels)
  }, [searchServiceName])

  useEffect(() => {
    const { labels } = getUniqueLabelsByValues(endpointNameOptions, searchEndpointName)
    setEndpoint?.(labels)
  }, [searchEndpointName])

  useEffect(() => {
    const { labels } = getUniqueLabelsByValues(podOptions, searchPods)
    setPod?.(labels)
  }, [searchPods])
  useEffect(() => {
    const { infos } = getUniqueLabelsByValues(instanceOptions, searchInstances)
    setInstance?.(infos)
    console.log(infos)
  }, [searchInstances])

  useEffect(() => {
    if (initCluster) {
      setSearchCluster(initCluster)
    }
  }, [initCluster])
  useEffect(() => {
    if (initNamespace) {
      setSearchNamespace(initNamespace)
    }
  }, [initNamespace])
  useEffect(() => {
    if (initServiceName) {
      setSearchServiceName(initServiceName)
    }
  }, [initServiceName])
  useEffect(() => {
    if (initEndpointName) {
      setSearchEndpointName(initEndpointName)
    }
  }, [initEndpointName])
  useEffect(() => {
    if (initPod) {
      setSearchPods(initPod)
    }
  }, [initPod])
  useDebounce(
    () => {
      setIsFilterDone?.(true)
    },
    300,
    [
      clusterOptions,
      namespaceOptions,
      serviceNameOptions,
      endpointNameOptions,
      podOptions,
      searchCluster,
      searchNamespace,
      searchServiceName,
      searchEndpointName,
      searchPods,
    ],
  )
  return (
    <>
      <div className={`flex flex-row w-full ${className} flex-wrap`}>
        <FilterSelector
          mode={category === 'log' ? null : 'multiple'}
          label={t('tableFilter.clusterLabel')}
          placeholder={t('tableFilter.clusterPlaceholder')}
          value={searchCluster}
          onChange={(e) => {
            console.log(e)
            updateFilteredOptions(
              e,
              searchNamespace,
              searchServiceName,
              searchEndpointName,
              searchPods,
              searchInstances,
            )
          }}
          options={clusterOptions}
          id="cluster"
        />

        <FilterSelector
          mode={category === 'log' ? null : 'multiple'}
          label={t('tableFilter.namespacesLabel')}
          placeholder={t('tableFilter.namespacePlaceholder')}
          value={searchNamespace}
          onChange={(e) => {
            updateFilteredOptions(
              searchCluster,
              e,
              searchServiceName,
              searchEndpointName,
              searchPods,
              searchInstances,
            )
          }}
          options={namespaceOptions}
          id="namespace"
        />
        <FilterSelector
          label={t('tableFilter.applicationsLabel')}
          placeholder={t('tableFilter.applicationsPlaceholder')}
          value={searchServiceName}
          onChange={(e) => {
            updateFilteredOptions(
              searchCluster,
              searchNamespace,
              e,
              searchEndpointName,
              searchPods,
              searchInstances,
            )
          }}
          options={serviceNameOptions}
          id="serviceName"
        />
        {extra === 'endpoint' ? (
          <>
            {' '}
            <FilterSelector
              label={t('tableFilter.endpointsLabel')}
              placeholder={t('tableFilter.endpointsPlaceholder')}
              value={searchEndpointName}
              onChange={(e) => setSearchEndpointName(e)}
              options={endpointNameOptions}
              id="endpointName"
            />
          </>
        ) : (
          // <FilterSelector
          //   mode={null}
          //   label={t('tableFilter.podsLabel')}
          //   placeholder={t('tableFilter.podsPlaceholder')}
          //   value={searchPods}
          //   onChange={(e) => setSearchPods(e)}
          //   options={podOptions}
          //   id="podName"
          // />
          <FilterSelector
            mode={null}
            label={t('tableFilter.instancesLabel')}
            placeholder={t('tableFilter.instancesPlaceholder')}
            value={searchInstances}
            onChange={(e) => setSearchInstances(e)}
            options={instanceOptions}
            id="instanceName"
          />
        )}

        <div>{/* <ThresholdCofigModal /> */}</div>
      </div>
    </>
  )
}
export default React.memo(DataSourceFilter)
