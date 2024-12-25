import React from 'react'
import { useTranslation } from 'react-i18next' // 引入i18n

export default function TraceErrorType({ type }) {
  const { t } = useTranslation('LogsTraceFilter') // 使用i18n
  const ErrorTypeMap = {
    error: {
      name: t('traceErrorType.error'),
      color: 'rgba(220,38,38,.85)',
      background: '#ef444433',
      border: 'rgb(75,85,99)',
    },
    slow: {
      name: t('traceErrorType.slow'),
      color: 'rgba(245,158,11,1)',
      background: '#f59e0b33',
      border: 'rgb(75,85,99)',
    },
    normal: {
      name: t('traceErrorType.normal'),
      color: 'rgba(52,211,153,1)',
      background: '#10b98133',
      border: 'rgb(75,85,99)',
    },
  }

  return (
    <>
      {/* <div
        className="px-2 rounded-xl py-1 text-xs mr-1"
        style={{
          background: ErrorTypeMap[type].background,
          border: ErrorTypeMap[type].border,
          color: ErrorTypeMap[type].color,
        }}
      >
        <span>{ErrorTypeMap[type].name}</span>
      </div> */}
      <div className="px-2 rounded-xl py-1 text-xs border text-[#9ca3af] border-[#9ca3af]">
        {' '}
        <span
          className="rounded-full w-2 h-2 inline-block mr-2"
          style={{
            background: ErrorTypeMap[type].color,
            color: ErrorTypeMap[type].border,
          }}
        ></span>
        <span>{ErrorTypeMap[type].name}</span>
      </div>
    </>
  )
}
