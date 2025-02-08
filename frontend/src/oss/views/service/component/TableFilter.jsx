/**
 * Copyright 2024 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import { useState, useEffect } from 'react'
import {
  getServiceListApi,
  getNamespacesApi,
  getServiceEndpointNameApi,
} from 'src/core/api/service'
import { useSelector } from 'react-redux'
import { selectSecondsTimeRange } from 'src/core/store/reducers/timeRangeReducer'
import { Select } from 'antd'
import { getStep } from 'src/core/utils/step'
import { useTranslation } from 'react-i18next'
import { getDatasourceByGroupApi } from 'src/core/api/dataGroup'

export const TableFilter = (props) => {
  const { t } = useTranslation('oss/service')
  const { setServiceName, setEndpoint, setNamespace, groupId } = props
  const [serviceNameOptions, setServiceNameOptions] = useState([])
  const [endpointNameOptions, setEndpointNameOptions] = useState([])
  const [namespaceOptions, setNamespaceOptions] = useState([])
  const [serachServiceName, setSerachServiceName] = useState(null)
  const [serachEndpointName, setSerachEndpointName] = useState(null)
  const [serachNamespace, setSerachNamespace] = useState(null)
  const [datasource, setDatasource] = useState()

  //根据选定的服务名称获取并设置端点名称选项。
  const getEndpointNameOptions = () => {
    const endpoints = []
    const endpointsSet = new Set([])
    const filterOptions =
      serachServiceName?.length > 0
        ? serviceNameOptions.filter((service) => serachServiceName.includes(service.label))
        : serviceNameOptions

    filterOptions.map((option) => {
      endpoints.push({
        label: option.label,
        title: option.label,
        options: datasource?.serviceMap[option.label]?.map((item) => {
          endpointsSet.add(item)
          return {
            label: <span>{item}</span>,
            value: item,
          }
        }),
      })
    })
    setSerachEndpointName(serachEndpointName?.filter((endpoint) => endpointsSet.has(endpoint)))
    setEndpointNameOptions(endpoints)
  }
  const onChangeNamespace = (event) => {
    setSerachNamespace(event)
    setNamespaceOptions(
      event.serviceMap.map((service) => ({
        label: service,
        value: service,
      })),
    )
  }
  //处理服务名称选择更改。
  const onChangeServiceName = (event) => {
    setSerachServiceName(event)
  }

  useEffect(() => {
    if (serachNamespace && serachNamespace?.length > 0) {
      const services = []
      serachNamespace.map((namespace) => {
        datasource?.namespaceMap[namespace]?.map((service) => {
          services.push({
            label: service,
            value: service,
          })
        })
      })
      setServiceNameOptions(services)
    } else {
      setServiceNameOptions(
        Object.entries(datasource?.serviceMap || []).map(([service, endpoints]) => ({
          label: service,
          value: service,
          endpoints: endpoints,
        })),
      )
    }
  }, [serachNamespace])
  const getDatasourceByGroup = () => {
    getDatasourceByGroupApi({
      groupId: groupId,
      category: 'apm',
    }).then((res) => {
      //todo null
      const namespaceOptions = Object.entries(res.namespaceMap).map(([namespace, service]) => ({
        label: namespace,
        value: namespace,
      }))
      const serviceOptions = Object.entries(res.serviceMap || []).map(([service, endpoints]) => ({
        label: service,
        value: service,
        endpoints: endpoints,
      }))
      setDatasource(res)
      setNamespaceOptions(namespaceOptions)
      setServiceNameOptions(serviceOptions)
    })
  }

  useEffect(() => {
    getDatasourceByGroup()
  }, [groupId])

  useEffect(() => {
    getEndpointNameOptions()
  }, [serviceNameOptions, serachServiceName])
  useEffect(() => {
    setNamespace(serachNamespace)
  }, [serachNamespace])
  useEffect(() => {
    setServiceName(serachServiceName)
  }, [serachServiceName])
  useEffect(() => {
    setEndpoint(serachEndpointName)
  }, [serachEndpointName])

  return (
    <>
      <div className="mb-2 flex flex-row w-full">
        <div className="flex flex-row items-center mr-5 text-sm min-w-[280px]">
          <span className="text-nowrap">{t('tableFilter.namespacesLabel')}：</span>
          <Select
            mode="multiple"
            allowClear
            id="namespace"
            className="w-full"
            placeholder={t('tableFilter.namespacePlaceholder')}
            value={serachNamespace}
            onChange={onChangeNamespace}
            options={namespaceOptions}
            maxTagCount={2}
            maxTagPlaceholder={(omittedValues) => `+${omittedValues.length}...`}
          />
        </div>
        <div className="flex flex-row items-center mr-5 text-sm min-w-[280px]">
          <span className="text-nowrap">{t('tableFilter.applicationsLabel')}：</span>
          <Select
            mode="multiple"
            allowClear
            className="w-full"
            id="serviceName"
            placeholder={t('tableFilter.applicationsPlaceholder')}
            value={serachServiceName}
            onChange={onChangeServiceName}
            options={serviceNameOptions}
            popupMatchSelectWidth={false}
            maxTagCount={2}
            maxTagPlaceholder={(omittedValues) => `+${omittedValues.length}...`}
          />
        </div>
        <div className="flex flex-row items-center mr-5 text-sm min-w-[280px]">
          <span className="text-nowrap">{t('tableFilter.endpointsLabel')}：</span>
          <Select
            mode="multiple"
            id="endpointName"
            placeholder={t('tableFilter.endpointsPlaceholder')}
            className="w-full"
            value={serachEndpointName}
            popupMatchSelectWidth={false}
            onChange={(e) => setSerachEndpointName(e)}
            options={endpointNameOptions}
            maxTagCount={2}
            maxTagPlaceholder={(omittedValues) => `+${omittedValues.length}...`}
            allowClear
          />
        </div>
        <div>{/* <ThresholdCofigModal /> */}</div>
      </div>
    </>
  )
}
