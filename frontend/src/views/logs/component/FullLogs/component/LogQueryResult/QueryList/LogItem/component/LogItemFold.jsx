import { Tag, Tooltip } from 'antd'
import React, { useEffect, useState } from 'react'
import LogTagDropDown from './LogTagDropdown'
const LogItemFold = ({ log }) => {
  return (
    <div className=" overflow-hidden whitespace-nowrap text-ellipsis text-wrap line-clamp-2 flex">
      {Object.entries(log.tags).map(
        ([key, value]) =>
          value &&
          key !== 'timestamp' && (
            <div key={key}>
              <LogTagDropDown
                objKey={key}
                value={value}
                children={
                  <Tooltip title={key + '= "' + value + '"'} key={key}>
                    <Tag className="max-w-[200px] overflow-hidden whitespace-nowrap text-ellipsis cursor-pointer text-gray-400">
                      {value}
                    </Tag>
                  </Tooltip>
                }
              ></LogTagDropDown>
            </div>
          ),
      )}
    </div>
  )
}
export default LogItemFold
