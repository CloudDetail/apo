/**
 * Copyright 2024 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import storage from 'redux-persist/lib/storage'

const userPersistConfig = {
  key: 'user',
  storage,
}

export default userPersistConfig
