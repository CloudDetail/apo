import React, { createContext, useContext, useEffect, useMemo, useReducer } from 'react'
import {
  getFullLogApi,
  getFullLogChartApi,
  getLogIndexApi,
  // @ts-ignore
  getLogRuleApi,
  getLogTableInfoAPi,
} from 'core/api/logs'
import { useDispatch, useSelector } from 'react-redux'

const LogsContext = createContext(null)

export const useLogsContext = () => useContext(LogsContext)

export const LogsProvider = ({ children }) => {
  const dispatch = useDispatch()
  const logs = useSelector((state) => state.logsReducer)
  const fetchData = async ({ startTime, endTime }) => {
    // @ts-ignore
    dispatch({ type: 'updateLoading', payload: true })

    try {
      const params = {
        startTime: startTime,
        endTime: endTime,
        pageNum: logs.pagination.pageIndex,
        pageSize: logs.pagination.pageSize,
        tableName: logs.tableInfo?.tableName,
        dataBase: logs.tableInfo?.dataBase,
        timeField: logs.tableInfo?.timeField,
        query: logs.query,
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

      //由tableName和type组成的唯一标识
      const tableId = `${logs.tableInfo.tableName}_${logs.tableInfo.type}`

      // @ts-ignore
      dispatch({
        type: 'setLogState',
        payload: {
          // @ts-ignore
          logs: res1?.logs ?? [],
          defaultFields: defaultFields,
          hiddenFields: hiddenFields,
          //判断displayFields对象中是否包含某个Table的信息，如果包含就不做处理，如果不包含就全部显示
          //${logs.tableInfo.tableName}_${logs.tableInfo.type}tableName+tableInfo作为唯一标识
          displayFields: Object.keys(logs.displayFields).some((key) => {
            return key === tableId
          }) ? logs.displayFields :
            { ...logs.displayFields, [tableId]: [...defaultFields, ...hiddenFields] },
          // @ts-ignore
          logsChartData: res2?.histograms ?? [],
          pagination: {
            // @ts-ignore
            total: res2?.count ?? 0,
            pageIndex: logs.pagination.pageIndex,
            pageSize: logs.pagination.pageSize,
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
            pageIndex: logs.pagination.pageIndex,
            pageSize: logs.pagination.pageSize,
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
        tableName: logs.tableInfo?.tableName,
        dataBase: logs.tableInfo?.dataBase,
        timeField: logs.tableInfo?.timeField,
        query: logs.query,
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
            type: 'logLibrary'
          },
        })
      }
    })
  }

  useEffect(() => {
    getLogTableInfo()
  }, [])

  const memoizedValue = useMemo(
    () => ({
      logs: logs.logs,
      pagination: logs.pagination,
      logsChartData: logs.logsChartData,
      defaultFields: logs.defaultFields,
      hiddenFields: logs.hiddenFields,
      displayFields: logs.displayFields,
      query: logs.query,
      loading: logs.loading,
      fieldIndexMap: logs.fieldIndexMap,
      tableInfo: logs.tableInfo,
      logRules: logs.logRules,
      instances: logs.instances,
      searchValue: logs.searchValue,
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
      updateHiddenFields: (data) => dispatch({ type: 'addHiddenFields', payload: data }),
      // @ts-ignore
      addDisplayFields: (data) => dispatch({ type: 'addDisplayFields', payload: data }),
      // @ts-ignore
      removeDisplayFields: (data) => dispatch({ type: 'removeDisplayFields', payload: data }),
      // @ts-ignore
      resetDisplayFields: (data) => dispatch({ type: 'resetDisplayFields', payload: data }),
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
      logs.logs,
      logs.pagination,
      logs.logsChartData,
      logs.defaultFields,
      logs.hiddenFields,
      logs.displayFields,
      logs.query,
      logs.loading,
      logs.fieldIndexMap,
      logs.tableInfo,
      logs.logRules,
      logs.instances,
      logs.searchValue,
    ],
  )

  // @ts-ignore
  return <LogsContext.Provider value={memoizedValue}>{children}</LogsContext.Provider>
}
