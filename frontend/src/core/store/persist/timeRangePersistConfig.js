/**
 * Copyright 2024 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */


import sessionStorage from 'redux-persist/lib/storage/session'; // 引入 sessionStorage

const timeRangePersistConfig = {
  key: 'timeRange',
  storage: sessionStorage,
//   whitelist: ['name'], // 仅持久化 name 属性
};

export default timeRangePersistConfig;
