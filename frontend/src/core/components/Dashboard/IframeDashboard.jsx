/**
 * Copyright 2024 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import React, { useEffect, useRef, useState } from 'react'
import { useSelector } from 'react-redux'
import { selectProcessedTimeRange, timeRangeList } from 'src/core/store/reducers/timeRangeReducer'

const isGrafanaUrl = (url) => Boolean(url?.includes('grafana/'))

const addEmbedCleanParam = (url) => {
  if (!url || url.includes('embedClean=') || !isGrafanaUrl(url)) {
    return url
  }

  const connector = url.includes('?') ? '&' : '?'
  return `${url}${connector}embedClean=1`
}

const injectEmbedCleanStyle = (iframe, observerRef) => {
  if (!iframe || !isGrafanaUrl(iframe.src)) {
    return
  }

  try {
    observerRef.current?.disconnect()
    observerRef.current = null

    const iframeDocument = iframe.contentDocument || iframe.contentWindow?.document
    if (!iframeDocument || iframeDocument.querySelector('style[data-grafana-embed-clean]')) {
      return
    }

    const closeMegaMenuIfOpen = () => {
      const toggle = iframeDocument.querySelector('[data-testid="data-testid Toggle menu"]')
      if (/关闭菜单|Close menu/i.test(toggle?.getAttribute('aria-label') || '')) {
        toggle.click()
      }
    }

    const style = iframeDocument.createElement('style')
    style.setAttribute('data-grafana-embed-clean', 'true')
    style.textContent = `
      body.page-dashboard .main-view > div:has([data-testid="data-testid Nav toolbar"]) > div:first-child {
        display: none !important;
      }

      body.page-dashboard [data-testid="data-testid Nav toolbar"] {
        display: none !important;
      }

      body.page-dashboard main {
        padding-top: 0 !important;
      }

      body.page-dashboard [data-testid="data-testid navigation mega-menu"],
      body.page-dashboard [class*="css-4sumg1"] {
        display: none !important;
        visibility: hidden !important;
        pointer-events: none !important;
      }
    `
    iframeDocument.head.appendChild(style)

    closeMegaMenuIfOpen()

    const observer = new iframe.contentWindow.MutationObserver(closeMegaMenuIfOpen)
    observer.observe(iframeDocument.documentElement, {
      childList: true,
      subtree: true,
      attributes: true,
      attributeFilter: ['aria-label'],
    })
    observerRef.current = observer
  } catch (error) {
    console.warn('Unable to apply Grafana iframe embed clean styles.', error)
  }
}

//receiving 'dashboardKey' as a prop ..
function IframeDashboard({ dashboardKey, src, srcProp }) {
  const [iframeSrc, setIframeSrc] = useState(addEmbedCleanParam(srcProp || src))
  const iframeRef = useRef(null)
  const observerRef = useRef(null)
  const storeTimeRange = useSelector((state) => state.timeRange)
  const { startTime, endTime } = useSelector(selectProcessedTimeRange)
  const menuItems = useSelector((state) => state.userReducer.menuItems)
  const { theme } = useSelector((state) => state.settingReducer)

  // Update src when srcProp changes
  useEffect(() => {
    setIframeSrc(addEmbedCleanParam(srcProp || src))
  }, [srcProp, src])

  useEffect(() => {
    // Find the item in menuItems where key === dashboardKey
    const dashItem = menuItems.find((item) => item.key === dashboardKey)
    if (!dashItem?.router?.page?.url) {
      return
    }
    let baseUrl = dashItem.router.page.url

    // Append time range parameters
    // If baseUrl already contains '?', it means there are initial query parameters,
    // so use '&' to concatenate additional parameters.
    let connector = baseUrl.includes('?') ? '&' : '?'

    if (storeTimeRange.rangeType) {
      // If rangeType exists, find it in timeRangeList
      const storeTimeRangeItem = timeRangeList.find((x) => x.rangeType === storeTimeRange.rangeType)
      if (storeTimeRangeItem) {
        baseUrl += `${connector}from=${storeTimeRangeItem.from}&to=${storeTimeRangeItem.to}`
      }
    } else {
      // Otherwise, use processedTimeRange
      baseUrl += `${connector}from=${Math.round(startTime / 1000)}&to=${Math.round(endTime / 1000)}`
    }
    baseUrl += `&theme=${theme}`
    setIframeSrc(addEmbedCleanParam(baseUrl))
  }, [dashboardKey, menuItems, storeTimeRange, startTime, endTime, theme])

  useEffect(() => {
    return () => {
      observerRef.current?.disconnect()
      observerRef.current = null
    }
  }, [])

  return (
    <iframe
      id="iframe"
      ref={iframeRef}
      src={iframeSrc}
      width="100%"
      height="100%"
      frameBorder={0}
      onLoad={() => injectEmbedCleanStyle(iframeRef.current, observerRef)}
      key={iframeSrc}
    ></iframe>
  )
}

export default IframeDashboard
