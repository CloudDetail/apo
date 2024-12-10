import { Select } from 'antd'
import React, { useEffect, useState, useCallback } from 'react'

const CustomSelect = React.memo((props) => {
  const { options, value, onChange, defaultValue, isClearable = false } = props
  const [standardOptions, setStandardOptions] = useState([])
  const [selectedValue, setSelectedValue] = useState(null)

  const darkThemeStyles = {
    indicatorSeparator: () => ({ display: 'none' }),
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
      height: 25,
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
      width: 'max-content',
      minWidth: '100%',
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

  useEffect(() => {
    const updatedOptions = options.map((option) => ({
      value: option,
      label: option,
    }))
    setStandardOptions(updatedOptions)
  }, [options])

  useEffect(() => {
    if (value) {
      const item = standardOptions.find((item) => item.value === value)

      setSelectedValue(item)
    } else {
      setSelectedValue(null)
    }
  }, [value, standardOptions])

  const handleChange = (item) => {
    onChange(item ?? '')
  }

  return (
    // <Select
    //   options={standardOptions}
    //   value={selectedValue}
    //   className="w-full"
    //   styles={darkThemeStyles}
    //   onChange={handleChange}
    //   defaultValue={defaultValue}
    //   isClearable={isClearable}
    //   placeholder={''}
    // />
    <Select
      options={standardOptions}
      value={selectedValue}
      onChange={handleChange}
      defaultValue={defaultValue}
      allowClear={isClearable}
      popupMatchSelectWidth={false}
      className="w-full"
      placeholder="请选择"
      showSearch
    />
  )
})

export default CustomSelect
