import React, { createContext, useContext, useEffect, useMemo, useReducer } from 'react'
import {
  getFullLogApi,
  getFullLogChartApi,
  getLogIndexApi,
  // @ts-ignore
  getLogRuleApi,
  getLogTableInfoAPi,
} from 'src/api/logs'
import logsReducer, { logsInitialState } from 'src/store/reducers/logsReducer'

const LogsContext = createContext(logsInitialState)

export const useLogsContext = () => useContext(LogsContext)

export const LogsProvider = ({ children }) => {
  const [state, dispatch] = useReducer(logsReducer, logsInitialState)
  const fetchData = async ({ startTime, endTime }) => {
    // @ts-ignore
    dispatch({ type: 'updateLoading', payload: true })

    try {
      const params = {
        startTime: startTime,
        endTime: endTime,
        pageNum: state.pagination.pageIndex,
        pageSize: state.pagination.pageSize,
        tableName: state.tableInfo?.tableName,
        dataBase: state.tableInfo?.dataBase,
        timeField: state.tableInfo?.timeField,
        query: state.query,
      }

      const [res1, res2] = await Promise.all([
        getFullLogApi(params),
        getFullLogChartApi(params),
        // getLogRuleApi({ tableName: 'test_logs', dataBase: 'default' }),
      ])
      // @ts-ignore
      let defaultFields = (res1?.defaultFields ?? []).sort()
      // @ts-ignore
      let hiddenFields = (res1?.hiddenFields ?? []).sort()
      // @ts-ignore
      dispatch({
        type: 'setLogState',
        payload: {
          // @ts-ignore
          logs: res1?.logs ?? [],
          defaultFields: defaultFields,
          hiddenFields: hiddenFields,
          // @ts-ignore
          logsChartData: res2?.histograms ?? [],
          pagination: {
            // @ts-ignore
            total: res2?.count ?? 0,
            pageIndex: state.pagination.pageIndex,
            pageSize: state.pagination.pageSize,
          },
          // logRule: res3,
        },
      })
    } catch (error) {
      console.error('请求出错:', error)
      // @ts-ignore
      dispatch({
        type: 'setLogState',
        payload: {
          logs: [],
          defaultFields: [],
          hiddenFields: [],
          logsChartData: [],
          loading: false,
          pagination: {
            total: 0,
            pageIndex: state.pagination.pageIndex,
            pageSize: state.pagination.pageSize,
          },
          // logRule: res3,
        },
      })
    } finally {
      // @ts-ignore
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
        timeField: state.tableInfo?.timeField,
        query: state.query,
      })

      // @ts-ignore
      dispatch({
        type: 'updateFieldIndexMap',
        payload: {
          // @ts-ignore
          [column]: res.indexs ?? [],
        },
      })

      return res // 返回响应结果，方便调用方处理
    } catch (error) {
      dispatch({
        type: 'updateFieldIndexMap',
        payload: {
          // @ts-ignore
          [column]: [],
        },
      })
      console.error('Error fetching field index data:', error)
      throw error // 如果发生错误，可以抛出异常让调用方处理
    }
  }
  const getLogTableInfo = () => {
    // @ts-ignore
    dispatch({ type: 'updateLoading', payload: true })
    getLogTableInfoAPi().then((res) => {
      // @ts-ignore
      dispatch({ type: 'setLogRules', payload: res.parses ?? [] })
      // @ts-ignore

      dispatch({ type: 'setInstances', payload: res.instances ?? [] })
      if (res?.parses?.length > 0) {
        // @ts-ignore
        dispatch({
          type: 'updateTableInfo',
          payload: {
            dataBase: res.parses[0].dataBase,
            tableName: res.parses[0].tableName,
            parseName: res.parses[0]?.parseName,
          },
        })
      }
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
      logRules: state.logRules,
      instances: state.instances,
      searchValue: state.searchValue,
      fetchData,
      getLogTableInfo,
      getFieldIndexData,
      // @ts-ignore
      updateLogs: (logs) => dispatch({ type: 'setLogs', payload: logs }),
      updateLogsPagination: (pagination) =>
        // @ts-ignore
        dispatch({ type: 'setPagination', payload: { ...state.pagination, ...pagination } }),
      // @ts-ignore
      updateLogsChartData: (data) => dispatch({ type: 'setLogsChartData', payload: data }),
      // @ts-ignore
      updateDefaultFields: (data) => dispatch({ type: 'updateDefaultFields', payload: data }),
      // @ts-ignore
      updateHiddenFields: (data) => dispatch({ type: 'updateHiddenFields', payload: data }),
      // @ts-ignore
      updateQuery: (data) => dispatch({ type: 'updateQuery', payload: data }),
      // @ts-ignore
      updateTableInfo: (data) => dispatch({ type: 'updateTableInfo', payload: data }),
      // @ts-ignore
      setLogRules: (data) => dispatch({ type: 'setLogRules', payload: data }),
      // @ts-ignore
      setInstances: (data) => dispatch({ type: 'setInstances', payload: data }),
      // @ts-ignore
      updateLoading: (data) => dispatch({ type: 'updateLoading', payload: data }),
      // @ts-ignore
      setSearchValue: (data) => dispatch({ type: 'setSearchValue', payload: data }),
      // @ts-ignore
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
      state.logRules,
      state.instances,
      state.searchValue,
    ],
  )

  // @ts-ignore
  return <LogsContext.Provider value={memoizedValue}>{children}</LogsContext.Provider>
}
