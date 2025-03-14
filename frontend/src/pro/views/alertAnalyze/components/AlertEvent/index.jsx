import { Pagination } from 'antd'
import React, { useEffect, useMemo, useState } from 'react'
import ReactJson from 'react-json-view'
import { useSelector } from 'react-redux'
import BasicTable from 'src/core/components/Table/basicTable'
import { useAlertAnalyzeContext } from 'src/pro/contexts/AlertAnalyzeContext'
import { convertUTCToBeijing } from 'src/core/utils/time'
import ServiceEndpointTag from 'src/pro/components/ServiceEndpointTag'
import { useTranslation } from 'react-i18next'
function isJSONString(str) {
  try {
    JSON.parse(str)
    return true
  } catch (e) {
    return false
  }
}
const i18nKey = 'alertEvent'
const AlertEventTable = ({ scrolllHeight }) => {
  const { t } = useTranslation('pro/alertAnalyze')

  const { alertEvents, loading, isAnalyze, pagination, updatePagination } = useAlertAnalyzeContext()
  const { groupLabel } = useSelector((state) => state.groupLabelReducer)
  const columns = [
    {
      title: t(`${i18nKey}.alertType`),
      accessor: 'group',
      customWidth: 150,
      Cell: ({ value }) => {
        return groupLabel[value]
      },
    },
    {
      title: t(`${i18nKey}.alertName`),
      accessor: 'name',
      customWidth: 150,
    },
    {
      title: t(`${i18nKey}.lastAlertTime`),
      accessor: 'receivedTime',
      customWidth: 180,
      Cell: ({ value }) => {
        return convertUTCToBeijing(value)
      },
    },
    {
      title: t(`${i18nKey}.alertSource`),
      accessor: 'source',
      customWidth: 150,
    },
    {
      title: t(`${i18nKey}.alertDetail`),
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
      title: t(`${i18nKey}.relatedApps`),
      accessor: 'impactEndpoints',
      justifyContent: 'left',
      Cell: ({ value }) => {
        return (
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
              <span className="text-gray-400">{t(`${i18nKey}.noRelatedApps`)}</span>
            )}
          </>
        )
      },
    },
    {
      title: t(`${i18nKey}.affectedEntries`),
      accessor: 'impactEntrys',
      justifyContent: 'left',
      Cell: ({ value }) => {
        return (
          <>
            {value?.length > 0 ? (
              value?.map((item) => (
                <div className="flex-shrink-0 ">
                  <ServiceEndpointTag service={item.serviceName} endpoint={item.endpoint} />
                </div>
              ))
            ) : (
              <span className="text-gray-400">{t(`${i18nKey}.noAffectedEntries`)}</span>
            )}
          </>
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
        {/* <div className="text-base px-4 pt-2 font-bold">{t(`${i18nKey}.alertEvents`)}</div> */}
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
