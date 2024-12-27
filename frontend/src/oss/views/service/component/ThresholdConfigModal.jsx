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
        {t('thresholdConfigModal.configurationComparedtoThresholdText')}
      </CButton>
      <CModal
        visible={visible}
        alignment="center"
        onClose={() => setVisible(false)}
        aria-labelledby="LiveDemoExampleLabel"
      >
        <CModalHeader>
          <CModalTitle>
            {t('thresholdConfigModal.configurationComparedtoThresholdText')}
          </CModalTitle>
        </CModalHeader>

        <CModalBody className="w-[500px] text-sm">
          <CForm>
            <CFormLabel htmlFor="basic-url">
              {t('thresholdConfigModal.averageResponseThresholdLabel')}
            </CFormLabel>
            <CInputGroup className="mb-3">
              <CFormInput
                placeholder={t('thresholdConfigModal.averageResponseThresholdPlaceHolder')}
                aria-describedby="basic-addon2"
                className="text-sm"
              />
              <CInputGroupText id="basic-addon2">%</CInputGroupText>
            </CInputGroup>
            <CFormLabel htmlFor="basic-url">
              {t('thresholdConfigModal.errorRateThresholdLabel')}
            </CFormLabel>
            <CInputGroup className="mb-3">
              <CFormInput
                placeholder={t('thresholdConfigModal.errorRateThresholdPlaceHolder')}
                aria-describedby="basic-addon2"
                className="text-sm"
              />
              <CInputGroupText id="basic-addon2">%</CInputGroupText>
            </CInputGroup>
            <CFormLabel htmlFor="basic-url">
              {t('thresholdConfigModal.requestsThresholdLabel')}
            </CFormLabel>
            <CInputGroup className="mb-3">
              <CFormInput
                placeholder={t('thresholdConfigModal.requestsThresholdPlaceHolder')}
                aria-describedby="basic-addon2"
                className="text-sm"
              />
              <CInputGroupText id="basic-addon2">%</CInputGroupText>
            </CInputGroup>
            <CFormLabel htmlFor="basic-url">
              {t('thresholdConfigModal.logFaultCountThresholdLabel')}
            </CFormLabel>
            <CInputGroup className="mb-3">
              <CFormInput
                placeholder={t('thresholdConfigModal.logFaultCountThresholdPlaceHolder')}
                aria-describedby="basic-addon2"
                className="text-sm"
              />
              <CInputGroupText id="basic-addon2">%</CInputGroupText>
            </CInputGroup>
          </CForm>
        </CModalBody>
        <CModalFooter>
          <CButton color="primary">{t('thresholdConfigModal.saveText')}</CButton>
        </CModalFooter>
      </CModal>
    </>
  )
}

export default ThresholdConfigModal
