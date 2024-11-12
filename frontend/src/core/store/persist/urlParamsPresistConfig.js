import sessionStorage from 'redux-persist/lib/storage/session' // 引入 sessionStorage

const urlParamsPresistConfig = {
  key: 'urlParams',
  storage: sessionStorage,
  //   blacklist: ['allTopologyData', 'displayData', 'modalDataUrl'],
  whitelist: [], // 仅持久化
}

export default urlParamsPresistConfig
