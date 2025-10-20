import { Button, Modal, Upload } from "antd"
import { useTranslation } from "react-i18next"

const TrialModal = ({ visible, onClose }) => {

  const { t } = useTranslation('pro/licenseManage')

  return (
    <Modal
      title={<p className='text-2xl font-bold'>{t('trial.title')}</p>}
      open={visible}
      onCancel={onClose}
      footer={null}
      destroyOnClose
    >
      <div className="flex flex-col space-y-4">
        <p className='my-2 text-base'>
          {t('trial.description1')}
        </p>
        <p className='my-2 text-base'>
          {t('trial.description2_part1')}
          <ul className="list-disc list-inside ml-2 mt-2">
            <li>
              {t('trial.phone')} 18888888888
            </li>
            <li>
              {t('trial.email')} 188@apo.com
            </li>
          </ul>
        </p>

        <div className="flex justify-end space-x-4 mt-4">
          <Button onClick={onClose}>{t('trial.cancel')}</Button>
          <Button type="primary" onClick={() => window.open('https://baidu.com', '_blank')}>
            {/* {t('upload.confirm')} */}
            {t('trial.start')}
          </Button>
        </div>
      </div>
    </Modal>
  )
}

export default TrialModal