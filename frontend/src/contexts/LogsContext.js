import React, { createContext, useContext, useMemo, useReducer } from 'react'
import { getFullLogApi, getFullLogChartApi, getLogRuleApi } from 'src/api/logs'
import logsReducer, { logsInitialState } from 'src/store/reducers/logsReducer'
import { ISOToTimestamp } from 'src/utils/time'

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
        tableName: 'test_logs',
        dataBase: 'default',
        query: state.query,
      }

      const [res1, res2] = await Promise.all([
        getFullLogApi(params),
        getFullLogChartApi(params),
        // getLogRuleApi({ tableName: 'test_logs', dataBase: 'default' }),
      ])
      dispatch({
        type: 'setLogState',
        payload: {
          logs: res1?.logs ?? [],
          defaultFields: res1?.defaultFields ?? [],
          hiddenFields: res1?.hiddenFields ?? [],
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
    } finally {
      dispatch({ type: 'updateLoading', payload: false })
    }
  }

  const memoizedValue = useMemo(
    () => ({
      logs: state.logs,
      pagination: state.pagination,
      logsChartData: state.logsChartData,
      defaultFields: state.defaultFields,
      hiddenFields: state.hiddenFields,
      query: state.query,
      loading: state.loading,
      fetchData,
      updateLogs: (logs) => dispatch({ type: 'setLogs', payload: logs }),
      updateLogsPagination: (pagination) =>
        dispatch({ type: 'setPagination', payload: pagination }),
      updateLogsChartData: (data) => dispatch({ type: 'setLogsChartData', payload: data }),
      updateDefaultFields: (data) => dispatch({ type: 'updateDefaultFields', payload: data }),
      updateHiddenFields: (data) => dispatch({ type: 'updateHiddenFields', payload: data }),
      updateQuery: (data) => dispatch({ type: 'updateQuery', payload: data }),
    }),
    [
      state.logs,
      state.pagination,
      state.logsChartData,
      state.defaultFields,
      state.hiddenFields,
      state.query,
      state.loading,
    ],
  )

  return <LogsContext.Provider value={memoizedValue}>{children}</LogsContext.Provider>
}
