/**
 * Copyright 2024 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import React, { createContext, useContext, useState, useEffect } from 'react'
import { getDeployModeApi } from '../api/app'
import { isProbeManagementPreviewRoute } from '../utils/probeManagementRoute'

const DeployModeContext = createContext(null)

export const useDeployMode = () => {
  const context = useContext(DeployModeContext)
  if (!context) {
    throw new Error('useDeployMode must be used within a DeployModeProvider')
  }
  return context
}

export const DeployModeProvider = ({ children }) => {
  const [deployMode, setDeployMode] = useState('apo') // 默认为 'apo'
  const [loading, setLoading] = useState(true)

  const fetchDeployMode = async () => {
    try {
      setLoading(true)
      const response = await getDeployModeApi()
      // 假设 API 返回格式为 { deployMode: 'apo' | 'syncause' }
      setDeployMode(response || 'apo')
    } catch (error) {
      console.error('Failed to fetch deploy mode:', error)
      // 出错时使用默认值
      setDeployMode('apo')
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    if (isProbeManagementPreviewRoute()) {
      setLoading(false)
      return
    }
    fetchDeployMode()
  }, [])

  const getLogo = () => {
    return deployMode === 'syncause' ? '/syncause.svg' : '/apo.svg'
  }

  const getFavicon = () => {
    return deployMode === 'syncause' ? '/syncause.svg' : '/favicon.ico'
  }

  const getTitle = () => {
    return deployMode === 'syncause' ? 'Syncause' : 'APO'
  }

  const value = {
    deployMode,
    loading,
    getLogo,
    getFavicon,
    getTitle,
    refetch: fetchDeployMode,
  }

  return <DeployModeContext.Provider value={value}>{children}</DeployModeContext.Provider>
}
