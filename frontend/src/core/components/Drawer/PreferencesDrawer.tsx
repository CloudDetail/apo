/**
 * Copyright 2025 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import { Button, Drawer, Segmented } from "antd"
import { IoLanguageOutline } from "react-icons/io5"
import { VscColorMode } from "react-icons/vsc"
import { SunFilled, MoonFilled, CloseOutlined, SettingOutlined, SkinOutlined } from '@ant-design/icons'
import { useTranslation } from "react-i18next";
import { useDispatch, useSelector } from "react-redux";
import { useColorModes } from "@coreui/react";
import i18next from "i18next";

const PreferencesDrawer = ({ open, onClose }) => {
  const dispatch = useDispatch()
  const { theme } = useSelector((state) => state.settingReducer)
  const { t, i18n } = useTranslation('core/userToolBox')

  const { setColorMode } = useColorModes('coreui-free-react-admin-template-theme')

  const toggleTheme = (value: 'light' | 'dark') => {
    setColorMode(value)
    dispatch({ type: 'setTheme', payload: value })
  }

  const toggleLanguage = (value: 'zh' | 'en') => {
    i18next
      .changeLanguage(value)
      .then(() => {
        dispatch({ type: 'setLanguage', payload: value })
      })
  }

  const PreferenceOption = ({ icon, title, children }) => {
    return (
      <div className="w-full flex flex-col justify-center items-start gap-2">
        <div className="flex items-center justify-center gap-2">
          {icon}
          <p className="text-base">{title}</p>
        </div>
        {children}
      </div>
    );
  };

  return (
    <Drawer
      title={
        <div className="flex justify-start items-center gap-2">
          <SkinOutlined />
          <p className="p-1">{t('preferences')}</p>
        </div>
      }
      placement='right'
      size='default'
      closable={false}
      onClose={onClose}
      open={open}
      extra={
        <Button type='text' icon={<CloseOutlined />} onClick={onClose}></Button>
      }
    >
      <div className="flex flex-col justify-center items-start gap-8">
        <PreferenceOption
          icon={<VscColorMode className="text-lg" title={t("colorMode")} />}
          title={t("colorMode")}
        >
          <Segmented
            defaultValue={theme}
            onChange={(value) => toggleTheme(value)}
            size="middle"
            shape="round"
            options={[
              { label: t('darkMode'), value: 'dark', icon: <MoonFilled /> },
              { label: t('lightMode'), value: 'light', icon: <SunFilled /> },
            ]}
          />
        </PreferenceOption>
        <PreferenceOption
          icon={<IoLanguageOutline className="text-lg" title={t("language")} />}
          title={t("language")}
        >
          <Segmented
            defaultValue={i18n.language}
            onChange={(value) => toggleLanguage(value)}
            size="middle"
            shape="round"
            options={[
              { label: '中文', value: 'zh' },
              { label: 'English', value: 'en' },
            ]}
          />
        </PreferenceOption>
      </div>
    </Drawer>
  )
}

export default PreferencesDrawer