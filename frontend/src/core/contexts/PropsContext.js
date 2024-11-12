import React, { createContext, useContext } from 'react';

const PropsContext = createContext({});

export const usePropsContext = () => useContext(PropsContext);

export const PropsProvider = ({ children, value }) => (
  <PropsContext.Provider value={value}>
    {children}
  </PropsContext.Provider>
);
