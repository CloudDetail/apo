import LogsPng from 'src/pro/assets/dataViewTypeExample/zh/logs.png'
import { Image } from 'antd'
import { t } from 'i18next'
const ErrorDataTypeExample = () => {
  return (
    <div className="text-center">
      <div>{t(`pro/chatAI:causeConfirm.faultRecommendationDescription`)}</div>
      <Image src={LogsPng} width={400} />
    </div>
  )
}
export default ErrorDataTypeExample
