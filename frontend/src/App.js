/**
 * Copyright 2024 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import React, { Suspense, useEffect } from 'react'
import { HashRouter, Route, Routes } from 'react-router-dom'
import { useDispatch, useSelector } from 'react-redux'

import { useColorModes } from '@coreui/react'
import './index.css'
import 'src/core/scss/style.scss'
import { promLanguageDefinition } from 'monaco-promql'
import { getRuleGroupLabelApi } from 'src/core/api/alerts'
import { Spin } from 'antd'
import { useTranslation } from 'react-i18next'
// Containers
const DefaultLayout = React.lazy(() => import('src/core/layout/DefaultLayout'))
const Login = React.lazy(() => import('./core/views/Login/Login'))

// // Pages
// const Login = React.lazy(() => import('./community/1/pages/login/Login'))
// const Register = React.lazy(() => import('./community/1/pages/register/Register'))
// const Page404 = React.lazy(() => import('./community/1/pages/page404/Page404'))
// const Page500 = React.lazy(() => import('./community/1/pages/page500/Page500'))
const App = () => {
  const { isColorModeSet, setColorMode } = useColorModes('coreui-free-react-admin-template-theme')
  // const { isColorModeSet, setColorMode } = useColorModes('dark')
  const { theme } = useSelector((state) => state.settingReducer)
  const { t, i18n } = useTranslation()
  const dispatch = useDispatch()
  const setGroupLabel = (value) => {
    dispatch({ type: 'setGroupLabel', payload: value })
  }
  const setMonacoPromqlConfig = (value) => {
    dispatch({ type: 'setMonacoPromqlConfig', payload: value })
  }
  const getRuleGroupLabels = () => {
    getRuleGroupLabelApi().then((res) => {
      setGroupLabel(res?.groupsLabel ?? [])
    })
  }
  const getMonacoPromqlConfig = () => {
    promLanguageDefinition
      .loader()
      .then((mod) => {
        setMonacoPromqlConfig(mod)
      })
      .catch((err) => {
        console.error('Error loading PromQL module:', err)
      })
  }
  useEffect(() => {
    setColorMode(theme)
  }, [theme])
  useEffect(() => {
    document.title = t('apoTitle');
  }, [i18n.language, t]);
  useEffect(() => {
    if (window.location.hash !== '#/login') {
      getRuleGroupLabels()
    }
    getMonacoPromqlConfig()
  }, [])

  return (
    <HashRouter>
      <Suspense
        fallback={
          <div className="pt-3 text-center">
            <Spin />
          </div>
        }
      >
        <Routes>
          <Route exact path="/login" name="Login Page" element={<Login />} />
          <Route path="*" name="Home" element={<DefaultLayout />} />
        </Routes>
      </Suspense>
    </HashRouter>
  )
}

export default App
