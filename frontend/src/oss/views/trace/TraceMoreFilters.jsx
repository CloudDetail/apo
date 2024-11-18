import { CCol, CCollapse, CRow } from '@coreui/react'
import React, { useEffect, useState } from 'react'
import { Button, Checkbox, Input, InputNumber, Tag } from 'antd'
import TraceErrorType from './component/TraceErrorType'

function TraceMoreFilters(props) {
  const { visible, confirmFIlter } = props
  const [inputNamespace, setInputNamespace] = useState(null)

  const [sliderRange, setSliderRange] = useState([0, 0])
  const [isSlow, setIsSlow] = useState(true)
  const [isError, setIsError] = useState(false)
  const [faultTypeList, setFaultTypeList] = useState([])
  const [filters, setFilters] = useState([])
  const options = [
    { label: <TraceErrorType type="slow" />, value: 'slow' },
    { label: <TraceErrorType type="error" />, value: 'error' },
    { label: <TraceErrorType type="normal" />, value: 'normal' },
  ]
  // const changeSlider = (props) => {
  //   setSliderRange(props)
  // }
  const confirmQuery = () => {
    confirmFIlter({
      namespace: inputNamespace,
      duration: sliderRange[1] > 0 ? sliderRange : null,
      isSlow,
      isError,
      faultTypeList,
    })
  }
  useEffect(() => {
    setInputNamespace(props.namespace)
    if (props.duration && props.duration[1] > props.duration[0]) {
      setSliderRange(props.duration)
    }
    if (props.faultTypeList) {
      setFaultTypeList(props.faultTypeList)
    }
    setIsSlow(props.isSlow)
    setIsError(props.isError)
  }, [visible])
  useEffect(() => {
    let filters = []
    if (sliderRange && sliderRange[1] > sliderRange[0]) {
      filters.push({
        key: 'sliderRange',
        label: (
          <span className="flex flex-row items-center">
            {' '}
            持续时间范围 ： [{sliderRange[0]}ms, {sliderRange[1]}ms]
          </span>
        ),
        onClear: () => {
          setSliderRange([0, 0])
        },
      })
    }

    if (inputNamespace) {
      filters.push({
        key: 'sliderRange',
        label: `命名空间 = ${inputNamespace}`,
        onClear: () => setInputNamespace(''),
      })
    }
    if (faultTypeList?.length > 0) {
      const onClear = () => setFaultTypeList([])
      if (faultTypeList.length === 3) {
        filters.push({
          key: 'sliderRange',
          label: (
            <span className="flex flex-row items-center">
              故障类型包含其中任意： {options.map((item) => item.label)}
            </span>
          ),
          onClear: onClear,
        })
      } else if (faultTypeList.includes('normal')) {
        filters.push({
          key: 'sliderRange',
          label: (
            <span className="flex flex-row items-center">
              故障类型为：{' '}
              {faultTypeList.map((item, index) => (
                <>
                  {index > 0 && <span className="px-2">或</span>}
                  <TraceErrorType type={item} />
                </>
              ))}
            </span>
          ),
          onClear: onClear,
        })
      } else {
        filters.push({
          key: 'sliderRange',
          label: (
            <span className="flex flex-row items-center">
              故障类型<span>{faultTypeList.length === 2 ? '为' : '包含'}</span>：
              {faultTypeList.map((item) => (
                <TraceErrorType type={item} />
              ))}{' '}
            </span>
          ),
          onClear: onClear,
        })
      }
    }
    setFilters(filters)
  }, [inputNamespace, sliderRange, faultTypeList])
  return (
    <CCollapse visible={visible}>
      <div className="p-3 text-xs ">
        <CRow xs={{ cols: 3 }} className="gap-y-3 w-full">
          {/* Duration */}
          <CCol>
            <h2 className="text-sm font-bold mb-2">持续时间</h2>
            <div className="flex flex-row ">
              <div className="pr-2">
                <InputNumber
                  addonBefore="MIN"
                  addonAfter="ms"
                  min={0}
                  value={sliderRange[0]}
                  status={sliderRange[0] > sliderRange[1] && 'error'}
                  onChange={(value) => setSliderRange([value, sliderRange[1]])}
                />
              </div>
              <div className="pl-2">
                <InputNumber
                  addonBefore="MAX"
                  addonAfter="ms"
                  min={0}
                  value={sliderRange[1]}
                  onChange={(value) => setSliderRange([sliderRange[0], value])}
                />
              </div>
            </div>
            {/* <div className="px-3 pt-3">
              <Slider range max={1000} value={sliderRange} onChange={changeSlider} />
            </div> */}
          </CCol>

          <CCol>
            <h2 className="text-sm font-bold mb-2">命名空间</h2>

            <Input
              value={inputNamespace}
              placeholder="检索"
              onChange={(event) => setInputNamespace(event.target.value)}
            />
          </CCol>
          <CCol className="text-base">
            <h2 className="text-sm font-bold mb-2">故障类型</h2>
            <Checkbox.Group
              onChange={setFaultTypeList}
              options={options}
              value={faultTypeList}
            ></Checkbox.Group>
          </CCol>

          {/*
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
            </CCol> */}
        </CRow>
        <div className="flex items-center justify-between flex-row  py-2">
          <div className="flex-1 flex flex-row  text-neutral-400 items-stretch flex-nowrap">
            <div className="flex items-center flex-grow-0 flex-shrink-0">当前筛选器：</div>
            <div className="flex flex-row flex-grow flex-shrink flex-wrap">
              {filters.map((filter) => (
                <Tag
                  bordered={false}
                  closable
                  className="flex items-center"
                  onClose={() => filter.onClear()}
                >
                  {filter.label}
                </Tag>
                // <div className="px-2 rounded-xl py-1 text-xs border text-[#9ca3af] border-[#9ca3af] flex items-center">
                //   {filter}
                // </div>
              ))}
            </div>
          </div>
          <Button type="primary" onClick={confirmQuery}>
            执行筛选
          </Button>
        </div>
      </div>
    </CCollapse>
  )
}

export default TraceMoreFilters
