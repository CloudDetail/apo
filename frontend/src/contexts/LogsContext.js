import React, { createContext, useContext, useEffect, useMemo, useReducer } from 'react'
import {
  getFullLogApi,
  getFullLogChartApi,
  getLogIndexApi,
  getLogRuleApi,
  getLogTableInfoAPi,
} from 'src/api/logs'
import logsReducer, { logsInitialState } from 'src/store/reducers/logsReducer'

const LogsContext = createContext(logsInitialState)

export const useLogsContext = () => useContext(LogsContext)

export const LogsProvider = ({ children }) => {
  const [state, dispatch] = useReducer(logsReducer, logsInitialState)
  const fetchData = async ({ startTime, endTime }) => {
    dispatch({ type: 'updateLoading', payload: true })

    try {
      const params = {
        startTime: startTime,
        endTime: endTime,
        pageNum: state.pagination.pageIndex,
        pageSize: state.pagination.pageSize,
        tableName: state.tableInfo?.tableName,
        dataBase: state.tableInfo?.dataBase,
        query: state.query,
      }

      const [res1, res2] = await Promise.all([
        getFullLogApi(params),
        getFullLogChartApi(params),
        // getLogRuleApi({ tableName: 'test_logs', dataBase: 'default' }),
      ])
      let defaultFields = (res1?.defaultFields ?? []).sort()
      let hiddenFields = (res1?.hiddenFields ?? []).sort()
      dispatch({
        type: 'setLogState',
        payload: {
          logs: res1?.logs ?? [],
          defaultFields: defaultFields,
          hiddenFields: hiddenFields,
          logsChartData: res2?.histograms ?? [],
          pagination: {
            total: res2?.count ?? 0,
            pageIndex: state.pagination.pageIndex,
            pageSize: state.pagination.pageSize,
          },
          // logRule: res3,
        },
      })
    } catch (error) {
      console.error('请求出错:', error)
      dispatch({
        type: 'setLogState',
        payload: {
          logs: [],
          defaultFields: [],
          hiddenFields: [],
          logsChartData: [],
          pagination: {
            total: 0,
            pageIndex: state.pagination.pageIndex,
            pageSize: state.pagination.pageSize,
          },
          // logRule: res3,
        },
      })
    } finally {
      dispatch({ type: 'updateLoading', payload: false })
    }
  }

  const getFieldIndexData = async ({ startTime, endTime, column }) => {
    try {
      const res = await getLogIndexApi({
        startTime,
        endTime,
        column,
        tableName: state.tableInfo?.tableName,
        dataBase: state.tableInfo?.dataBase,
        query: state.query,
      })

      dispatch({
        type: 'updateFieldIndexMap',
        payload: {
          [column]: res.indexs,
        },
      })

      return res // 返回响应结果，方便调用方处理
    } catch (error) {
      console.error('Error fetching field index data:', error)
      throw error // 如果发生错误，可以抛出异常让调用方处理
    }
  }

  const getLogTableInfo = () => {
    getLogTableInfoAPi().then((res) => {
      const dataBase = Object.keys(res.logTables)[0]
      const tableList = res.logTables[dataBase][0]
      dispatch({
        type: 'updateTableInfo',
        payload: {
          dataBase: dataBase,
          tableName: tableList?.tableName,
          cluster: tableList?.cluster,
        },
      })
    })
  }
  useEffect(() => {
    console.log('获取database')
    getLogTableInfo()
  }, [])
  const memoizedValue = useMemo(
    () => ({
      logs: state.logs,
      pagination: state.pagination,
      logsChartData: state.logsChartData,
      defaultFields: state.defaultFields,
      hiddenFields: state.hiddenFields,
      query: state.query,
      loading: state.loading,
      fieldIndexMap: state.fieldIndexMap,
      tableInfo: state.tableInfo,
      fetchData,
      getFieldIndexData,
      updateLogs: (logs) => dispatch({ type: 'setLogs', payload: logs }),
      updateLogsPagination: (pagination) =>
        dispatch({ type: 'setPagination', payload: { ...state.pagination, ...pagination } }),
      updateLogsChartData: (data) => dispatch({ type: 'setLogsChartData', payload: data }),
      updateDefaultFields: (data) => dispatch({ type: 'updateDefaultFields', payload: data }),
      updateHiddenFields: (data) => dispatch({ type: 'updateHiddenFields', payload: data }),
      updateQuery: (data) => dispatch({ type: 'updateQuery', payload: data }),
      updateTableName: (data) => dispatch({ type: 'updateTableName', payload: data }),
      clearFieldIndexMap: () => dispatch({ type: 'clearFieldIndexMap' }),
    }),
    [
      state.logs,
      state.pagination,
      state.logsChartData,
      state.defaultFields,
      state.hiddenFields,
      state.query,
      state.loading,
      state.fieldIndexMap,
      state.tableInfo,
    ],
  )

  return <LogsContext.Provider value={memoizedValue}>{children}</LogsContext.Provider>
}
