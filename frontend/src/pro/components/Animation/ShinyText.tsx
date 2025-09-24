import React from 'react'
import { useSelector } from 'react-redux'

interface ShinyTextProps {
  text: string
  disabled?: boolean
  speed?: number
  className?: string
}

const ShinyText: React.FC<ShinyTextProps> = ({
  text,
  disabled = false,
  speed = 5,
  className = '',
}) => {
  const animationDuration = `${speed}s`

  const commonStyles = {
    backgroundSize: '200% 100%',
    WebkitBackgroundClip: 'text',
    animationDuration: animationDuration,
  }
  const { theme } = useSelector((state) => state.settingReducer)

  return (
    <div
      className={`text-[var(--ant-color-text-tertiary)] bg-clip-text inline-block  ${disabled ? '' : 'animate-shine'} ${className}`}
      style={{
        ...(disabled
          ? {}
          : {
              backgroundImage:
                theme === 'dark'
                  ? 'linear-gradient(120deg, rgba(255, 255, 255, 0) 40%, rgba(255, 255, 255, 0.8) 50%, rgba(255, 255, 255, 0) 60%)'
                  : 'linear-gradient(120deg, rgba(0, 0, 0, 0) 40%, rgba(0, 0, 0, 0.6) 50%, rgba(0, 0, 0, 0) 60%)',
            }),
        ...commonStyles,
      }}
    >
      {text}
    </div>
  )
}

export default React.memo(ShinyText)

// tailwind.config.js
// module.exports = {
//   theme: {
//     extend: {
//       keyframes: {
//         shine: {
//           '0%': { 'background-position': '100%' },
//           '100%': { 'background-position': '-100%' },
//         },
//       },
//       animation: {
//         shine: 'shine 5s linear infinite',
//       },
//     },
//   },
//   plugins: [],
// }
