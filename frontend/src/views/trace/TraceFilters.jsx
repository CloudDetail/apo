import {
  CButton,
  CCard,
  CCardBody,
  CCol,
  CCollapse,
  CFormCheck,
  CFormInput,
  CFormLabel,
  CFormSelect,
  CInputGroup,
  CInputGroupText,
  CRow,
} from '@coreui/react'
import React, { useState } from 'react'
import { useLocation } from 'react-router-dom'
import { serviceMock } from 'src/components/ReactFlow/mock'
import { FaAnglesDown } from 'react-icons/fa6'
import { BsChevronDoubleDown } from 'react-icons/bs'
import 'rc-slider/assets/index.css'
import Slider from 'rc-slider'
import DateTimeRangePickerCom from 'src/components/DateTime/DateTimeRangePickerCom'
import LogsTraceFilter from 'src/components/Filter/LogsTraceFilter'

function TraceFilters() {
  const [visible, setVisible] = useState(false)
  const location = useLocation()
  const searchParams = new URLSearchParams(location.search)

  const serviceName = searchParams.get('service-name')
  const instanceName = searchParams.get('instance-name')
  const timestamp = searchParams.get('timestamp')
  const [selectServiceName, setSelectServiceName] = useState()
  const [selectInstance, setSelectInstance] = useState()
  const [selectOpreation, setSelectOpreation] = useState()

  const [sliderRange, setSliderRange] = useState([0, 0])

  const changeSlider = (props) => {
    console.log(props)
    setSliderRange(props)
  }
  useState(() => {
    setSelectServiceName(serviceName)
    setSelectInstance(instanceName)
  }, [serviceName, instanceName])
  return (
    <CCard className="mb-3 px-3">
      <LogsTraceFilter type={'trace'} />

      <CCollapse visible={visible}>
        <div className="p-3">
          <CRow xs={{ cols: 3 }} className="gap-y-3">
            {/* Duration */}
            <CCol >
              <h2 className="text-sm font-bold mb-2">Duration</h2>
              <div className="flex flex-row ">
                <div className="pr-2">
                  <CInputGroup size="sm">
                    <CInputGroupText>MIN</CInputGroupText>
                    <CFormInput
                      type="number"
                      value={sliderRange[0]}
                      onChange={(event) => setSliderRange([event.target.value, sliderRange[1]])}
                    ></CFormInput>
                    <CInputGroupText>ms</CInputGroupText>
                  </CInputGroup>
                </div>
                <div className="pl-2">
                  <CInputGroup size="sm">
                    <CInputGroupText>MAX</CInputGroupText>

                    <CFormInput
                      type="number"
                      value={sliderRange[1]}
                      onChange={(event) => setSliderRange([sliderRange[0], event.target.value])}
                    ></CFormInput>
                    <CInputGroupText>ms</CInputGroupText>
                  </CInputGroup>
                </div>
              </div>
              <div className="px-3 pt-3">
                <Slider
                  range
                  max={10000}
                  value={sliderRange}
                  onChange={changeSlider}
                />
              </div>
            </CCol>
            <CCol className="text-base">
              <h2 className="text-sm font-bold mb-2">Status</h2>
              <CFormCheck
                label={
                  <div className="flex flex-row items-center justify-center">
                    <div className="w-1 h-5 rounded bg-[#E5484D] mx-1"></div>Error
                  </div>
                }
              />
              <CFormCheck
                label={
                  <div className="flex flex-row items-center justify-center">
                    <div className="w-1 h-5 rounded bg-[#25e192] mx-1"></div>OK
                  </div>
                }
              />
            </CCol>
            <CCol>
              <h2 className="text-sm font-bold mb-2">Operation / Name</h2>

              <CFormSelect size="sm" value={selectOpreation}></CFormSelect>
            </CCol>
            <CCol>
              <h2 className="text-sm font-bold mb-2">RPC Method</h2>
              <CFormSelect size="sm" value={selectOpreation}></CFormSelect>
            </CCol>
            <CCol>
              <h2 className="text-sm font-bold mb-2">Status Code</h2>
              <CFormSelect size="sm" value={selectOpreation}></CFormSelect>
            </CCol>
            <CCol>
              <h2 className="text-sm font-bold mb-2">HTTP Host</h2>
              <CFormSelect size="sm" value={selectOpreation}></CFormSelect>
            </CCol>
            <CCol>
              <h2 className="text-sm font-bold mb-2">HTTP Method</h2>
              <CFormSelect size="sm" value={selectOpreation}></CFormSelect>
            </CCol>
            <CCol>
              <h2 className="text-sm font-bold mb-2">HTTP Route</h2>
              <CFormSelect size="sm" value={selectOpreation}></CFormSelect>
            </CCol>
            <CCol>
              <h2 className="text-sm font-bold mb-2">HTTP URL</h2>
              <CFormSelect size="sm" value={selectOpreation}></CFormSelect>
            </CCol>
          </CRow>
        </div>
      </CCollapse>
    </CCard>
  )
}

export default TraceFilters
