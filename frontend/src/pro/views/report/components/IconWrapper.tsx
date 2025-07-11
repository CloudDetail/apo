const IconWrapper = ({ icon, size = 'auto' }) => {
  return (
    <div
      className="rounded-md bg-[var(--ant-color-bg-layout)] p-2 "
      style={{ width: size, height: size }}
    >
      {icon}
    </div>
  )
}
export default IconWrapper
