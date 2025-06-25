import { Result } from 'antd'
import { useTranslation } from 'react-i18next'

const NoTraceHint = () => {
  const { t } = useTranslation('pro/report')
  return <Result title={t('noTrace')} />
}
export default NoTraceHint
