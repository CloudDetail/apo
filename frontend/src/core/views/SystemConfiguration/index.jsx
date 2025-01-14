/**
 * Copyright 2024 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import React from 'react'
import { CCard } from '@coreui/react'
import LanguageSwitcher from 'src/core/components/LanguageSwitcher'
import { IoIosOptions } from 'react-icons/io'
import { useTranslation } from 'react-i18next'

const SystemConfiguration = () => {
  const { t } = useTranslation('core/systemConfiguration')

  return (
    <>
      <CCard className="p-3" style={{ height: 'calc(100vh - 120px)' }}>
        <div className="flex items-center">
          <IoIosOptions size={22} />
          <p className="text-base ml-2">{t('configItem')}</p>
        </div>
        <div className="flex flex-col w-full mt-2">
          <div className="flex flex-col ml-12 mt-2">
            <p className="text-sm mb-1">{t('language')}</p>
            <LanguageSwitcher />
          </div>
        </div>
      </CCard>
    </>
  )
}

export default SystemConfiguration
