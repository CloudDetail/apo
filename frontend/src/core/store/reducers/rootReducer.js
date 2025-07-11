/**
 * Copyright 2024 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

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
import groupLabelReducer from './groupLabelReducer'
import groupLabelPresistConfig from '../persist/groupLabelPresistConfig'
import userReducer from './userReducer'
import userPersistConfig from '../persist/userPresistConfig'
import logsReducer from './logsReducer'
import logsPresistConfig from '../persist/logsPresistConfig'
import dataGroupPersistConfig from '../persist/dataGroupConfig'
import dataGroupReducer from './dataGroupReducer'

const rootReducer = combineReducers({
  timeRange: persistReducer(timeRangePersistConfig, timeRangeReducer),
  settingReducer: persistReducer(settingPersistConfig, settingReducer),
  topologyReducer: persistReducer(topologyPresistConfig, topologyReducer),
  urlParamsReducer: persistReducer(urlParamsPresistConfig, urlParamsReducer),
  groupLabelReducer: persistReducer(groupLabelPresistConfig, groupLabelReducer),
  userReducer: persistReducer(userPersistConfig, userReducer),
  logsReducer: persistReducer(logsPresistConfig, logsReducer),
  dataGroupReducer: persistReducer(dataGroupPersistConfig, dataGroupReducer),
})

export default rootReducer
