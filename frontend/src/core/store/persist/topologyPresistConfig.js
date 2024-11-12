import sessionStorage from 'redux-persist/lib/storage/session' // 引入 sessionStorage

const topologyPresistConfig = {
  key: 'topology',
  storage: sessionStorage,
  blacklist: ['allTopologyData', 'displayData', 'modalDataUrl'],
  //   whitelist: ['name'], // 仅持久化 name 属性
}

export default topologyPresistConfig
