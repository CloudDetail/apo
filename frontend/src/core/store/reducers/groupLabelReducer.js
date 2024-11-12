import { createSelector } from 'reselect'

const initialState = {
  groupLabel: {},
  // 为了下拉选择
  groupLabelSelectOptions: [],
}

const groupLabelReducer = (state = initialState, action) => {
  switch (action.type) {
    case 'setGroupLabel':
      let groupLabelSelectOptions = Object.entries(action.payload).map(([key, value]) => ({
        label: value,
        value: key,
      }))
      return { ...state, groupLabel: action.payload, groupLabelSelectOptions }
    default:
      return state
  }
}

export default groupLabelReducer
