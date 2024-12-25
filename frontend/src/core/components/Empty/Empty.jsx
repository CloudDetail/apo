import { CImage } from '@coreui/react'
import React, { useEffect, useState } from 'react'
import emptyImg from 'src/core/assets/images/empty.svg'
import { useTranslation } from 'react-i18next'
function Empty({ propContext }) {
  const [context, setContext] = useState('No data')
  const { i18n } = useTranslation()
  useEffect(() => {
    if (propContext) {
      setContext(propContext)
      return
    }
    setContext(i18n.language === 'en' ? 'No data' : '暂无数据')
  }, [i18n.language])
  return (
    <div className="w-full h-full flex flex-col justify-center items-center py-4">
      <CImage src={emptyImg} width={100} />
      {context}
    </div>
  )
}
export default Empty
