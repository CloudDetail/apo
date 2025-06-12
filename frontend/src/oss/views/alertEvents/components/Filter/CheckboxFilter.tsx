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

const CheckboxFilter = ({ filters, item, addFilter }: FilterRenderProps) => {
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
  }, [item.key])
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
        <>
          <Checkbox.Group
            value={value}
            options={options}
            defaultValue={['firing']}
            className="w-full flex flex-col m-2 items-center"
            onChange={setValue}
          ></Checkbox.Group>
          <Button
            onClick={() =>
              addFilter({
                key: item.key,
                selected: value,
                name: item.name,
                selectedOptions: getSelectedItems(),
              })
            }
          >
            确定
          </Button>
        </>
      )}
    </div>
  )
}
export default CheckboxFilter
