import { useLogsContext } from 'src/core/contexts/LogsContext'
import LogValueTag from './LogValueTag'
import LogKeyTag from './LogKeyTag'

const LogItemFold = ({ tags, fields }) => {
  const { tableInfo } = useLogsContext()

  return (
    <>
      {/* 渲染 tags */}
      <div className="text-ellipsis text-wrap flex" style={{ display: '-webkit-box' }}>
        {Object.entries(tags).map(([key, value]) => {
          if (
            value !== '' && // 确保 value 存在且非空
            key !== (tableInfo?.timeField || 'timestamp') && // 排除与 timeField 相同的键
            typeof value !== 'object' // 确保 value 不是对象
          ) {
            return <LogValueTag key={key} objKey={key} value={String(value)} />
          }
          return null // 不符合条件时返回 null
        })}
      </div>

      {/* 渲染 fields */}
      <div className="text-ellipsis text-wrap flex flex-col overflow-hidden">
        {fields
          ? Object.entries(fields).map(([key, value]) => (
              <div key={key}>
                <LogKeyTag key={key} title={key} description={value} />
              </div>
            ))
          : null}
      </div>
    </>
  )
}

export default LogItemFold
