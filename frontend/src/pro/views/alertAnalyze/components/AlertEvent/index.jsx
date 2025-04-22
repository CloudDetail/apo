import { Button, Pagination, Tabs, Tag } from 'antd'
import React, { useMemo, useState } from 'react'
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
      customWidth: 120,
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
    // {
    //   title: t(`${i18nKey}.alertSource`),
    //   accessor: 'source',
    //   customWidth: 100,
    // },
    {
      title: t(`${i18nKey}.alertDetail`),
      accessor: 'rawTags',
      justifyContent: 'left',
      customWidth: '40%',

      Cell: ({ value, row }) => {
        const [visible, setVisible] = useState(false)

        const { detail } = row.original
        return (
          <div className="overflow-hidden">
            {Object.entries(value || {}).map(([key, tagValue]) => (
              <Tag className="text-pretty mb-1 break-all">
                {key} = {tagValue}
              </Tag>
            ))}

            {isJSONString(detail) && (
              <Button
                color="primary"
                variant="text"
                size="small"
                onClick={() => setVisible(!visible)}
              >
                {visible ? t('collapse') : t('expand')}
              </Button>
            )}

            {visible && (
              <ReactJson
                src={JSON.parse(detail || '')}
                theme="brewer"
                collapsed={false}
                displayDataTypes={false}
                style={{ width: '100%' }}
                enableClipboard={false}
              />
            )}
          </div>
        )
      },
    },
    {
      title: t(`${i18nKey}.relatedApps`),
      accessor: 'impactEndpoints',
      justifyContent: 'left',
      Cell: ({ value }) => {
        return (
          <div className="w-full">
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
    }
  }, [alertEvents])
  return (
    <>
      <div className="flex flex-col h-full text-xs">
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
