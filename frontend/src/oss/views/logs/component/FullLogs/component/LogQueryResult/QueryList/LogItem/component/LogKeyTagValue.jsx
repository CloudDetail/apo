/**
 * Copyright 2024 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import { useEffect, useState } from 'react'
import ReactJson from 'react-json-view'
import { Virtuoso } from 'react-virtuoso'
import LogTagDropDown from './LogTagDropdown'
function isJSONString(str) {
  try {
    return (typeof JSON.parse(str) === 'object' && JSON.parse(str) !== null)
  } catch (e) {
    return false
  }
}
const determineTypeAndValue = (description, title) => {
  if (typeof description === 'string') {
    if (isJSONString(description)) {
      return { type: 'object', value: JSON.parse(description) };
    }
    if (description.length < 1000 && title !== 'content') {
      return { type: 'string', value: [description] };
    }
    return { type: 'longString', value: description.split('\n') };
  }
  if (typeof description === 'number' || typeof description === 'boolean') {
    return { type: typeof description, value: [String(description)] };
  }
  if (typeof description === 'object') {
    return { type: 'object', value: [description] };
  }
  return { type: null, value: [null] }; // 默认情况
};

const formatValue = (value) => (typeof value === 'object' ? JSON.stringify(value) : value)

const LogKeyTagValue = ({ title, description }) => {
  const [type, setType] = useState(null)
  const [value, setValue] = useState([null])
  useEffect(() => {
    const { type, value } = determineTypeAndValue(description, title);
    setType(type)
    setValue(value)
  }, [description, title])
  return (
    <div
      className="break-all  cursor-pointer w-full"
      style={{
        whiteSpace: 'break-spaces',
        wordBreak: 'break-all',
        overflow: 'hidden',
      }}
    >
      {type === 'object' ? (
        <div
          onClick={(e) => e.stopPropagation()} // 阻止事件冒泡
          style={{ width: '100%' }}
        >
          <ReactJson
            collapsed={1}
            src={value[0]}
            theme="brewer"
            displayDataTypes={false}
            style={{ width: '100%' }}
            enableClipboard={true}
            name={false}
          />
        </div>
      ) : type === 'longString' ? (
        <pre
          className=" text-gray-300  h-full w-full overflow-hidden bg-[#0d0d0e] text-xs p-2 leading-relaxed"
          style={{
            whiteSpace: 'break-spaces',
            wordBreak: 'break-all',
            overflow: 'hidden',
            marginBottom: 3,
          }}
        >
          {value?.length > 10 ? (
            <Virtuoso
              style={{ height: 400, width: '100%' }}
              overscan={1000}
              data={value}
              itemContent={(index, paragraph) => <div key={index}>{value[index]}</div>}
            />
          ) : (
            <>{value.join('\n')}</>
          )}
        </pre>
      ) : (
        <LogTagDropDown
          objKey={formatValue(title)}
          value={formatValue(description)}
          children={<div className="hover:underline">{value[0]}</div>}
        />
      )}
    </div>
  )
}
export default LogKeyTagValue
