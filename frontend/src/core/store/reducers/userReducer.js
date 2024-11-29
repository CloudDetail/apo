export const initialState = {
    user: { username: "anonymous" },
    token: { token: null, refreshToken: null }
}

const userReducer = (state = initialState, action) => {
    switch (action.type) {
        case "addUser":
            return { user: action.payload }
        case "removeUser":
            return { user: "anonymous" }
        default:
            return state
    }
}

export default userReducer