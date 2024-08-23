import { combineReducers } from 'redux'
import { persistReducer } from 'redux-persist'
import timeRangePersistConfig from '../persist/timeRangePersistConfig'
import timeRangeReducer from './timeRangeReducer'
import settingPersistConfig from '../persist/settingPresistConfig'
import settingReducer from './settingReducer'
import topologyPresistConfig from '../persist/topologyPresistConfig'
import topologyReducer from './topologyReducer'
import urlParamsReducer from './urlParamsReducer'
import urlParamsPresistConfig from '../persist/urlParamsPresistConfig'

const rootReducer = combineReducers({
  timeRange: persistReducer(timeRangePersistConfig, timeRangeReducer),
  settingReducer: persistReducer(settingPersistConfig, settingReducer),
  topologyReducer: persistReducer(topologyPresistConfig, topologyReducer),
  urlParamsReducer: persistReducer(urlParamsPresistConfig, urlParamsReducer),
})

export default rootReducer
