import Popover from 'antd/es/popover'
import FilterRenderer from './FilterRenderer'
import { Tag, Button } from 'antd'
import { useState } from 'react'
import { IoClose } from 'react-icons/io5'
import ValueTag from './ValueTag'
import { FilterTagItemProps } from './type'

const FilterTagItem = ({ filters, onDeleteFilter, onChangeFilter, item }: FilterTagItemProps) => {
  const [open, setOpen] = useState(false)
  return (
    <div className="border rounded border-[var(--ant-color-border-secondary)] px-1 flex items-center mr-2 mb-1 p-1  overflow-hidden flex-shrink-0 ">
      <div className="text-[var(--ant-color-text-secondary)] px-1 shrink-0">{item.name}</div>
      <Popover
        trigger="click"
        placement="bottomLeft"
        open={open}
        content={
          <FilterRenderer
            item={item}
            addFilter={(value) => {
              onChangeFilter(value)
              setOpen(false)
            }}
            filters={filters}
          />
        }
        onOpenChange={setOpen}
        className="cursor-pointer "
        destroyTooltipOnHide
      >
        <div className="flex items-center flex-1 overflow-hidden">
          {item.isLabelKey && (
            <Tag color="blue" bordered={false} className=" text-[var(--ant-color-text)]">
              {item.key}
            </Tag>
          )}

          <div className="flex items-center gap-1 max-w-[300px]  whitespace-nowrap overflow-hidden text-ellipsis">
            {item.selectedOptions
              // ?.slice(0, 3)
              ?.map((option) => option.label)}

            {/* {item.selectedOptions && item.selectedOptions.length > 2 && (
              <span className="text-gray-400">...</span>
            )} */}
          </div>

          {item.matchExpr && <span className="ml-1">{item.matchExpr}</span>}

          {!item.selectedOptions && (
            <span className="flex items-center gap-1 max-w-[300px] whitespace-nowrap overflow-hidden text-ellipsis">
              {item.selected
                // ?.slice(0, 3)
                ?.map((selectedItem) => (
                  <ValueTag
                    key={selectedItem}
                    itemKey={item.key}
                    value={selectedItem}
                    display={undefined}
                  />
                ))}
              {/* {item.selected && item.selected.length > 2 && (
                <span className="text-gray-400">...</span>
              )} */}
            </span>
          )}
        </div>
      </Popover>

      <Button
        type="text"
        icon={<IoClose />}
        size="small"
        className="p-0 m-0"
        onClick={() => onDeleteFilter(item)}
      ></Button>
    </div>
  )
}
export default FilterTagItem
