/**
 * Copyright 2024 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import storage from 'redux-persist/lib/storage'

const logsPresistConfig = {
    key: 'logs',
    storage,
    whitelist: ['displayFields']
}

export default logsPresistConfig
