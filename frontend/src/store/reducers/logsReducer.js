export const logsInitialState = {
  tableInfo: {
    dataBase: '',
    tableName: '',
    parseName: '',

    cluster: '',
    instanceName: '',
    timeField: '',
  },
  logRules: [],
  instances: [],
  logs: [],
  pagination: {
    pageIndex: 1,
    pageSize: 50,
    total: 0,
  },

  logsChartData: [],
  defaultFields: [],

  hiddenFields: [],
  query: '',
  loading: true,

  // 保存字段和index索引map 当dataBase、时间改变清空
  fieldIndexMap: {},
}

const logsReducer = (state = logsInitialState, action) => {
  switch (action.type) {
    case 'setLogs':
      return { ...state, logs: action.payload }
    case 'setPagination':
      return { ...state, pagination: action.payload }
    case 'setLogsChartData':
      return { ...state, logsChartData: action.payload }
    case 'updateDefaultFields':
      return { ...state, defaultFields: action.payload }
    case 'updateHiddenFields':
      return { ...state, hiddenFields: action.payload }
    case 'updateQuery':
      return { ...state, query: action.payload }
    case 'updateLoading':
      return { ...state, loading: action.payload }
    case 'updateDataBase':
      return { ...state, database: action.payload }
    case 'updateTableInfo':
      return { ...state, tableInfo: action.payload }
    case 'setLogState':
      return { ...state, ...action.payload }
    case 'updateFieldIndexMap':
      //增量更新
      return { ...state, fieldIndexMap: { ...state.fieldIndexMap, ...action.payload } }
    case 'setLogRules':
      return { ...state, logRules: action.payload }
    case 'setInstances':
      return { ...state, instances: action.payload }
    case 'clearFieldIndexMap':
      return { ...state, fieldIndexMap: {} }
    default:
      return state
  }
}

export default logsReducer
