/**
 * Copyright 2025 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import i18next from 'i18next'
import { useEffect, useRef } from 'react'
import { workflowLoginApi } from 'src/core/api/workflows'

const WorkflowsIframe = ({ src }) => {
  const language = i18next.language
  const workflowRef = useRef(null)
  const intervalRef = useRef<any>(null)
  useEffect(() => {
    const sendMessageToB = (data) => {
      if (workflowRef.current) {
        workflowRef.current.contentWindow.postMessage(data, '*')
      }
    }

    const handleMessage = (event: MessageEvent) => {
      if (event.data === 'got' && intervalRef.current) {
        clearInterval(intervalRef.current)
        intervalRef.current = null
      }
    }

    if (localStorage.getItem('difyToken') && localStorage.getItem('difyRefreshToken')) {
      const data = {
        action: 'auto-login',
        data: {
          token: localStorage.getItem('difyToken'),
          refreshToken: localStorage.getItem('difyRefreshToken'),
        },
        src: src.slice(5),
      }
      intervalRef.current = setInterval(() => sendMessageToB(data), 1000)
      window.addEventListener('message', handleMessage)
    } else {
      console.error('token not found')
    }

    return () => {
      if (intervalRef.current) {
        clearInterval(intervalRef.current)
      }
      window.removeEventListener('message', handleMessage)
    }
  }, [])

  return (
    <iframe
      ref={workflowRef}
      src={src}
      width="100%"
      height="100%"
      frameBorder={0}
    ></iframe>
  )
}
export default WorkflowsIframe
