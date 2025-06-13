/**
 * Copyright 2025 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */
import { Popover, Button } from 'antd'
import { useState } from 'react'
import { FilterProps } from './type'
import FilterRenderer from './FilterRenderer'
import FilterTagItem from './FilterTagItem'
import { LuTrash } from 'react-icons/lu'
import { RiAddLargeFill } from 'react-icons/ri'
import _ from 'lodash'
import { useTranslation } from 'react-i18next'
const AddFilter = ({ filters, keys, onAddFilter, labelKeys }) => {
  const { t } = useTranslation('oss/alertEvents')
  const [selectedItem, setSelectedItem] = useState(null)
  const existingKeys = filters.map((f) => f.key)
  const filteredKeys = keys.filter((item) => !existingKeys.includes(item.key))
  return (
    <>
      {/* <Input.Search allowClear onSearch={onSearch} /> */}
      {selectedItem ? (
        <FilterRenderer item={selectedItem} addFilter={onAddFilter} filters={filters} />
      ) : (
        <>
          {filteredKeys.map((item) => (
            <div
              key={item.key}
              className="mx-2 my-1 cursor-pointer"
              onClick={() => setSelectedItem(item)}
            >
              {item.name}
            </div>
          ))}
          <div
            className="mx-2  cursor-pointer"
            onClick={() =>
              setSelectedItem({ key: 'label', name: t('alertDetail'), labelKeys: labelKeys })
            }
          >
            {t('expand')}
          </div>
        </>
      )}
    </>
  )
}

const Filter = ({ keys, labelKeys, filters, setFilters }: FilterProps) => {
  const { t } = useTranslation('oss/alertEvents')
  const [open, setOpen] = useState(false)
  const onClear = () => {
    setFilters([])
  }
  const onAddFilter = (filter) => {
    setFilters((prev) => [...prev, filter])
  }
  const onChangeFilter = (filter) => {
    setFilters((prev) =>
      prev.map((item) => {
        if (item.key === filter.key || item.key === filter.oldKey) item = _.cloneDeep(filter)
        return item
      }),
    )
  }
  const onDeleteFilter = (filter) => {
    setFilters((prev) => prev.filter((item) => item.key !== filter.key))
  }
  return (
    <div className="flex items-center mb-1 text-wrap flex-wrap text-xs">
      {filters?.map((item) => (
        <FilterTagItem
          key={item.key}
          filters={filters}
          item={item}
          onDeleteFilter={onDeleteFilter}
          onChangeFilter={onChangeFilter}
        />
      ))}

      <Popover
        open={open}
        trigger="click"
        placement="bottomLeft"
        content={
          <AddFilter
            keys={keys}
            onAddFilter={(value) => {
              onAddFilter(value)
              setOpen(false)
            }}
            filters={filters}
            labelKeys={labelKeys}
          />
        }
        destroyTooltipOnHide
        onOpenChange={setOpen}
      >
        <Button
          className="text-xs w-auto mb-1 mr-2"
          onClick={() => setOpen(!open)}
          color="primary"
          variant="outlined"
          size="small"
          icon={<RiAddLargeFill />}
        >
          {t('addFilter')}
        </Button>
      </Popover>
      <Button
        className="text-xs w-auto mb-1"
        onClick={onClear}
        color="primary"
        variant="outlined"
        size="small"
        icon={<LuTrash />}
      >
        {t('clearAll')}
      </Button>
    </div>
  )
}
export default Filter
