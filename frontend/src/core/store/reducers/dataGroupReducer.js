/**
 * Copyright 2024 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */
const initialState = {
  dataGroupId: '',
}

export default function dataGroupReducer(state = initialState, action) {
  switch (action.type) {
    case 'setSelectedDataGroupId':
      return { ...state, dataGroupId: action.payload }
    default:
      return state
  }
}
