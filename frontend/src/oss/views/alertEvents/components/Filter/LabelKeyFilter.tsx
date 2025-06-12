/**
 * Copyright 2025 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */
import { Select, Segmented, Input, Button } from 'antd'
import { useState, useEffect } from 'react'
import { useSelector } from 'react-redux'
import { getAlertsFilterValuesApi } from 'src/core/api/alerts'
import { FilterRenderProps } from './type'

const LabelKeyFilter = ({ item, addFilter, filters }: FilterRenderProps) => {
  const [options, setOptions] = useState([])
  const [key, setKey] = useState()
  const [mode, setMode] = useState('selected')
  const [selected, setSelected] = useState([])
  const [matchExpr, setMatchExpr] = useState<string>()
  const [loading, setLoading] = useState(false)
  const { startTime, endTime } = useSelector((state) => state.timeRange)
  const modeOptions = [
    {
      label: '精准筛选',
      value: 'selected',
    },
    {
      label: '模糊匹配',
      value: 'matchExpr',
    },
  ]
  const getAlertsFilterValues = () => {
    setLoading(true)
    getAlertsFilterValuesApi({ searchKey: key, startTime, endTime })
      .then((res) => {
        setOptions(
          res.options?.map((option) => ({
            label: option.display,
            value: option.value,
          })),
        )
      })
      .finally(() => {
        setLoading(false)
      })
  }
  useEffect(() => {
    if (key) getAlertsFilterValues()
  }, [key])
  const getSelectedItems = () => {
    return options
      .filter((opt) => selected.includes(opt.value))
      .map((opt) => ({
        label: <div className="mx-1">{opt.label}</div>,
        value: opt.value,
      }))
  }
  useEffect(() => {
    const oldValue = filters.find((filterItem) => filterItem.key === item.key)
    if (oldValue) {
      setKey(item.key)
      if (item.matchExpr) {
        setMode('matchExpr')
        setMatchExpr(item.matchExpr)
      } else {
        setMode('selected')
        setSelected(item.selected)
      }
    }
  }, [filters])
  const existingKeys = filters.map((f) => f.key)
  const filteredKeys = item.labelKeys.filter((item) => !existingKeys.includes(item))
  return (
    <div>
      过滤字段
      <Select
        placeholder="请选择需要过滤的字段"
        options={filteredKeys.map((item) => ({
          label: item,
          value: item,
        }))}
        value={key}
        onChange={(key) => {
          setKey(key)
          setSelected([])
        }}
        className="min-w-[100px] m-1"
        showSearch
      />
      <div>
        筛选方式
        <Segmented options={modeOptions} value={mode} onChange={setMode} className="m-1" />
      </div>
      <div className="m-1">
        {mode === 'matchExpr' ? (
          <Input
            value={matchExpr}
            onChange={(e) => setMatchExpr(e.target.value)}
            placeholder="请输入"
            // onPressEnter={() =>
            //   addFilter({
            //     key: item.key,
            //     matchExpr: value,
            //     name: item.name,
            //   })
            // }
          />
        ) : (
          <>
            <Select
              allowClear
              className="min-w-[220px] max-w-[400px]"
              mode="multiple"
              placeholder="请选择"
              options={options}
              showSearch
              value={selected}
              onChange={setSelected}
            />
          </>
        )}
        <div className="text-right">
          <Button
            type="primary"
            block
            className="my-1"
            onClick={() => {
              if (key) {
                if (mode === 'matchExpr' && matchExpr?.length > 0) {
                  addFilter({
                    key: key,
                    matchExpr: matchExpr,
                    name: '告警详情',
                    isLabelKey: true,
                    labelKeys: item.labelKeys,
                    oldKey: item.key,
                  })
                } else if (mode === 'selected' && selected?.length > 0) {
                  addFilter({
                    key: key,
                    selected: selected,
                    name: '告警详情',
                    isLabelKey: true,
                    labelKeys: item.labelKeys,
                    selectedOptions: getSelectedItems(),
                    oldKey: item.key,
                  })
                }
              }
            }}
          >
            确定
          </Button>
        </div>
      </div>
    </div>
  )
}
export default LabelKeyFilter
