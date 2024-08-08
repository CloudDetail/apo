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
function ThresholdCofigModal() {
  const [visible, setVisible] = useState(false)
  return (
    <>
      <CButton color="primary" size="sm" onClick={() => setVisible(true)}>
        配置同比阈值
      </CButton>
      <CModal
        visible={visible}
        alignment="center"
        onClose={() => setVisible(false)}
        aria-labelledby="LiveDemoExampleLabel"
      >
        <CModalHeader>
          <CModalTitle >配置同比阈值</CModalTitle>
        </CModalHeader>

        <CModalBody className="w-[500px] text-sm">
          <CForm>
          <CFormLabel htmlFor="basic-url">平均响应时间同比阈值</CFormLabel>
            <CInputGroup className="mb-3">
              <CFormInput
                placeholder="输入平均响应时间 类同比阈值"
                aria-describedby="basic-addon2"
                className="text-sm"
              />
              <CInputGroupText id="basic-addon2">%</CInputGroupText>
            </CInputGroup>
            <CFormLabel htmlFor="basic-url">错误率同比阈值</CFormLabel>
            <CInputGroup className="mb-3">
              <CFormInput
                placeholder="输入错误率同比阈值"
                aria-describedby="basic-addon2"
                className="text-sm"
              />
              <CInputGroupText id="basic-addon2">%</CInputGroupText>
            </CInputGroup>
            <CFormLabel htmlFor="basic-url">请求次数同比阈值</CFormLabel>
            <CInputGroup className="mb-3">
              <CFormInput
                placeholder="输入请求次数同比阈值"
                aria-describedby="basic-addon2"
                className="text-sm"
              />
              <CInputGroupText id="basic-addon2">%</CInputGroupText>
            </CInputGroup>
            <CFormLabel htmlFor="basic-url">日志错误数量同比阈值</CFormLabel>
            <CInputGroup className="mb-3">
              <CFormInput
                placeholder="输入日志错误数量同比阈值"
                aria-describedby="basic-addon2"
                className="text-sm"
              />
              <CInputGroupText id="basic-addon2">%</CInputGroupText>
            </CInputGroup>
          </CForm>
        </CModalBody>
        <CModalFooter>
          <CButton color="primary">保存</CButton>
        </CModalFooter>
      </CModal>
    </>
  )
}

export default ThresholdCofigModal
