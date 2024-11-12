import { useEffect, useState } from 'react'
import ReactJson from 'react-json-view'
import { Virtuoso } from 'react-virtuoso'
import LogTagDropDown from './LogTagDropdown'
function isJSONString(str) {
  try {
    JSON.parse(str)
    return true
  } catch (e) {
    return false
  }
}
const formatValue = (value) => (typeof value === 'object' ? JSON.stringify(value) : value)

const LogKeyTagValue = ({ title, description }) => {
  const [type, setType] = useState(null)
  const [value, setValue] = useState([null])
  useEffect(() => {
    if (typeof description === 'string') {
      if (isJSONString(description)) {
        setType('object')
        setValue(JSON.parse(description))
      } else {
        if (description?.length < 1000 && title !== 'content') {
          setType('string')
          setValue([description])
        } else {
          setType('longString')
          setValue(description?.split('\n'))
        }
      }
    } else if (typeof description === 'object') {
      // 非字符串类型，直接显示
      setType('object')
      setValue([description])
    }
  }, [description])
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
          className=" text-gray-300  h-full w-full overflow-hidden bg-[#0d0d0e] text-xs p-2"
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
