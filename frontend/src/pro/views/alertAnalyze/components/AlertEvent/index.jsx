import { Tabs } from 'antd'
import React, { useMemo } from 'react'
import ReactJson from 'react-json-view'
import { useSelector } from 'react-redux'
import BasicTable from 'src/core/components/Table/basicTable'
import { useAlertAnalyzeContext } from 'src/pro/contexts/AlertAnalyzeContext'
import { convertUTCToLocal } from 'src/core/utils/time'
import ServiceEndpointTag from 'src/pro/components/ServiceEndpointTag'
import { useTranslation } from 'react-i18next'
import styles from './index.module.scss'
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
      customWidth: 100,
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
      customWidth: 100,
      Cell: ({ value }) => {
        const result = convertUTCToLocal(value)
        return (
          <div>
            <div>{result.split(' ')[0]}</div>
            <div>{result.split(' ')[1]}</div>
          </div>
        )
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
          <div className="w-full">
            {value?.length > 0 ? (
              value?.map((item) => (
                <Tabs
                  className={styles.impactEntrys}
                  defaultActiveKey="1"
                  tabPosition="top"
                  activeKey="0"
                  tabBarStyle={{ whiteSpace: 'nowrap', flexShrink: 0 }}
                  items={value?.map((item) => {
                    const id = item.serviceName + item.endpoint
                    return {
                      label: (
                        <div className=" whitespace-nowrap flex items-center">
                          <ServiceEndpointTag service={item.serviceName} endpoint={item.endpoint} />
                        </div>
                      ),
                      key: id,
                      children: null,
                    }
                  })}
                />
              ))
            ) : (
              <span className="text-gray-400">{t(`${i18nKey}.noRelatedApps`)}</span>
            )}
          </div>
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
              <Tabs
                className={styles.impactEntrys}
                defaultActiveKey="1"
                tabPosition="top"
                activeKey="0"
                tabBarStyle={{ whiteSpace: 'nowrap', flexShrink: 0 }}
                items={value?.map((item) => {
                  const id = item.serviceName + item.endpoint
                  return {
                    label: (
                      <div className=" whitespace-nowrap flex items-center">
                        <ServiceEndpointTag service={item.serviceName} endpoint={item.endpoint} />
                      </div>
                    ),
                    key: id,
                    children: null,
                  }
                })}
              />
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
      tdPadding: '2px',
      onChange: changePagination,
      pagination: {
        pageSize: pagination.pageSize,
        pageIndex: pagination.pageIndex,
        total: pagination.total,
      },
    }
  }, [alertEvents])
  return (
    <>
      <div className="flex flex-col h-full text-xs">
        <BasicTable {...tableProps} />
      </div>
    </>
  )
}
export default AlertEventTable
