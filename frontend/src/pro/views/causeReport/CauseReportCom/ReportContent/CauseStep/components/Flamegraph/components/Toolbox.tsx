import { Button, Input } from 'antd'
import React from 'react'
import { IoMdTime } from 'react-icons/io'
import { RiArrowGoBackFill } from 'react-icons/ri'
import { convertTime } from 'src/core/utils/time'
interface ToolboxProps {
  searchValue: string
  setSearchValue: any
  isFocused: boolean
  resetView: any
  timeRange?: {
    startTime: number
    endTime: number
  }
}
const Toolbox = ({
  searchValue,
  setSearchValue,
  isFocused,
  resetView,
  timeRange,
}: ToolboxProps) => {
  return (
    <div className="flex justify-between mb-3">
      <div className="flex">
        <Input.Search
          placeholder="input search text"
          value={searchValue}
          onChange={(e) => setSearchValue(e.target.value)}
          style={{ width: 200 }}
          className="mr-3"
        />

        {timeRange && (
          <div className="flex items-center">
            <IoMdTime size={20} />
            {convertTime(timeRange?.startTime / 1000, 'yyyy-mm-dd hh:mm:ss')} to{' '}
            {convertTime(timeRange?.endTime / 1000, 'yyyy-mm-dd hh:mm:ss')}
          </div>
        )}
      </div>
      <Button disabled={!isFocused} onClick={resetView}>
        <RiArrowGoBackFill />
      </Button>
    </div>
  )
}
export default React.memo(Toolbox)
