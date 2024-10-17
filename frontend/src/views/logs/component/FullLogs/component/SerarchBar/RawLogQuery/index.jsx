import React, { useEffect, useState } from 'react'
import CodeMirrorSearch from './CodeMirrorSearch'
import './index.less'
import { Button } from 'antd'
import { IoSearch } from 'react-icons/io5'
import { useLogsContext } from 'src/contexts/LogsContext'
import { useSearchParams } from 'react-router-dom'
import FullTextSearch from './FullTextSearch'
const RawLogQuery = () => {
  const { searchValue, setSearchValue, query, updateQuery, getLogTableInfo } = useLogsContext()
  // 分析字段的代码提示
  const [analysisFieldTips, setAnalysisFieldTips] = useState([])
  // 输入框自动填充历史记录
  const [historicalRecord, setHistoricalRecord] = useState([])
  const [isDefault, setIsDefault] = useState(true)
  const [queryKeyword, setQueryKeyword] = useState()
  const [isMultipleLines, setIsMultipleLines] = useState(false)

  const [searchParams] = useSearchParams()
  useEffect(() => {
    if (isDefault) {
      setSearchValue(queryKeyword)
      setIsDefault(false)
    }
  }, [queryKeyword, isDefault])
  useEffect(() => {
    setSearchValue(query)
  }, [query])

  const clickFullTextSearch = (value) => {
    setSearchValue(value)
    updateQuery(value)
  }

  return (
    <>
      <div className="searchBarMain">
        {/* <FullTextSearch searchValue={searchValue} setSearchValue={clickFullTextSearch} /> */}
        <div className="inputBox" style={{ overflowX: isMultipleLines ? 'visible' : 'hidden' }}>
          <CodeMirrorSearch
            title="logInput"
            value={searchValue}
            placeholder="请输入查询语句"
            onPressEnter={() => updateQuery(queryKeyword)}
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
        {/* <DateTimeRangePickerCom type="log" /> */}
        <Button
          type="primary"
          icon={<IoSearch />}
          onClick={() => updateQuery(queryKeyword)}
        ></Button>
        {/* <Button
          type="primary"
          icon={<LuRefreshCw />}
          className="ml-2"
          onClick={() => getLogTableInfo()}
        ></Button> */}
      </div>
    </>
  )
}

export default RawLogQuery
