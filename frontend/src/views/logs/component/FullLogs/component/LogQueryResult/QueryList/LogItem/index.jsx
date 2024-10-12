import React, { useEffect, useState } from 'react'
import { AiFillCaretDown, AiFillCaretRight } from 'react-icons/ai'
import LogItemFold from './LogItemFold'
import LogItemDetail from './LogItemDetail'
import { Button } from 'antd'
const LogItem = (props) => {
  const { log, foldingChecked } = props
  // 是否折叠日志，true 为是，false 为否
  const [isFold, setIsFold] = useState(true)

  const handleFoldClick = () => setIsFold(() => !isFold)
  useEffect(() => {
    setIsFold(foldingChecked ?? true)
  }, [foldingChecked])
  return (
    <div className="flex overflow-hidden px-2">
      {/* icon 和 时间 */}
      <div className="flex-grow-0 flex-shrink-0  w-[360px]">
        <div className="flex items-center pl-3">
          {/* <Button
            color="primary"
            type="text"
            onClick={() => setIsFold(!isFold)}
            className="mx-2"
            icon={isFold ? <AiFillCaretRight /> : <AiFillCaretDown />}
          ></Button> */}
          {log?.tags.timestamp}
        </div>
      </div>
      {/* 具体日志 */}
      <div className="flex-1 overflow-hidden">
        <LogItemFold log={log} />
        <LogItemDetail log={log} />
        {/* {isFold ? <LogItemFold log={log} /> : <LogItemDetail log={log} />} */}
      </div>
    </div>
  )
}
export default LogItem
