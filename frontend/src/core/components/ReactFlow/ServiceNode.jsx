import React, { useCallback } from 'react'
import { Handle, Position } from 'reactflow'
import { BiCctv } from 'react-icons/bi'
const handleStyle = { left: 10 }
const ServiceNode = React.memo(({ data, isConnectable }) => {
  const onChange = useCallback((evt) => {}, [])

  return (
    <div className="text-updater-node">
      <Handle
        type="target"
        position={Position.Left}
        isConnectable={isConnectable}
        className="invisible"
      />
      <div
        className="w-[200px] min-h-[60px] p-2 rounded-md border-2 border-solid border-[#6293ff] text-[#6293ff] overflow-hidden"
        style={{ backgroundColor: 'rgba(19, 25, 32, 0.6)' }}
      >
        <div className="absolute top-0 left-0">
          {' '}
          <BiCctv size={35} color={data.isTraced ? '#80ce8dff' : '#a1a1a1'} />
        </div>
        <div className="text-center text-lg pt-2 px-2">
          {data.label}
          <div className="text-xs text-[#9e9e9e] text-left  break-all">{data.endpoint}</div>
        </div>
      </div>
      <Handle
        type="source"
        position={Position.Right}
        id="b"
        isConnectable={isConnectable}
        className="invisible"
      />
    </div>
  )
})

export default ServiceNode
