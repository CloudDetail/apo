import React, { useState } from 'react'
import { Button, Descriptions } from 'antd'
import { useTranslation } from 'react-i18next'
import { UploadOutlined } from '@ant-design/icons'
import UploadModal from './modals/UploadModal'
import { getLicense } from 'src/pro/api/license'
import { notify } from 'src/core/utils/notify'
import { useLicense } from 'src/pro/contexts/LicenseContext'
import { BasicCard } from 'src/core/components/Card/BasicCard'

const LicenseInfoCard = ({ licenseData }) => {
  const feature = licenseData?.feature
  const validity = licenseData?.validity
  const metaData = licenseData?.metaData

  const isLicenseExpired = licenseData?.isLicenseExpired

  const [uploadModalVisible, setUploadModalVisible] = useState(false)

  const { t } = useTranslation('pro/licenseManage')

  const { setLicenseInfo } = useLicense()

  const fetchLicenseInfo = async () => {
    try {
      const res = await getLicense()
      setLicenseInfo(res)
    } catch {
      notify({ message: t('error.fetch'), type: 'error' })
    }
  }

  return (
    <BasicCard>
      {/* CardHeader */}
      <div className="w-full flex items-center justify-between pb-4">
        <div className="flex items-center gap-2 text-xl font-bold text-[var(--ant-color-text)]">
          {t('index.trialLicense')}
          {/* APO Enterprise Edition Trial */}

          {isLicenseExpired && (
            <div className="text-xl font-semibold flex items-center">
              |<span className="text-red-500 ml-4">{t('pro/licenseManage:index.expired')} ⚠️</span>
            </div>
          )}
          {/* {
            true && (
              <div className='text-xl font-semibold flex items-center'>
                |
              <span className= 'ml-4'>
                {licenseData.mode} {t('licenseInfo.trial')}, {t('licenseInfo.remain', { days: validity?.remainDate})}
              </span>
              </div>
            )
          } */}
        </div>

        <Button
          type="primary"
          icon={<UploadOutlined />}
          onClick={() => setUploadModalVisible(true)}
          className="text-sm"
        >
          {t('upload.title')}
        </Button>
      </div>

      <Descriptions
        bordered
        column={2}
        size="small"
        className="mb-4"
        labelStyle={{ width: '240px' }}
      >
        <Descriptions.Item span={3} label={t('metaData.license')}>
          {metaData?.licenseId}
        </Descriptions.Item>

        <Descriptions.Item label={t('feature.activeStatus')}>{feature?.active}</Descriptions.Item>
        <Descriptions.Item label={t('feature.seatsCurrentlyInUse')}>
          {feature?.seatsCurrentlyInUse + '/' + feature?.seatsInLicense}
        </Descriptions.Item>
        <Descriptions.Item label={t('validity.startDate')}>{validity?.startDate}</Descriptions.Item>
        <Descriptions.Item label={t('validity.endDate')}>
          {licenseData?.type === 'permanent' ? t('validity.permanent') : validity?.endDate}
        </Descriptions.Item>

        {/* <Descriptions.Item label={t('validity.duration')}>{validity?.duration}</Descriptions.Item> */}
      </Descriptions>

      <UploadModal
        visible={uploadModalVisible}
        onClose={() => setUploadModalVisible(false)}
        onUploadSuccess={() => {
          setUploadModalVisible(false)
          fetchLicenseInfo()
        }}
      />
    </BasicCard>
  )
}

export default LicenseInfoCard
