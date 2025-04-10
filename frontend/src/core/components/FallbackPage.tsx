import { Spin, Image } from 'antd'
import ErrorPage from 'src/core/assets/errorPage.svg'
import { t } from 'i18next'

interface FallbackPageProps {
  errorInfo?: string
}

const FallbackPage = ({ errorInfo }: FallbackPageProps) => {
  return (
    <div className="w-screen h-screen flex  items-center justify-center flex-col fixed top-0 left-0">
      <Image src={ErrorPage} width={'30%'} preview={false} />
      <div>
        { errorInfo ? errorInfo : t('common:error') }
      </div>
    </div>
  )
}

export default FallbackPage