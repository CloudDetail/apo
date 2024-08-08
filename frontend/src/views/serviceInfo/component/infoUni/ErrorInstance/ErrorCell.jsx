import React, { useEffect, useState } from 'react'
import Select from 'react-select'
import Empty from 'src/components/Empty/Empty'
import { convertTime } from 'src/utils/time'

function ErrorCell(props) {
  const { data, update } = props
  const [options, setOptions] = useState([])
  const [selectTraceError, setSelectTraceError] = useState()
  const [menuIsOpen, setMenuIsOpen] = useState(false)
  useEffect(() => {
    const options = []
    data.map((item) => {
      item.errors.map((error) => {
        const key = convertTime(item.timestamp, 'yyyy-mm-dd hh:mm:ss') + ' ' + error.type
        options.push({
          value: key,
          label: key,
          customAbbreviation: {
            traceId: item.traceId,
            children: item.children,
            current: item.current,
            parents: item.parents,
            error: error,
            timestamp: item.timestamp,
          },
        })
      })
    })
    setOptions(options)
    setSelectTraceError(options[0])
    update(options[0])
  }, [data])
  const darkThemeStyles = {
    control: (styles) => ({
      ...styles,
      backgroundColor: 'none',
      color: 'white',
      borderWidth: '0.1rem',
      borderColor: 'rgba(255, 255, 255, 0.2)',
      whiteSpace: 'prewrap',
      width: '100%',
      flexWrap: 'nowrap',
      padding: 5,
    }),
    input: (styles) => ({
      ...styles,
      color: 'white',
      border: 0,
    }),
    singleValue: (styles) => ({
      ...styles,
      color: 'white',
    }),
    menu: (styles) => ({
      ...styles,
      backgroundColor: 'var(--cui-dark-bg-subtle)',
    }),
    option: (styles, { isFocused, isSelected }) => ({
      ...styles,
      backgroundColor: isFocused
        ? 'rgba(255, 255, 255, 0.1)'
        : isSelected
          ? 'rgba(255, 255, 255, 0.2)'
          : undefined,
      color: 'white',
    }),
    placeholder: (styles) => ({
      ...styles,
      color: 'gray',
    }),
    multiValue: (provided) => ({
      ...provided,
      display: 'flex',
      whiteSpace: 'normal',
      flexWrap: 'wrap',
      wordBreak: 'break-word',
    }),
    multiValueLabel: (provided) => ({
      ...provided,
      whiteSpace: 'normal',
      wordBreak: 'break-word',
    }),
  }
  const CustomSingleValue = (props) => {
    const { data } = props
    return (
      (data ?? selectTraceError) && (
        <div className="w-full flex-shrink">
          <div className=" overflow-x-hidden whitespace-pre-wrap w-full flex flex-row">
            <div className="text-gray-400 flex-shrink-0">Time：</div>
            <div className="flex-1 w-0 whitespace-nowrap text-wrap break-all">
              {convertTime(
                (data ?? selectTraceError)?.customAbbreviation.timestamp,
                'yyyy-mm-dd hh:mm:ss',
              )}
            </div>
          </div>
          <div className=" overflow-x-hidden  w-full flex flex-row">
            <div className="text-gray-400 flex-shrink-0">ErrorType：</div>
            <div className="flex-1 w-0 whitespace-nowrap text-wrap break-all">
              {(data ?? selectTraceError)?.customAbbreviation.error.type}
            </div>
          </div>
        </div>
      )
    )
  }
  const CustomContainer = (props) => {
    const { data } = props
    return (
      <div className="w-full flex-shrink">
        <div className=" overflow-x-hidden whitespace-pre-wrap w-full flex flex-row">
          <div className="flex-1 w-0 whitespace-nowrap text-wrap break-all">
            {(data ?? selectTraceError) &&
              convertTime(
                (data ?? selectTraceError)?.customAbbreviation.timestamp,
                'yyyy-mm-dd hh:mm:ss',
              )}{' '}
            {''}
            {(data ?? selectTraceError)?.customAbbreviation.error.type}
          </div>
        </div>
      </div>
    )
  }
  const onSelect = (selectTraceError) => {
    setMenuIsOpen(false)
    setSelectTraceError(selectTraceError)
    update(selectTraceError)
  }
  const formatOptionLabel = (props) => {
    const { data } = props
    return (
      <div
        className="overflow-x-hidden w-full text-sm p-1 cursor-pointer hover:bg-[#2a303d] border-b border-gray-700"
        onClick={() => onSelect(data)}
      >
        <CustomSingleValue data={data} />
      </div>
    )
  }
  return options && options.length > 0 ? (
    <div className="w-full h-full">
      <Select
        options={options}
        value={selectTraceError}
        className="w-full"
        styles={darkThemeStyles}
        menuIsOpen={menuIsOpen}
        components={{
          // SingleValue: CustomSingleValue,
          Option: formatOptionLabel,
          ValueContainer: (event) => CustomContainer(event),
        }}
        onMenuOpen={() => setMenuIsOpen(true)}
        onMenuClose={() => setMenuIsOpen(false)}
        menuPortalTarget={document.body}
      ></Select>

      <div className="p-2">
        {selectTraceError && (
          <>
            <span className="text-gray-400 flex-shrink-0">Error Message：</span>
            {selectTraceError?.customAbbreviation?.error?.message}
          </>
        )}
      </div>
    </div>
  ) : (
    <Empty />
  )
}
export default ErrorCell
