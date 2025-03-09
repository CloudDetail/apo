import { CAccordionBody } from '@coreui/react'
import React, { useMemo, useState, useEffect, useRef } from 'react'
import BasicTable from 'src/core/components/Table/basicTable'
import DelayLineChart from 'components/Chart/DelayLineChart'
import { usePropsContext } from 'src/core/contexts/PropsContext'
import { useSelector } from 'react-redux'
import { selectSecondsTimeRange } from 'src/core/store/reducers/timeRangeReducer'
import { getStep } from 'src/core/utils/step'
import { getServiceErrorInstancesApi } from 'src/core/api/serviceInfo'
import ErrorCell from '../../../../../../core/components/ErrorInstance/ErrorCell'
import { ErrorChain } from '../../../../../../core/components/ErrorInstance/ErrorChain'
import { useDebounce } from 'react-use'
import Timeline from '../../../../../../core/components/TimeLine'
import { Button, Skeleton, Flex } from 'antd'
import { useAlertAnalyzeContext } from 'src/pro/contexts/AlertAnalyzeContext'
import { useTranslation } from 'react-i18next'
const i18nKey = 'node'
export default function ErrorInstanceInfo(props) {
  const { t } = useTranslation('pro/serviceAnalyze')
  //showLogs 为true 服务于service info的tab，展示日志后两列，且支持点击节点跳转
  //showLogs false 服务于告警分析二级页面，不展示日志后两列，且不支持点击节点跳转
  const { selectedNodeInfo, selectedEndpointInfo } = useAlertAnalyzeContext()
  const { handlePanelStatus = null, service, endpoint, showLogs = true, errorInstanceOpen } = props
  const [status, setStatus] = useState()
  const [data, setData] = useState()
  const [loading, setLoading] = useState(false)
  const { startTime, endTime } = useSelector(selectSecondsTimeRange)
  const [reRender, setReRender] = useState(0)
  const tableRef = useRef(null)

  const column = [
    {
      title: t(`${i18nKey}.instanceName`),
      accessor: 'name',
      customWidth: 150,
    },
    {
      title: t(`${i18nKey}.currentNodeException`),
      accessor: 'propations',
      Cell: (props) => {
        const { value, updateSelectedValue, column } = props
        // updateSelectedValue(column.id, value[0]);
        const update = (item) => {
          updateSelectedValue(column.id, item)
        }

        return <ErrorCell data={value} update={update} />
      },
    },
    {
      title: t(`${i18nKey}.errorPropagationChain`),
      accessor: 'chain',
      dependsOn: 'propations', // 该列依赖于第一列的值
      Cell: (props) => {
        const { value, dependsValue, originalRow } = props
        return (
          <ErrorChain
            data={dependsValue?.customAbbreviation}
            instance={originalRow.name}
            chartId={originalRow.name}
            endpoint={endpoint}
            canClickTo={showLogs}
          />
        )
        // return  nodes?.length > 0 ? <Topology canZoom={false} data={{nodes, edges}}  /> : <Empty />
      },
    },
    {
      title: t(`${i18nKey}.logErrorCount`),
      accessor: 'logs',
      Cell: (props) => {
        const { value } = props
        return (
          <div className="h-52 w-80">
            <DelayLineChart data={value} timeRange={{ startTime, endTime }} type="logs" />
          </div>
        )
      },
    },
    {
      title: t(`${i18nKey}.logInfo`),
      accessor: 'detail',
      customWidth: 320,
      Cell: (props) => {
        const { value, row } = props
        return (
          <Timeline
            instance={row.original.name}
            nodeName={row.original.nodeName}
            pid={row.original.pid}
            containerId={row.original.containerId}
            type="errorLogs"
            startTime={startTime}
            endTime={endTime}
            service={service}
            endpoint={endpoint}
          />
        )
      },
    },
    {
      title: t(`${i18nKey}.action`),
      accessor: 'action',
      dependsOn: 'propations',
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
        service: service,
        endpoint: endpoint,
        step: getStep(startTime, endTime),
        ignoreEmptyPropationsInstance: true,
        entryService: selectedEndpointInfo.service,
        entryEndpoint: selectedEndpointInfo.endpoint,
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
      if (selectedNodeInfo.endpoint) {
        getData()
      }
    },
    300,
    [service, startTime, endTime, endpoint],
  )

  useEffect(() => {
    if (errorInstanceOpen) {
      setReRender(reRender + 1)
    }
  }, [errorInstanceOpen])

  const tableProps = useMemo(() => {
    return {
      columns: showLogs ? column : [...column.slice(0, 3), ...column.slice(-1)],
      data: data,
      showBorder: false,
      loading: false,
    }
  }, [data, service])

  return (
    <div key={reRender}>
      {/* <CAccordionHeader onClick={() => handleToggle('error')}>
        {status && <StatusInfo status={status} />}
        <span className="ml-2">{service}的错误实例</span>
      </CAccordionHeader> */}
      {selectedNodeInfo.endpoint ? (
        <Skeleton loading={loading}>
          {data && (
            <div ref={tableRef} className="text-xs">
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
