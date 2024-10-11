import React, { useEffect, useState } from 'react'
import CodeMirrorSearch from './CodeMirrorSearch'
import './index.less'
import DateTimeRangePickerCom from 'src/components/DateTime/DateTimeRangePickerCom'
import { Button } from 'antd'
import { IoSearch } from 'react-icons/io5'
import { useLogsContext } from 'src/contexts/LogsContext'
import { ISOToTimestamp } from 'src/utils/time'
import { useSearchParams } from 'react-router-dom'
const RawLogQuery = () => {
  const { query, updateQuery, fetchData } = useLogsContext()
  // 分析字段的代码提示
  const [analysisFieldTips, setAnalysisFieldTips] = useState([])
  // 输入框自动填充历史记录
  const [historicalRecord, setHistoricalRecord] = useState([])
  const [isDefault, setIsDefault] = useState(true)
  const [queryKeyword, setQueryKeyword] = useState()
  const [searchValue, setSearchValue] = useState('')
  const [isMultipleLines, setIsMultipleLines] = useState(false)

  const [searchParams] = useSearchParams()
  useEffect(() => {
    if (queryKeyword) {
      setSearchValue(queryKeyword)
      setIsDefault(false)
    }
  }, [queryKeyword, isDefault])
  useEffect(() => {
    setSearchValue(query)
  }, [query])

  return (
    <>
      <div className="searchBarMain">
        <div className="inputBox" style={{ overflowX: isMultipleLines ? 'visible' : 'hidden' }}>
          <CodeMirrorSearch
            title="logInput"
            value={searchValue}
            // onPressEnter={() => doSearchLog.run()}
            onChange={setQueryKeyword}
            tables={analysisFieldTips}
            historicalRecord={historicalRecord}
            onChangeHistoricalRecord={setHistoricalRecord}
            // currentTid={currentLogLibrary?.id as number}
            // logQueryHistoricalList={logQueryHistoricalList}
            // collectingHistorical={collectingHistorical}
            isMultipleLines={isMultipleLines}
            onChangeIsMultipleLines={setIsMultipleLines}
            onChangeIsDefault={setIsDefault}
          />
        </div>
        <DateTimeRangePickerCom type="log" />
        <Button
          type="primary"
          icon={<IoSearch />}
          onClick={() =>
            fetchData({
              startTime: ISOToTimestamp(searchParams.get('log-from')),
              endTime: ISOToTimestamp(searchParams.get('log-to')),
            })
          }
        ></Button>
      </div>
    </>
  )
}

export default RawLogQuery
