/**
 * Copyright 2024 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import React, { createContext, useContext } from 'react';

const PropsContext = createContext({});

export const usePropsContext = () => useContext(PropsContext);

export const PropsProvider = ({ children, value }) => (
  <PropsContext.Provider value={value}>
    {children}
  </PropsContext.Provider>
);
