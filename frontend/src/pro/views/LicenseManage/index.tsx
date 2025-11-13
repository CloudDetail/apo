import React from 'react'
import LicenseInfoCard from './components/LicenseInfoCard'
import { useLicense } from 'src/pro/contexts/LicenseContext'
import UnlicensedCard from './components/UnlicensedCard'
import { useTranslation } from 'react-i18next'

function getDaysDifference(timestamp: number) {
  const now = Date.now() // milliseconds
  const diff = timestamp - now // milliseconds
  return Math.floor(diff / 86400000) // 86400000 = 24*60*60*1000
}

function getDuration(startDate: number, endDate: number) {
  const diff = endDate - startDate
  return Math.floor(diff / 86400000)
}

const LicenseManagementPage = () => {
  const { licenseInfo } = useLicense()

  const { t } = useTranslation('pro/licenseManage')

  const hasLicense = licenseInfo.active === 1 // 已激活
  const isLicenseTrial = licenseInfo.active === 0 // 未激活
  const isLicenseExpired = licenseInfo.active === 2 // 已过期
  const licenseMode = licenseInfo.mode === 0 ? t('mode.public') : t('mode.private')

  const activeStatus = {
    0: t('activeStatus.0'),
    1: t('activeStatus.1'),
    2: t('activeStatus.2'),
    '-1': t('activeStatus.-1'),
  }

  const licenseData = {
    feature: {
      active: activeStatus[licenseInfo.active],
      seatsInLicense: licenseInfo.seatsInLicense,
      seatsCurrentlyInUse: licenseInfo.seatsCurrentlyInUse,
      maxNodes: licenseInfo.maxNodes,
    },
    validity: {
      startDate: new Date(licenseInfo.issuedAt).toLocaleDateString(),
      endDate: new Date(licenseInfo.expiresAt).toLocaleDateString(),
      remainDate: getDaysDifference(licenseInfo.expiresAt),
      // duration: `${getDuration(licenseInfo.issuedAt, licenseInfo.expiresAt)} ${t('validity.days')}`,
    },
    metaData: {
      appId: licenseInfo.appId,
      licenseId: licenseInfo.licenseId,
      version: licenseInfo.appVersion,
    },
    hasLicense: hasLicense,
    isLicenseTrial: isLicenseTrial,
    isLicenseExpired: isLicenseExpired,
    mode: licenseMode,
    type: licenseInfo.type,
  }

  return (
    <>
      {hasLicense && <LicenseInfoCard licenseData={licenseData} />}
      {/* prepare for trial */}
      {!hasLicense && <UnlicensedCard />}
    </>
  )
}

export default LicenseManagementPage
