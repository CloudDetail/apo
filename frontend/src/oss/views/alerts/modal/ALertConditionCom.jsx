import { Descriptions, Dropdown, Space, theme } from 'antd'
import React, { useEffect, useState, useMemo, useCallback } from 'react'
import { MdOutlineManageSearch } from 'react-icons/md'
import { getRuleMetricsApi } from 'core/api/alerts'
import MonacoEditorWrapper from 'src/core/components/Editor/MonacoEditor'
import { useTranslation } from 'react-i18next'

export default function ALertConditionCom({ expr, setExpr }) {
  const { t } = useTranslation('oss/alert')
  const [metricsList, setMetricsList] = useState([])
  const [metricsDetail, setMetricsDetail] = useState()
  const { useToken } = theme
  const { token } = useToken()

  const contentStyle = useMemo(
    () => ({
      backgroundColor: token.colorBgElevated,
      borderRadius: token.borderRadiusLG,
      boxShadow: token.boxShadowSecondary,
    }),
    [token],
  )

  const menuStyle = useMemo(
    () => ({
      boxShadow: 'none',
    }),
    [],
  )

  useEffect(() => {
    if (metricsList.length <= 0) {
      getRuleMetricsApi().then((res) => {
        setMetricsList(res.alertMetricsData ?? [])
      })
    }
  }, [metricsList.length])

  const handlePopoverOpen = useCallback((item) => {
    setMetricsDetail(item)
  }, [])

  const convertMetricsListToMenuItems = useCallback(() => {
    return metricsList.map((item, index) => ({
      key: `${item.name}-${index}`, // 确保 key 是唯一的
      label: <div onMouseEnter={() => handlePopoverOpen(item)}>{item.name}</div>,
      onClick: () => setExpr(item.pql),
    }))
  }, [metricsList, handlePopoverOpen, setExpr])

  return (
    <>
      <div className="flex border-1 border-solid rounded border-[#424242] hover:border-[#3c89e8] focus:border-[#3c89e8]">
        <div className="flex-grow-0 flex-shrink-0 flex items-center px-2">
          <Dropdown
            menu={{
              items: convertMetricsListToMenuItems(),
            }}
            dropdownRender={(menu) => (
              <div className="flex w-full" style={contentStyle}>
                {React.cloneElement(menu, {
                  style: menuStyle,
                })}
                {metricsDetail && (
                  <div className="w-[300px] overflow-hidden p-2">
                    <Descriptions
                      column={1}
                      title={metricsDetail.name}
                      items={[
                        {
                          key: 'pql',
                          label: t('alertConditionCom.expression'),
                          children: metricsDetail.pql,
                        },
                        {
                          key: 'unit',
                          label: t('alertConditionCom.unit'),
                          children: metricsDetail.unit,
                        },
                      ]}
                    />
                  </div>
                )}
              </div>
            )}
          >
            <a onClick={(e) => e.preventDefault()}>
              <Space className="cursor-pointer text-blue-400">
                {t('alertConditionCom.quickMetrics')} <MdOutlineManageSearch />
              </Space>
            </a>
          </Dropdown>
        </div>
        <div className="flex-1">
          <MonacoEditorWrapper defaultValue={expr} handleEditorChange={setExpr} />
        </div>
      </div>
    </>
  )
}
