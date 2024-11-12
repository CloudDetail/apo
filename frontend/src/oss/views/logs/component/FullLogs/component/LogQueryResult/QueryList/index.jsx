import React from 'react'
import { Virtuoso } from 'react-virtuoso'
import LogItem from './LogItem'
const QueryList = ({ logs, openContextModal = null }) => {
  return (
    <Virtuoso
      style={{ height: '100%', width: '100%' }}
      data={logs}
      itemContent={(index) => (
        <div style={{ padding: '10px' }}>
          <LogItem log={logs[index]} openContextModal={openContextModal} />
        </div>
      )}
    />
  )
}

export default QueryList
