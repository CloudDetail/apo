import storage from 'redux-persist/lib/storage';

const settingPersistConfig = {
  key: 'setting',
  storage,
//   whitelist: ['name'], // 仅持久化 name 属性
};

export default settingPersistConfig;
