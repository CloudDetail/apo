import {
  CButton,
  CForm,
  CFormInput,
  CFormLabel,
  CInputGroup,
  CInputGroupText,
  CModal,
  CModalBody,
  CModalFooter,
  CModalHeader,
  CModalTitle,
} from '@coreui/react'
import React, { useState } from 'react'
import { useTranslation } from 'react-i18next'
function ThresholdConfigModal() {
  const { t } = useTranslation('oss/service')
  const [visible, setVisible] = useState(false)
  return (
    <>
      <CButton color="primary" size="sm" onClick={() => setVisible(true)}>
        {t('ThresholdConfigModal.configurationComparedtoThresholdText')}
      </CButton>
      <CModal
        visible={visible}
        alignment="center"
        onClose={() => setVisible(false)}
        aria-labelledby="LiveDemoExampleLabel"
      >
        <CModalHeader>
          <CModalTitle>
            {t('ThresholdConfigModal.configurationComparedtoThresholdText')}
          </CModalTitle>
        </CModalHeader>

        <CModalBody className="w-[500px] text-sm">
          <CForm>
            <CFormLabel htmlFor="basic-url">
              {t('ThresholdConfigModal.averageResponseThresholdLabel')}
            </CFormLabel>
            <CInputGroup className="mb-3">
              <CFormInput
                placeholder={t('ThresholdConfigModal.averageResponseThresholdPlaceHolder')}
                aria-describedby="basic-addon2"
                className="text-sm"
              />
              <CInputGroupText id="basic-addon2">%</CInputGroupText>
            </CInputGroup>
            <CFormLabel htmlFor="basic-url">
              {t('ThresholdConfigModal.errorRateThresholdLabel')}
            </CFormLabel>
            <CInputGroup className="mb-3">
              <CFormInput
                placeholder={t('ThresholdConfigModal.errorRateThresholdPlaceHolder')}
                aria-describedby="basic-addon2"
                className="text-sm"
              />
              <CInputGroupText id="basic-addon2">%</CInputGroupText>
            </CInputGroup>
            <CFormLabel htmlFor="basic-url">
              {t('ThresholdConfigModal.requestsThresholdLabel')}
            </CFormLabel>
            <CInputGroup className="mb-3">
              <CFormInput
                placeholder={t('ThresholdConfigModal.requestsThresholdPlaceHolder')}
                aria-describedby="basic-addon2"
                className="text-sm"
              />
              <CInputGroupText id="basic-addon2">%</CInputGroupText>
            </CInputGroup>
            <CFormLabel htmlFor="basic-url">
              {t('ThresholdConfigModal.logFaultCountThresholdLabel')}
            </CFormLabel>
            <CInputGroup className="mb-3">
              <CFormInput
                placeholder={t('ThresholdConfigModal.logFaultCountThresholdPlaceHolder')}
                aria-describedby="basic-addon2"
                className="text-sm"
              />
              <CInputGroupText id="basic-addon2">%</CInputGroupText>
            </CInputGroup>
          </CForm>
        </CModalBody>
        <CModalFooter>
          <CButton color="primary">{t('ThresholdConfigModal.saveText')}</CButton>
        </CModalFooter>
      </CModal>
    </>
  )
}

export default ThresholdConfigModal
