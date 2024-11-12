import React from 'react'

export default function Tag(props) {
  const { type = 'default', color, children } = props
  const typeColorMap = {
    default: { color: '#a1a1a1', borderColor: '#fafafa' },
    success: { color: '#6abe39', borderColor: '#274916', backgroundColor: '#162312' },
    error: { color: '#e84749', borderColor: '#58181c', backgroundColor: '#2a1215' },
    warning: { color: '#e89a3c', background: '#2b1d11', borderColor: '#593815' },
    primary: {
      color: '#3c89e8',
      backgroundColor: '#111a2c',
      borderColor: '#15325b',
    },
  }
  return (
    <span
      style={{
        ...typeColorMap[type],
        border: '1px solid',
        borderColor: color ?? typeColorMap[type],
        padding: '2px 5px',
        borderRadius: 2,
        fontSize: 10,
      }}
    >
      {children}
    </span>
  )
}
