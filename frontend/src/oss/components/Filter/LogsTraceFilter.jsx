/**
 * Copyright 2024 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import React, { useState } from 'react'
import { useSearchParams } from 'react-router-dom'
import DateTimeRangePickerCom from 'src/core/components/DateTime/DateTimeRangePickerCom'
import { Checkbox, Input, InputNumber, Segmented, Tooltip } from 'antd'
import TraceErrorType from 'src/oss/views/trace/component/TraceErrorType'
import { useTranslation } from 'react-i18next'
import DataSourceFilter from 'src/core/components/Filter/DataSourceFilter'
import { useLogsTraceFilterContext } from 'src/oss/contexts/LogsTraceFilterContext'
import { swTraceIDToTraceID } from 'src/core/utils/trace'

const LogsTraceFilter = React.memo(({ type }) => {
  const { t } = useTranslation(['common', 'oss/trace'])
  const [searchParams, setSearchParams] = useSearchParams()
  const [traceType, setTraceType] = useState('TraceID')
  const [inputSWTraceId, setInputSWTraceId] = useState('')
  const [convertTraceId, setConvertSWTraceId] = useState('')
  const {
    traceId,
    startTime,
    endTime,
    minDuration,
    maxDuration,
    endpoint,
    faultTypeList,
    setClusterIds,
    setInstance,
    setTraceId,
    setNamespaces,
    setIsFilterDone,
    setMinDuration,
    setMaxDuration,
    setEndpoint,
    setFaultTypeList,
    setServices,
  } = useLogsTraceFilterContext((ctx) => ctx)
  const options = [
    { label: <TraceErrorType type="slow" />, value: 'slow' },
    { label: <TraceErrorType type="error" />, value: 'error' },
    { label: <TraceErrorType type="normal" />, value: 'normal' },
    { label: <TraceErrorType type="slowAndError" />, value: 'slowAndError' },
  ]
  const params = searchParams

  // 转换函数
  const stringToArray = (value) => {
    if (!value) return null
    return value.split(',').filter(Boolean)
  }
  const initialParamsRef = React.useRef({
    clusterIds: stringToArray(params.get('clusterIds')),
    namespaces: stringToArray(params.get('namespaces')),
    instance: stringToArray(params.get('instance'))?.[0],
    traceId: params.get('traceId') || '',
  })

  const onChangeSWTraceId = (event) => {
    setInputSWTraceId(event.target.value)
    const convertTraceId = swTraceIDToTraceID(event.target.value)
    console.log(convertTraceId)
    setConvertSWTraceId(convertTraceId)
    setTraceId(convertTraceId)
  }

  const onChangeMinDuration = (value) => {
    setMinDuration(value)
  }
  const onChangeMaxDuration = (value) => {
    setMaxDuration(value)
  }
  return (
    <>
      <div className="flex flex-row mb-2 justify-between">
        <div className="flex flex-row  flex-wrap">
          <DataSourceFilter
            category={type === 'logs' ? 'log' : 'apm'}
            setCluster={(value) => setClusterIds(value)}
            setNamespace={(value) => setNamespaces(value)}
            setInstance={(value) => setInstance(value)}
            setServiceName={(value) => setServices(value)}
            setPod={(value) => setInstance(value[0])}
            initCluster={initialParamsRef?.current.clusterIds}
            initNamespace={initialParamsRef?.current.namespaces}
            startTime={startTime}
            endTime={endTime}
            setIsFilterDone={setIsFilterDone}
            extra="instance"
          />

          {type === 'logs' && (
            <div className="flex flex-row items-center  min-w-[200px]">
              <span className="text-nowrap ">{t('oss/trace:trace.traceId')}：</span>
              <Input
                placeholder={t('logsTraceFilter.search')}
                value={traceId}
                onChange={(e) => setTraceId(e.target.value)}
              />
            </div>
          )}
        </div>
        <div className="flex-grow-0 flex-shrink-0 flex">
          <DateTimeRangePickerCom type={type} />
        </div>
      </div>
      {type === 'trace' && (
        <>
          <div className="text-xs flex flex-row  flex-wrap w-full mb-2 items-center">
            <div className="flex flex-row items-center  ">
              <span className="text-nowrap">{t('logsTraceFilter.durationLabel')}：</span>
              <div className="flex-1 flex flex-row items-center mr-2">
                <div className="pr-2">
                  <InputNumber
                    addonBefore="MIN"
                    addonAfter="ms"
                    min={0}
                    value={minDuration}
                    onChange={onChangeMinDuration}
                    className=" w-[150px]"
                  />
                </div>
                {t('logsTraceFilter.toText')}
                <div className="pl-2">
                  <InputNumber
                    addonBefore="MAX"
                    addonAfter="ms"
                    min={0}
                    value={maxDuration}
                    onChange={onChangeMaxDuration}
                    className=" w-[150px]"
                  />
                </div>
              </div>
            </div>
            {type === 'trace' && (
              <div className="flex flex-row items-center  w-[150px]">
                <span className="text-nowrap ">{t('logsTraceFilter.endpointLabel')}：</span>
                <Input
                  placeholder={t('logsTraceFilter.search')}
                  value={endpoint}
                  onChange={(e) => setEndpoint(e.target.value)}
                />
              </div>
            )}
          </div>
          <div className="flex">
            <div className="flex flex-row items-center mr-2 w-[300px] text-sm">
              {type === 'trace' ? (
                <Segmented options={['TraceID', 'SWTraceId']} onChange={setTraceType} />
              ) : (
                <span className="text-nowrap text-sm">TraceId：</span>
              )}
              ：
              {traceType === 'TraceID' ? (
                <Input
                  placeholder={t('logsTraceFilter.search')}
                  value={traceId}
                  onChange={(e) => setTraceId(e.target.value)}
                />
              ) : (
                <Tooltip
                  title={
                    convertTraceId
                      ? t('logsTraceFilter.autoConvert') + convertTraceId
                      : t('logsTraceFilter.enterSWTraceId')
                  }
                >
                  <Input
                    placeholder={t('logsTraceFilter.search')}
                    value={inputSWTraceId}
                    onChange={onChangeSWTraceId}
                  />
                </Tooltip>
              )}
            </div>
            <div className="flex flex-row items-center text-xs  mr-2">
              <span className="text-nowrap">{t('logsTraceFilter.status')}</span>
              <Checkbox.Group
                onChange={(value) => setFaultTypeList(value)}
                options={options}
                value={faultTypeList}
              ></Checkbox.Group>
            </div>
          </div>
        </>
      )}
    </>
  )
})
export default LogsTraceFilter
