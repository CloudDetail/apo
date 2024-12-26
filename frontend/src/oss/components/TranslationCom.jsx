import React from 'react'
import { useTranslation } from 'react-i18next'

function TranslationCom({ text, space }) {
  const { t } = useTranslation(space)
  return t(text)
}

export default TranslationCom
