import React from 'react'

interface IconWrapperProps {
  icon: React.ReactNode
  size?: string | number
}

const IconWrapper: React.FC<IconWrapperProps> = ({ icon, size = 'auto' }) => {
  return (
    <div
      className="rounded-md bg-[var(--ant-color-bg-layout)] overflow-hidden"
      style={{ width: size, height: size }}
    >
      {icon}
    </div>
  )
}

export default IconWrapper
