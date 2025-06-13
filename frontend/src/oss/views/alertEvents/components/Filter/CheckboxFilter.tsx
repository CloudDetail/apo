/**
 * Copyright 2025 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */
import { Empty, Checkbox, Button } from 'antd'
import { useState, useEffect } from 'react'
import { useSelector } from 'react-redux'
import { getAlertsFilterValuesApi } from 'src/core/api/alerts'
import LoadingSpinner from 'src/core/components/Spinner'
import ValueTag from './ValueTag'
import { FilterRenderProps } from './type'
import { useTranslation } from 'react-i18next'

const CheckboxFilter = ({ filters, item, addFilter }: FilterRenderProps) => {
  const { t } = useTranslation('oss/alertEvents')
  const [options, setOptions] = useState([])
  const [value, setValue] = useState([])
  const [loading, setLoading] = useState(false)
  const { startTime, endTime } = useSelector((state) => state.timeRange)
  const getAlertsFilterValues = () => {
    setLoading(true)
    getAlertsFilterValuesApi({ searchKey: item.key, startTime, endTime })
      .then((res) => {
        setOptions(
          res.options?.map((option) => ({
            label: <ValueTag {...option} itemKey={item.key} key={option.value} />,
            // label: option.display,
            value: option.value,
          })),
        )
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
  useEffect(() => {
    getAlertsFilterValues()
  }, [item.key, startTime, endTime])
  useEffect(() => {
    const oldValue = filters.find((filterItem) => filterItem.key === item.key)
    if (oldValue) setValue(oldValue.selected)
  }, [filters])
  return (
    <div className=" relative">
      {loading ? (
        <LoadingSpinner loading={loading} />
      ) : !options || options?.length === 0 ? (
        <Empty />
      ) : (
        <div className=" max-h-[500px] overflow-y-auto overflow-x-hidden flex flex-col justify-between">
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
