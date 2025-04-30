import LogsPng from 'src/pro/assets/dataViewTypeExample/zh/logs.png'
import LogsPngEn from 'src/pro/assets/dataViewTypeExample/en/logs.png'
import { Image } from 'antd'
import i18next, { t } from 'i18next'
const ErrorDataTypeExample = () => {
  return (
    <div className="text-center">
      <div>{t(`pro/chatAI:causeConfirm.faultRecommendationDescription`)}</div>
      <Image src={i18next.language === 'en' ? LogsPngEn : LogsPng} width={400} />
    </div>
  )
}
export default ErrorDataTypeExample
