import { Pagination } from 'antd'
import React, { useEffect, useMemo, useState } from 'react'
import ReactJson from 'react-json-view'
import { useSelector } from 'react-redux'
import BasicTable from 'src/core/components/Table/basicTable'
import { useAlertAnalyzeContext } from 'src/pro/contexts/AlertAnalyzeContext'
import { convertUTCToBeijing } from 'src/core/utils/time'
import ServiceEndpointTag from 'src/pro/components/ServiceEndpointTag'
function isJSONString(str) {
  try {
    JSON.parse(str)
    return true
  } catch (e) {
    return false
  }
}
const AlertEventTable = ({ scrolllHeight }) => {
  const { alertEvents, loading, isAnalyze, pagination, updatePagination } = useAlertAnalyzeContext()
  const { groupLabel } = useSelector((state) => state.groupLabelReducer)
  const columns = [
    {
      title: '告警类型',
      accessor: 'group',
      customWidth: 150,
      Cell: ({ value }) => {
        return groupLabel[value]
      },
    },
    {
      title: '告警事件名',
      accessor: 'name',
      customWidth: 150,
    },
    {
      title: '最近告警时间',
      accessor: 'receivedTime',
      customWidth: 180,
      Cell: ({ value }) => {
        return convertUTCToBeijing(value)
      },
    },
    {
      title: '告警来源',
      accessor: 'source',
      customWidth: 150,
    },
    {
      title: '告警详情',
      accessor: 'detail',
      justifyContent: 'left',
      Cell: ({ value }) =>
        isJSONString(value) ? (
          <ReactJson
            src={JSON.parse(value || '')}
            theme="brewer"
            collapsed={true}
            displayDataTypes={false}
            style={{ width: '100%' }}
            enableClipboard={false}
          />
        ) : (
          value
        ),
    },
    {
      title: '告警相关应用',
      accessor: 'impactEndpoints',
      justifyContent: 'left',
      Cell: ({ value }) => {
        return isAnalyze ? (
          <>
            {value?.length > 0 ? (
              <div className="flex flex-wrap">
                {value?.map((item) => (
                  <div className="flex-shrink-0 w-full">
                    <ServiceEndpointTag service={item.serviceName} endpoint={item.endpoint} />
                  </div>
                ))}
              </div>
            ) : (
              <span className="text-gray-400">未分析出告警相关应用</span>
            )}
          </>
        ) : (
          <span className="text-gray-400">点击一键分析查看</span>
        )
      },
    },
    {
      title: '受影响业务入口',
      accessor: 'impactEntrys',
      justifyContent: 'left',
      Cell: ({ value }) => {
        return isAnalyze ? (
          <>
            {value?.length > 0 ? (
              value?.map((item) => (
                <div className="flex-shrink-0 ">
                  <ServiceEndpointTag service={item.serviceName} endpoint={item.endpoint} />
                </div>
              ))
            ) : (
              <span className="text-gray-400">未分析出受影响业务入口</span>
            )}
          </>
        ) : (
          <span className="text-gray-400">点击一键分析查看</span>
        )
      },
    },
  ]
  const changePagination = (page, pageSize) => {
    updatePagination({
      pageSize: pageSize,
      pageIndex: page,
      total: pagination.total,
    })
  }
  const tableProps = useMemo(() => {
    return {
      columns: columns,
      data: alertEvents,
      showBorder: false,
      loading: false,
    }
  }, [alertEvents])
  return (
    <>
      <div className="flex flex-col h-full">
        <div className="text-base px-4 pt-2 font-bold">告警事件</div>
        <BasicTable {...tableProps} />
        <Pagination
          defaultCurrent={1}
          total={pagination.total}
          current={pagination.pageIndex}
          pageSize={pagination.pageSize}
          className="flex-shrink-0 flex-grow-0 p-2"
          align="end"
          onChange={changePagination}
        />
      </div>
    </>
  )
}
export default AlertEventTable
