import { Button, Card } from 'antd'
import { ExclamationCircleFilled } from '@ant-design/icons'
import { useState } from 'react'
import UploadModal from './modals/UploadModal'
import { useTranslation } from 'react-i18next'

const UnlicensedCard = () => {
  const [uploadModalVisible, setUploadModalVisible] = useState(false)

  const { t } = useTranslation('pro/licenseManage')

  return (
    <Card>
      <div className="flex items-center gap-2 pb-4 border-b border-b-[var(--ant-color-border)]">
        <ExclamationCircleFilled className="ml-1 text-3xl text-[var(--ant-color-warning)]" />
        <p className="text-3xl font-bold m-0">{t('UnlicensedCard.title')}</p>
      </div>

      <div className="flex justify-between items-center gap-4 mt-4 h-48">
        <div className="p-2 w-1/2 h-full border border-[var(--ant-color-border)] flex flex-col items-center justify-around gap-2">
          <p className="text-2xl font-bold m-0">{t('UnlicensedCard.upload.title')}</p>
          <p className="text-base font-normal my-0 mx-8 text-center">
            {t('UnlicensedCard.upload.description')}
          </p>
          <Button
            type="primary"
            onClick={() => setUploadModalVisible(true)}
            className="flex-grow-0 flex-shrink-0"
          >
            <span className="font-medium text-base">{t('UnlicensedCard.upload.button')}</span>
          </Button>
        </div>

        {/* <div className="p-2 w-1/2 h-full border border-[var(--ant-color-border)] flex flex-col items-center justify-around gap-2">
          <p className="text-2xl font-bold m-0">{t('UnlicensedCard.trial.title')}</p>
          <p className="text-base font-normal my-0 mx-8 text-center">
            {t('UnlicensedCard.trial.description', {
              phone: '18888888888',
              email: '188@apo.com',
            })}
          </p>
          <Button
            type="primary"
            onClick={() => setTrialModalVisible(true)}
            className="flex-grow-0 flex-shrink-0"
          >
            <span className="font-medium text-base">{t('UnlicensedCard.trial.button')}</span>
          </Button>
        </div> */}
      </div>

      <UploadModal
        visible={uploadModalVisible}
        onClose={() => setUploadModalVisible(false)}
        onUploadSuccess={() => {
          setUploadModalVisible(false)
          // onUploadSuccess();
        }}
      />
      {/* <TrialModal visible={trialModalVisible} onClose={() => setTrialModalVisible(false)} /> */}
    </Card>
  )
}

export default UnlicensedCard
