/**
 * Copyright 2025 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */
import { useColorModes } from '@coreui/react'
import { Button } from 'antd'
import { MdOutlineDarkMode, MdOutlineLightMode } from 'react-icons/md'
import { useDispatch, useSelector } from 'react-redux'

const ThemeSwitcher = () => {
  const { setColorMode } = useColorModes('coreui-free-react-admin-template-theme')
  const { theme } = useSelector((state) => state.settingReducer)
  const dispatch = useDispatch()
  const changeTheme = (theme: 'light' | 'dark') => {
    setColorMode(theme)
    dispatch({ type: 'setTheme', payload: theme })
  }
  return (
    <>
      {theme === 'light' ? (
        <Button
          type="text"
          icon={<MdOutlineDarkMode />}
          onClick={() => {
            changeTheme('dark')
          }}
        ></Button>
      ) : (
        <Button
          type="text"
          icon={<MdOutlineLightMode />}
          onClick={() => {
            changeTheme('light')
          }}
        ></Button>
      )}
    </>
  )
}
export default ThemeSwitcher
