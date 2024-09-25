import { createSelector } from 'reselect'

const initialState = {
  groupLabel: {},
}

const groupLabelReducer = (state = initialState, action) => {
  switch (action.type) {
    case 'setGroupLabel':
      return { ...state, groupLabel: action.payload }
    default:
      return state
  }
}

export default groupLabelReducer
