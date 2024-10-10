import { Descriptions, Dropdown, Input, List, Menu, Popover, Select, Space, theme } from 'antd'
import { Tooltip } from 'chart.js'
import React, { useEffect, useState } from 'react'
import { IoMdSearch } from 'react-icons/io'
import { MdOutlineManageSearch } from 'react-icons/md'
import { getRuleMetricsApi } from 'src/api/alerts'
import MonacoEditorWrapper from 'src/components/Editor/MonacoEditor'

export default function ALertConditionCom({ expr, setExpr }) {
  const [metricsList, setMetricsList] = useState([])
  const [metricsDetail, setMetricsDetail] = useState()
  const { useToken } = theme
  const { token } = useToken()
  const contentStyle = {
    backgroundColor: token.colorBgElevated,
    borderRadius: token.borderRadiusLG,
    boxShadow: token.boxShadowSecondary,
  }
  const menuStyle = {
    boxShadow: 'none',
  }
  useEffect(() => {
    if (metricsList.length <= 0) {
      getRuleMetricsApi().then((res) => {
        setMetricsList(res.alertMetricsData ?? [])
      })
    }
  }, [])
  const handlePopoverOpen = (item) => {
    setMetricsDetail(item)
  }
  const convertMetricsListToMenuItems = () => {
    return metricsList.map((item) => ({
      key: item.name,
      label: <div onMouseEnter={() => handlePopoverOpen(item)}>{item.name}</div>,
      onClick: () => setExpr(item.pql),
    }))
  }
  return (
    <>
      <div className=" flex border-1 border-solid rounded  border-[#424242] hover:border-[#3c89e8]  focus:border-[#3c89e8] ">
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
                          label: '表达式',
                          children: metricsDetail.pql,
                        },
                        {
                          key: 'unit',
                          label: '单位',
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
              <Space className=" cursor-pointer text-blue-400">
                快速指标 <MdOutlineManageSearch />
              </Space>
            </a>
          </Dropdown>
        </div>
        <div className="flex-1">
          <MonacoEditorWrapper defaultValue={expr} handleEditorChange={setExpr} />
        </div>
      </div>

      {/* <div className="flex items-center mt-2">
        当
        <Popover
          content={
            <div className="w-[500px]">
              <MonacoEditorWrapper defaultValue={expr} handleEditorChange={setExpr} />
            </div>
          }
        >
          <span className="text-blue-400 cursor-pointer">指标</span>
        </Popover>
        <Select
          value={symbol}
          className="w-auto mx-1"
          defaultValue={'>'}
          options={[
            { label: '>', value: '>' },
            { label: '<', value: '<' },
            { label: '=', value: '=' },
          ]}
        />
        <Input value={condition} placeholder="条件值" className="w-[80px] mx-1" />
        <Input value={unit} placeholder="单位" className="w-[80px] mx-1" /> 触发通知, 查看
        <Popover
          title="最终查询语句"
          content={
            <div className="w-[500px]">
              <MonacoEditorWrapper defaultValue={expr} handleEditorChange={setExpr} />
            </div>
          }
        >
          <span className="text-blue-400 cursor-pointer inline-flex items-center">
            {' '}
            最终查询语句 <IoMdSearch />
          </span>
        </Popover>
      </div> */}
    </>
  )
}
