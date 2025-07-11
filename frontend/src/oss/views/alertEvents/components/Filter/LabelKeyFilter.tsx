/**
 * Copyright 2025 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */
import { Select, Segmented, Input, Button } from 'antd'
import { useState, useEffect } from 'react'
import { useSelector } from 'react-redux'
import { getAlertsFilterValuesApi } from 'src/core/api/alerts'
import { FilterRenderProps } from './type'
import { useTranslation } from 'react-i18next'
import { useDebounce } from 'react-use'

const LabelKeyFilter = ({ item, addFilter, filters }: FilterRenderProps) => {
  const { t } = useTranslation('oss/alertEvents')
  const [options, setOptions] = useState([])
  const [key, setKey] = useState()
  const [mode, setMode] = useState('selected')
  const [selected, setSelected] = useState([])
  const [matchExpr, setMatchExpr] = useState<string>()
  const [loading, setLoading] = useState(false)
  const { startTime, endTime } = useSelector((state) => state.timeRange)
  const { dataGroupId } = useSelector((state) => state.dataGroupReducer)
  const modeOptions = [
    {
      label: t('selected'),
      value: 'selected',
    },
    {
      label: t('matchExpr'),
      value: 'matchExpr',
    },
  ]
  const getAlertsFilterValues = () => {
    setLoading(true)
    getAlertsFilterValuesApi({ searchKey: key, startTime, endTime, groupId: dataGroupId })
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
  useDebounce(
    () => {
      if (key && startTime && endTime && dataGroupId !== null) {
        getAlertsFilterValues()
      }
    },
    300,
    [key, startTime, endTime, dataGroupId],
  )

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
  useEffect(() => {
    if (item.wildcard) setKey(item.key)
  }, [item])
  const existingKeys = filters.map((f) => f.key)
  const filteredKeys = item.labelKeys?.filter((item) => !existingKeys.includes(item))
  return (
    <div>
      {filteredKeys?.length > 0 ? (
        <>
          {t('filterKey')}
          <Select
            placeholder={t('selectPlaceholder')}
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
        </>
      ) : (
        item.name
      )}
      <div>
        {t('filterMode')}
        <Segmented options={modeOptions} value={mode} onChange={setMode} className="m-1" />
      </div>
      <div className="m-1">
        {mode === 'matchExpr' ? (
          <>
            <Input value={matchExpr} onChange={(e) => setMatchExpr(e.target.value)} />
            <div className=" py-1 text-xs text-[var(--ant-color-text-secondary)]">
              {t('exprHint')}
            </div>
          </>
        ) : (
          <>
            <Select
              allowClear
              className="min-w-[220px] max-w-[400px] w-full"
              mode="multiple"
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
                    name: item.wildcard ? item.name : t('expand'),
                    isLabelKey: !item.wildcard,
                    wildcard: item.wildcard,
                    labelKeys: item.labelKeys,
                    oldKey: item.key,
                  })
                } else if (mode === 'selected' && selected?.length > 0) {
                  addFilter({
                    key: key,
                    selected: selected,
                    name: item.wildcard ? item.name : t('expand'),
                    isLabelKey: !item.wildcard,
                    labelKeys: item.labelKeys,
                    selectedOptions: getSelectedItems(),
                    oldKey: item.key,
                    wildcard: item.wildcard,
                  })
                }
              }
            }}
          >
            {t('confirm')}
          </Button>
        </div>
      </div>
    </div>
  )
}
export default LabelKeyFilter
