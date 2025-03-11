/**
 * Copyright 2025 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import i18next from 'i18next'
import { useEffect, useRef } from 'react'

const WorkflowsIframe = ({ src }) => {
  const language = i18next.language
  const workflowRef = useRef(null)

  useEffect(() => {
    let interval: any = null 

    const sendMessageToB = () => {
      if (workflowRef.current) {
        workflowRef.current.contentWindow.postMessage(
          {
            action: 'auto-login',
            data: {
              account: {
                email: 'admin@admin.com',
                password: 'APO2024@admin',
                // email: 'test@163.com',
                // password: 'test123456',
                language: language,
                remember_me: true,
              },
              src: src.slice(5),
            },
          },
          '*',
        )
      }
    }

    const handleMessage = (event: MessageEvent) => {
      if (event.data === 'got' && interval) {
        clearInterval(interval)
        interval = null
      }
    }

    interval = setInterval(sendMessageToB, 1000)

    window.addEventListener('message', handleMessage)

    return () => {
      if (interval) {
        clearInterval(interval)
      }
      window.removeEventListener('message', handleMessage)
    }
  }, [])

  return <iframe ref={workflowRef} src={src} width="100%" height="100%" frameBorder={0}></iframe>
}
export default WorkflowsIframe
