// LicenseContext.tsx
import React, { createContext, useContext, useState, ReactNode, useEffect } from 'react'
import { getLicense } from '../api/license'
import { notify } from 'src/core/utils/notify'
import { useTranslation } from 'react-i18next'
import { useSelector } from 'react-redux'
import { isSystemHealthPreviewRoute } from 'src/core/utils/systemHealthRoute'

type LicenseInfo = {
  licenseId: string
  appId: string
  appVersion: string
  mode: number
  type: number
  active: number
  issuedAt: number
  expiresAt: number
  maxNodes: number
  seatsInLicense: number
  seatsCurrentlyInUse: number
}

// 定义 Context 的类型
interface LicenseContextType {
  licenseInfo: LicenseInfo
  setLicenseInfo: (licenseInfo: LicenseInfo) => void
  fetchLicenseInfo: () => void
}

// 创建 Context
const LicenseContext = createContext<LicenseContextType | undefined>(undefined)

// 默认的license
const defaultLicense: LicenseInfo = {
  licenseId: '',
  appId: '',
  appVersion: '',
  mode: 0,
  type: -1,
  active: null,
  issuedAt: 0,
  expiresAt: 0,
  maxNodes: 0,
  seatsInLicense: 0,
  seatsCurrentlyInUse: 0,
}

// 创建 Provider 组件
export const LicenseProvider = ({ children }: { children: ReactNode }) => {
  const [licenseInfo, setLicenseInfo] = useState<LicenseInfo>(defaultLicense)
  const { t } = useTranslation('pro/licenseManage')
  const { user } = useSelector((state) => state.userReducer)
  const fetchLicenseInfo = async () => {
    try {
      const res = await getLicense()
      setLicenseInfo(res)
    } catch (error) {
      notify({ message: t('error.fetch'), type: 'error' })
    }
  }
  useEffect(() => {
    if (user.userId) {
      if (isSystemHealthPreviewRoute()) return
      fetchLicenseInfo()
    }
  }, [user.userId])

  return (
    <LicenseContext.Provider value={{ licenseInfo, setLicenseInfo, fetchLicenseInfo }}>
      {children}
    </LicenseContext.Provider>
  )
}

// 自定义 Hook，方便使用 Context
export const useLicense = (): LicenseContextType => {
  const context = useContext(LicenseContext)
  if (!context) {
    throw new Error('useLicense must be used within a LicenseProvider')
  }

  return context
}
