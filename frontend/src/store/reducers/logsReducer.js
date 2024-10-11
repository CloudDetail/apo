export const logsInitialState = {
  database: '',
  tableName: '',
  logRule: {},
  logs: [],
  pagination: {
    pageIndex: 1,
    pageSize: 10,
    total: 0,
  },
  logsChartData: [],
  defaultFields: [],
  hiddenFields: [],
  query: '',
  loading: true,
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
    case 'updateTableName':
      return { ...state, tableName: action.payload }
    case 'setLogState':
      return { ...state, ...action.payload }
    default:
      return state
  }
}

export default logsReducer
