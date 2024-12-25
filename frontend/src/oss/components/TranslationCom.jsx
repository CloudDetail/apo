import React from 'react'
import { useTranslation } from 'react-i18next'

function TranslationCom({ text, url }) {
  const { t } = useTranslation(url)
  return t(text)
}

export default TranslationCom
