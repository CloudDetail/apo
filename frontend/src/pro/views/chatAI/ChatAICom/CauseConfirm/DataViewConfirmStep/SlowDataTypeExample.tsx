import PolarisPng from 'src/pro/assets/9-1.png'
import PolarisPngEn from 'src/pro/assets/en/9-1.png'
import { Image } from 'antd'
import i18next, { t } from 'i18next'
const SlowDataTypeExample = () => {
  return (
    <div className="text-center overflow-hidden">
      <div>{t('pro/chatAI:causeConfirm.slowDataDesc')}</div>
      <Image src={i18next.language === 'en' ? PolarisPngEn : PolarisPng} width={200} />
    </div>
  )
}
export default SlowDataTypeExample
