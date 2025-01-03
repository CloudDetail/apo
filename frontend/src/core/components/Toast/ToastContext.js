/**
 * Copyright 2024 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

// src/toastContext.js
import React, { createContext, useContext, useState, useCallback, useEffect } from 'react'
import { CToast, CToastBody, CToastClose, CToastHeader, CToaster } from '@coreui/react'
import { setAddToastFunction, showToast } from 'src/core/utils/toast'
import { FaCheck } from 'react-icons/fa'
import { MdInfoOutline } from 'react-icons/md'
import { PiWarningBold } from 'react-icons/pi'
const ToastContext = createContext()

export const ToastProvider = ({ children }) => {
  const [toasts, setToasts] = useState([])

  const addToast = useCallback(({ message = '', title = 'Notification', color = 'success' }) => {
    setToasts((prevToasts) => {
      return [...prevToasts, { message, title, color }]
    })
  }, [])
  
  const removeToast = useCallback(() => {
    setToasts((prevToasts) => prevToasts.slice(1))
  }, [])

  useEffect(() => {
    setAddToastFunction(addToast)
  }, [addToast])
  const vars = {
    '--cui-danger-rgb': '27, 31, 40',
    '--cui-success-rgb': '27, 31, 40',
    '--cui-info-rgb': '27, 31, 40',
  }
  const statusIconMap = {
    success: <FaCheck color="#3abf8c" />,
    danger: <PiWarningBold color="#f6786f" />,
    info: <MdInfoOutline color="1477ff" />,
  }
  return (
    <ToastContext.Provider value={addToast}>
      <CToaster placement="top-right" className="end-0 p-3 pt-5">
        {toasts.map((toast, index) => (
          <CToast
            key={index}
            autohide={true}
            visible={true}
            onClose={removeToast}
            color={toast.color}
            style={vars}
          >
            <CToastBody className="border-t-2 border-b-2 border-r-2 border-l-4 rounded-md relative">
              <div className="flex items-start   pr-7">
                <div className="mx-2 my-1 text-xl">{statusIconMap[toast.color]}</div>
                <div>
                  <div className="text-base">{toast.title}</div>
                  <div>{toast.message}</div>
                </div>
              </div>
              <CToastClose className="me-2 m-auto absolute right-0 top-2" white />
            </CToastBody>
          </CToast>
        ))}
      </CToaster>
      {children}
    </ToastContext.Provider>
  )
}

export const useToast = () => {
  return useContext(ToastContext)
}
