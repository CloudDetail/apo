const initialState = {
    logsField: {}
}

const personalizedSettingReducer = (state = initialState, action) => {
    switch (action.type) {
        case 'updateLogsField':
            return { ...state, logsField: { ...state.logsField, ...action.payload } }
        default:
            return state
    }
}

export default personalizedSettingReducer