/**
 * Copyright 2025 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */
import { Checkbox, Button, Input } from 'antd'
import { useState, useEffect } from 'react'
import { useSelector } from 'react-redux'
import { getAlertsFilterValuesApi } from 'src/core/api/alerts'
import LoadingSpinner from 'src/core/components/Spinner'
import ValueTag from './ValueTag'
import { FilterRenderProps } from './type'
import { useTranslation } from 'react-i18next'
import Empty from 'src/core/components/Empty/Empty'
import { LuSearch } from 'react-icons/lu'
import { useDebounce } from 'react-use'

const CheckboxFilter = ({ filters, item, addFilter }: FilterRenderProps) => {
  const { t } = useTranslation('oss/alertEvents')
  const [options, setOptions] = useState([])
  const [originalOptions, setOriginalOptions] = useState([])
  const [value, setValue] = useState([])
  const [loading, setLoading] = useState(false)
  const { startTime, endTime } = useSelector((state) => state.timeRange)
  const { dataGroupId } = useSelector((state) => state.dataGroupReducer)
  const getAlertsFilterValues = () => {
    setLoading(true)
    getAlertsFilterValuesApi({ searchKey: item.key, startTime, endTime, groupId: dataGroupId })
      .then((res) => {
        const options = res.options?.map((option) => ({
          label: <ValueTag {...option} itemKey={item.key} key={option.value} />,
          // label: option.display,
          value: option.value,
          display: option.display,
        }))
        setOptions(options)
        setOriginalOptions(options)
      })
      .finally(() => {
        setLoading(false)
      })
  }
  const getSelectedItems = () => {
    return options
      .filter((opt) => value.includes(opt.value))
      .map((opt) => ({
        label: opt.label,
        value: opt.value,
      }))
  }
  useDebounce(
    () => {
      if (startTime && endTime && dataGroupId !== null) {
        getAlertsFilterValues()
      }
    },
    300,
    [item.key, startTime, endTime, dataGroupId],
  )
  useEffect(() => {
    const oldValue = filters.find((filterItem) => filterItem.key === item.key)
    if (oldValue) setValue(oldValue.selected)
  }, [filters])

  function onSearch(value) {
    setOptions(originalOptions.filter((item) => item.display.includes(value)))
  }
  return (
    <div className=" relative">
      {loading ? (
        <LoadingSpinner loading={loading} />
      ) : !originalOptions || originalOptions?.length === 0 ? (
        <Empty />
      ) : (
        <div className=" max-h-[500px] overflow-y-auto overflow-x-hidden flex flex-col justify-between">
          <Input onChange={(e) => onSearch(e.target.value)} addonAfter={<LuSearch />} />
          <div className="flex-1 h-0 overflow-y-auto overflow-x-hidden">
            <Checkbox.Group
              value={value}
              options={options}
              className="w-full flex flex-col m-2 flex-1 "
              onChange={setValue}
            ></Checkbox.Group>
          </div>

          <Button
            type="primary"
            className="shrink-0 grow-0 "
            onClick={() =>
              addFilter({
                key: item.key,
                selected: value,
                name: item.name,
                selectedOptions: getSelectedItems(),
              })
            }
          >
            {t('confirm')}
          </Button>
        </div>
      )}
    </div>
  )
}
export default CheckboxFilter
