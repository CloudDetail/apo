import React, { useEffect, useState } from 'react'
import { AiFillCaretDown, AiFillCaretRight } from 'react-icons/ai'
import LogItemFold from './component/LogItemFold'
import LogItemDetail from './component/LogItemDetail'
import { Button } from 'antd'
import { useLogsContext } from 'src/contexts/LogsContext'
import { convertTime } from 'src/utils/time'

// 自配类规则日志默认展开不可收起，tag+铺平（仅content）
// 接入类数据库规则默认收起可展开，收起展示所有tag，展开展示所有（content + tag）
const LogItem = (props) => {
  const { log, foldingChecked } = props
  const { tableInfo } = useLogsContext()
  // 是否折叠日志，true 为是，false 为否
  const [isFold, setIsFold] = useState(true)

  useEffect(() => {
    setIsFold(foldingChecked ?? true)
  }, [foldingChecked])
  return (
    <div className="flex overflow-hidden px-2">
      {/* icon 和 时间 */}
      <div className="flex-grow-0 flex-shrink-0  w-[230px]">
        <div className="flex items-center pl-3">
          {tableInfo.timeField && (
            <Button
              color="primary"
              type="text"
              onClick={() => setIsFold(!isFold)}
              className="mx-2"
              icon={isFold ? <AiFillCaretRight /> : <AiFillCaretDown />}
            ></Button>
          )}
          {convertTime(log?.timestamp, 'yyyy-mm-dd hh:mm:ss')}
        </div>
      </div>
      {/* 具体日志 */}
      <div className="flex-1 overflow-hidden">
        {/* <LogItemFold log={log} isFold={isFold} />
        <LogItemDetail log={log} isFold={isFold} /> */}
        {/* <LogItemFold tags={!tableInfo?.timeField || isFold ? log.tags : []} /> */}
        {tableInfo.timeField ? (
          isFold ? (
            <LogItemFold tags={log.tags} />
          ) : (
            <LogItemDetail log={log} />
          )
        ) : (
          <>
            <LogItemFold tags={log.tags} />
            <LogItemDetail log={log} />
          </>
        )}
      </div>
    </div>
  )
}
export default LogItem
