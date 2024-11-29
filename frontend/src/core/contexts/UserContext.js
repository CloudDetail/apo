import React, { createContext, useContext, useMemo, useReducer, useState } from 'react'
import userReducer, { initialState } from '../store/reducers/userReducer'

const UserContext = createContext({})

export const useUserContext = () => useContext(UserContext)

export const UserProvider = ({ children }) => {
    const [state, dispatch] = useReducer(userReducer, initialState)
    const value = {
        user: state,
        dispatchUser: dispatch
    }

    return (
        <>
            <UserContext.Provider value={value}>{children}</UserContext.Provider>
        </>
    )
}
