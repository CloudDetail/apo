import React, { Suspense, useEffect } from 'react'
import { HashRouter, Route, Routes } from 'react-router-dom'
import { useDispatch, useSelector } from 'react-redux'

import { CSpinner, useColorModes } from '@coreui/react'
import './scss/style.scss'
import './index.css'
import { promLanguageDefinition } from 'monaco-promql'
import { getRuleGroupLabelApi } from './api/alerts'

// Containers
const DefaultLayout = React.lazy(() => import('./layout/DefaultLayout'))

// Pages
const Login = React.lazy(() => import('./views/pages/login/Login'))
const Register = React.lazy(() => import('./views/pages/register/Register'))
const Page404 = React.lazy(() => import('./views/pages/page404/Page404'))
const Page500 = React.lazy(() => import('./views/pages/page500/Page500'))
const App = () => {
  const { isColorModeSet, setColorMode } = useColorModes('coreui-free-react-admin-template-theme')
  // const { isColorModeSet, setColorMode } = useColorModes('dark')
  const storedTheme = useSelector((state) => state.theme)
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
    const urlParams = new URLSearchParams(window.location.href.split('?')[1])
    const theme = urlParams.get('theme') && urlParams.get('theme').match(/^[A-Za-z0-9\s]+/)[0]
    setColorMode('dark')
    getRuleGroupLabels()
    // if (theme) {
    //   setColorMode('light')
    // }

    // if (isColorModeSet()) {
    //   return
    // }
    getMonacoPromqlConfig()
    // setColorMode(storedTheme)
  }, []) // eslint-disable-line react-hooks/exhaustive-deps
  return (
    <HashRouter>
      <Suspense
        fallback={
          <div className="pt-3 text-center">
            <CSpinner color="primary" variant="grow" />
          </div>
        }
      >
        <Routes>
          <Route exact path="/login" name="Login Page" element={<Login />} />
          <Route exact path="/register" name="Register Page" element={<Register />} />
          <Route exact path="/404" name="Page 404" element={<Page404 />} />
          <Route exact path="/500" name="Page 500" element={<Page500 />} />
          <Route path="*" name="Home" element={<DefaultLayout />} />
        </Routes>
      </Suspense>
    </HashRouter>
  )
}

export default App
