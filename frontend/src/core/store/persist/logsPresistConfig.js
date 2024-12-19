import storage from 'redux-persist/lib/storage'

const logsPresistConfig = {
    key: 'logs',
    storage,
    whitelist: ['displayFields']
}

export default logsPresistConfig
