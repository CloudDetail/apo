import React, { useMemo, useState, useEffect, useRef } from 'react'
import BasicTable from 'src/core/components/Table/basicTable'
import { useSelector } from 'react-redux'
import { selectSecondsTimeRange } from 'src/core/store/reducers/timeRangeReducer'
import { getStep } from 'src/core/utils/step'
import { getServiceErrorInstancesApi } from 'src/core/api/serviceInfo'
import ErrorCell from '../../../../../../core/components/ErrorInstance/ErrorCell'
import { useDebounce } from 'react-use'
import { Button, Skeleton, Flex } from 'antd'
import { useServiceAnalyzeContext } from 'src/pro/contexts/ServiceAnalyzeContext'
import { useTranslation } from 'react-i18next'
const i18nKey = 'node'
export default function ErrorInstanceInfo(props) {
  const { t } = useTranslation('pro/serviceAnalyze')
  //showLogs 为true 服务于service info的tab，展示日志后两列，且支持点击节点跳转
  //showLogs false 服务于告警分析二级页面，不展示日志后两列，且不支持点击节点跳转
  const { selectedNodeInfo, service, endpoint } = useServiceAnalyzeContext()
  const { handlePanelStatus = null, showLogs = true, errorInstanceOpen } = props
  const [status, setStatus] = useState()
  const [data, setData] = useState()
  const [loading, setLoading] = useState(false)
  const { startTime, endTime } = useSelector(selectSecondsTimeRange)
  const [reRender, setReRender] = useState(0)
  const tableRef = useRef(null)
  const { dataGroupId } = useSelector((state) => state.dataGroupReducer)
  const { clusterIds } = useServiceAnalyzeContext()
  const column = [
    {
      title: t(`${i18nKey}.instanceName`),
      accessor: 'name',
      customWidth: 150,
    },
    {
      title: t(`${i18nKey}.currentNodeException`),
      accessor: 'propations',
      minWidth: 500,
      Cell: (props) => {
        const { value, column, updateSelectedValue, row } = props
        const update = (item) => {
          updateSelectedValue(column.id, item)
        }

        return (
          <ErrorCell data={value} instance={row.values.name} canClickTo={false} update={update} />
        )
      },
    },
    {
      title: t(`${i18nKey}.action`),
      accessor: 'action',
      dependsOn: 'propations',
      customWidth: 160,
      Cell: (props) => {
        const { dependsValue } = props
        //dependsValue里有traceid
        return (
          <Button
            type="link"
            href={
              dependsValue
                ? `/#/cause/error-report/${encodeURIComponent(dependsValue?.customAbbreviation?.traceId)}`
                : null
            }
            target="_blank"
            style={dependsValue ? {} : { color: 'gray' }}
          >
            {dependsValue ? t(`${i18nKey}.viewReport`) : t(`${i18nKey}.noReport`)}
          </Button>
        )
      },
    },
  ]
  const getData = () => {
    if (startTime && endTime) {
      setLoading(true)
      getServiceErrorInstancesApi({
        startTime: startTime,
        endTime: endTime,
        service: selectedNodeInfo.service,
        endpoint: selectedNodeInfo.endpoint,
        step: getStep(startTime, endTime),
        ignoreEmptyPropationsInstance: true,
        entryService: service,
        entryEndpoint: endpoint,
        clusterIds: clusterIds,
        groupId: dataGroupId,
      })
        .then((res) => {
          setStatus(res.status)
          handlePanelStatus?.(res.status)
          setData(res?.instances)
          setLoading(false)
        })
        .catch((error) => {
          setData([])
          handlePanelStatus?.('unknown')
          setLoading(false)
        })
    }
  }
  // useEffect(() => {
  //   getData()
  // }, [service, startTime, endTime])
  //防抖避免跳转使用旧时间

  useDebounce(
    () => {
      if (selectedNodeInfo.endpoint && dataGroupId !== null) {
        getData()
      }
    },
    300,
    [selectedNodeInfo, startTime, endTime, dataGroupId, clusterIds],
  )

  useEffect(() => {
    if (errorInstanceOpen) {
      setReRender(reRender + 1)
    }
  }, [errorInstanceOpen])

  const tableProps = useMemo(() => {
    return {
      columns: column,
      data: data,
      showBorder: false,
      loading: false,
    }
  }, [data, selectedNodeInfo])

  return (
    <div key={reRender} className="w-full overflow-auto">
      {/* <CAccordionHeader onClick={() => handleToggle('error')}>
        {status && <StatusInfo status={status} />}
        <span className="ml-2">{service}的错误实例</span>
      </CAccordionHeader> */}
      {selectedNodeInfo.endpoint ? (
        <Skeleton loading={loading}>
          {data && (
            <div ref={tableRef} className="text-xs w-full min-w-[800px]">
              <BasicTable {...tableProps} />
            </div>
          )}
        </Skeleton>
      ) : (
        <Flex>
          <p
            style={{
              color: '#9ca3af',
              fontSize: '12px',
            }}
          >
            {t(`${i18nKey}.selectNode`)}
          </p>
        </Flex>
      )}
    </div>
  )
}
