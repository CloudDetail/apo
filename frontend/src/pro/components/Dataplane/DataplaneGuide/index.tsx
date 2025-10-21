import MarkdownRender from 'src/core/components/Markdown/MarkdownRender'
import { getGuideMdMap } from './guideMd'
import { useTranslation } from 'react-i18next'
import i18n from 'src/i18n'
import { useSelector } from 'react-redux'

const SetupGuide = ({ typeName }: { typeName: string }) => {
  const { t } = useTranslation('pro/agent')
  const currentLanguage = i18n.language
  const language = currentLanguage.startsWith('zh') ? 'zh' : 'en'
  const guideMdMap = getGuideMdMap(language)
  const { theme } = useSelector((state: any) => state.settingReducer)

  return (
    <div className="h-full flex-1  border border-[var(--ant-color-border-tertiary)] rounded-lg py-2 flex flex-col overflow-hidden">
      <div className="text-lg font-bold px-4 my-2">{t('setupGuide', 'Setup Guide')}</div>
      <div className="flex-1 overflow-y-auto px-4">
        <MarkdownRender content={guideMdMap[typeName]} theme={theme} />
      </div>
    </div>
  )
}
export default SetupGuide
