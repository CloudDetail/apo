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
import { getDatasourceByGroup, getDatasourceByGroupApi } from 'src/core/api/dataGroup'

export const TableFilter = (props) => {
  const { t } = useTranslation('oss/service')
  const { setServiceName, setEndpoint, setNamespace, groupId } = props
  const [serviceNameOptions, setServiceNameOptions] = useState([])
  const [endpointNameOptions, setEndpointNameOptions] = useState([])
  const [namespaceOptions, setNamespaceOptions] = useState([])
  const [serachServiceName, setSerachServiceName] = useState(null)
  const [serachEndpointName, setSerachEndpointName] = useState(null)
  const [serachNamespace, setSerachNamespace] = useState(null)
  const [prevSearchServiceName, setPrevSearchServiceName] = useState(null)
  const [datasource, setDatasource] = useState()
  const { startTime, endTime } = useSelector(selectSecondsTimeRange)

  //根据选定的服务名称获取并设置端点名称选项。
  const getEndpointNameOptions = (serviceNameList) => {
    setEndpointNameOptions([])
    Promise.all(
      serviceNameList?.map((element) => {
        const params = {
          startTime,
          endTime,
          step: getStep(startTime, endTime),
          serviceName: element,
          sortRule: 1,
        }
        return getServiceEndpointNameApi(params).then((data) => ({
          label: <span>{element}</span>,
          title: element,
          options: data.map((item) => ({
            label: <span>{item?.endpoint}</span>,
            value: item?.endpoint,
          })),
        }))
      }),
    )
      .then((newEndpointNameOptions) => setEndpointNameOptions(newEndpointNameOptions))
      .catch((error) => console.error('获取 endpoint 失败:', error))
  }
  const onChangeNamespace = (event) => {
    setSerachNamespace(event)
    setNamespaceOptions(
      event.serviceList.map((service) => ({
        label: service,
        value: service,
      })),
    )
  }
  //处理服务名称选择更改。
  const onChangeServiceName = (event) => {
    setPrevSearchServiceName(serachServiceName)
    setSerachServiceName(event)
  }
  //移除不再相关的端点名称。
  const removeEndpointNames = () => {
    if (prevSearchServiceName?.length > serachServiceName?.length) {
      const removeServiceNameSet = new Set(
        prevSearchServiceName.filter((item) => !serachServiceName.includes(item)),
      )
      const removeEndpoint = endpointNameOptions
        .flatMap((item) => (removeServiceNameSet.has(item.title) ? item.options : []))
        .map((item) => item.value)
      const removedSearchedName = serachEndpointName?.filter(
        (item) => !removeEndpoint?.includes(item),
      )
      setSerachEndpointName(removedSearchedName)
    }
    getEndpointNameOptions(serachServiceName)
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
        (datasource?.serviceList || []).map((service) => ({
          label: service,
          value: service,
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
      const namespaceOptions = Object.entries(res.namespaceMap).map(([namespace, serviceList]) => ({
        label: namespace,
        value: namespace,
      }))
      const serviceOption = (res.serviceList || []).map((service) => ({
        label: service,
        value: service,
      }))
      setDatasource(res)
      setNamespaceOptions(namespaceOptions)
      setServiceNameOptions(serviceOption)
    })
  }

  useEffect(() => {
    if (startTime && endTime) {
      getDatasourceByGroup()
      // Promise.all([getServiceNameOptions(), getNamespaceOptions()])
    }
  }, [startTime, endTime, groupId])

  useEffect(() => {
    if (serachServiceName) {
      removeEndpointNames()
    }
  }, [serachServiceName])
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
