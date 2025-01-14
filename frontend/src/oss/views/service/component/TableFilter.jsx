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

export const TableFilter = (props) => {
  const { setServiceName, setEndpoint, setNamespace } = props
  const [serviceNameOptions, setServiceNameOptions] = useState([])
  const [endpointNameOptions, setEndpointNameOptions] = useState([])
  const [namespaceOptions, setNamespaceOptions] = useState([])
  const [serachServiceName, setSerachServiceName] = useState(null)
  const [serachEndpointName, setSerachEndpointName] = useState(null)
  const [serachNamespace, setSerachNamespace] = useState(null)
  const [prevSearchServiceName, setPrevSearchServiceName] = useState(null)
  const { startTime, endTime } = useSelector(selectSecondsTimeRange)

  //获取并设置服务名称选项
  const getServiceNameOptions = () => {
    const params = {
      startTime,
      endTime,
      namespace: serachNamespace || undefined,
    }
    getServiceListApi(params)
      .then((data) => {
        setServiceNameOptions(data.map((value) => ({ value, label: value })))
        //在改变namespace后过滤掉不包含在选中的namespace的服务名
        if (serachServiceName.length) {
          onChangeServiceName(serachServiceName.filter((item) => data.includes(item)))
        }
      })
      .catch((error) => console.error('获取数据失败:', error))
  }
  //获取并设置命名空间选项。
  const getNamespaceOptions = () => {
    const params = { startTime, endTime }
    getNamespacesApi(params)
      .then((data) => {
        setNamespaceOptions(
          (data?.namespaceList || []).map((namespace) => ({
            value: namespace,
            label: namespace,
          })),
        )
      })
      .catch((error) => console.error('获取数据失败:', error))
  }
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
  //处理命名空间选择更改。
  const onChangeNamespace = (event) => setSerachNamespace(event)
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
    if (startTime && endTime) {
      Promise.all([getServiceNameOptions(), getNamespaceOptions()])
    }
  }, [startTime, endTime, serachNamespace])

  useEffect(() => {
    if (serachServiceName) {
      removeEndpointNames()
    }
  }, [serachServiceName])

  useEffect(() => {
    setServiceName(serachServiceName)
    setEndpoint(serachEndpointName)
    setNamespace(serachNamespace)
  }, [serachServiceName, serachEndpointName, serachNamespace])

  return (
    <>
      <div className="p-2 my-2 flex flex-row w-full">
        <div className="flex flex-row items-center mr-5 text-sm min-w-[280px]">
          <span className="text-nowrap">命名空间：</span>
          <Select
            mode="multiple"
            allowClear
            id="namespace"
            className="w-full"
            placeholder="请选择"
            value={serachNamespace}
            onChange={onChangeNamespace}
            options={namespaceOptions}
            maxTagCount={2}
            maxTagPlaceholder={(omittedValues) => `+${omittedValues.length}...`}
          />
        </div>
        <div className="flex flex-row items-center mr-5 text-sm min-w-[280px]">
          <span className="text-nowrap">服务名：</span>
          <Select
            mode="multiple"
            allowClear
            className="w-full"
            id="serviceName"
            placeholder="请选择"
            value={serachServiceName}
            onChange={onChangeServiceName}
            options={serviceNameOptions}
            popupMatchSelectWidth={false}
            maxTagCount={2}
            maxTagPlaceholder={(omittedValues) => `+${omittedValues.length}...`}
          />
        </div>
        <div className="flex flex-row items-center mr-5 text-sm min-w-[280px]">
          <span className="text-nowrap">服务端点：</span>
          <Select
            mode="multiple"
            id="endpointName"
            placeholder="请选择"
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
