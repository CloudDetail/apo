import { CImage } from '@coreui/react'
import React from 'react'
import emptyImg from 'src/core/assets/images/empty.svg'
function Empty({ context = '暂无数据' }) {
  return (
    <div className="w-full h-full flex flex-col justify-center items-center py-4">
      <CImage src={emptyImg} width={100} />
      {context}
    </div>
  )
}
export default Empty
