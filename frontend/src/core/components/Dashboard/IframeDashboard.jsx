/**
 * Copyright 2024 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import React, { useEffect, useRef, useState } from 'react'
import { useSelector } from 'react-redux'
import { selectProcessedTimeRange, timeRangeList } from 'src/core/store/reducers/timeRangeReducer'

//receiving 'dashboardKey' as a prop ..
function IframeDashboard({ dashboardKey }) {
  const [src, setSrc] = useState()
  const storeTimeRange = useSelector((state) => state.timeRange)
  const { startTime, endTime } = useSelector(selectProcessedTimeRange)
  const menuItems = useSelector((state) => state.userReducer.menuItems);
  const iframeRef = useRef(null)
  // console.log(
  //   location,
  //   baseUrl +
  //     '/d/e9133aeb-5903-494e-90d4-4455eab65d47/e69585-e99a9c-e88a82-e782b9-e58886-e69e90-2?orgId=1',
  // )
  const handleLoad = () => {
    if (iframeRef.current) {
      const iframeDocument =
        iframeRef.current.contentDocument || iframeRef.current.contentWindow.document
      if (iframeDocument) {
        const someElement = iframeDocument.querySelector('selector') // 替换 'selector' 为你要选择的元素
      } else {
      }
    }
  }

  useEffect(() => {
    const iframe = iframeRef.current

    iframe.addEventListener('load', handleLoad)

    return () => {
      iframe.addEventListener('load', handleLoad)
    }
  }, [])

  useEffect(() => {
    const iframe = iframeRef.current

    const handleLoad = () => {
      const iframeDocument =
        iframeRef.current.contentDocument || iframeRef.current.contentWindow.document
      const someElement = iframeDocument.querySelector('selector') // 替换 'selector' 为你要选择的元素
      // const targetDiv = iframe.ownerDocument.documentElement.querySelector('#react-root .main-view > div:first-child');
      // console.log(targetDiv)
      // 获取 iframe 元素，并通过类型断言确保其为 HTMLIFrameElement
      // const iframeDocument = iframeRef.current.contentDocument || iframeRef.current.contentWindow.document;
      const test = document.getElementById('reactRoot')

      // 现在可以安全地访问 contentDocument 属性了
      const iframeDoc = iframe.contentDocument

      // 使用 iframeDoc 来进行 DOM 操作
      if (test) {
        console.log(test.querySelector('#reactRoot .grafana-app .main-view > div'))
      }
      // 确保先定义 onload 处理程序再设置 src 或执行其他加载操作
      iframe.onload = function () {
        // 在这里，iframe 已经加载完成，可以安全地访问内容
        const iframeDoc = iframe.contentDocument || iframe.contentWindow.document
        // console.log(iframeDoc.querySelector('#reactRoot .grafana-app .main-view'))
        const firstDiv = iframeDoc.querySelector('#reactRoot .grafana-app .main-view > div')
        // console.log(firstDiv) // 这应该会显示第一个 div 元素，如果它存在的话
        // 创建一个观察器实例并传入回调函数
        const observer = new MutationObserver(function (mutationsList, observer) {
          // 检查每一个变动记录是否有新增的子节点
          for (let mutation of mutationsList) {
            if (mutation.type === 'childList') {
              // 检查是否有新添加的子节点
              if (mutation.addedNodes.length > 0) {
                // 做一些事情，比如获取新节点或其它操作
                const first = iframeDoc.querySelector('.main-view > div > div:first-child')
                const last = iframeDoc.querySelector('.main-view > div > div:last-child')
                if (first && last) {
                  console.log(first)
                  console.log(last)
                  // if (top) {
                  first.style.display = 'none'
                  const icon = last.querySelector('div')
                  const nav = last.querySelector('nav')
                  if (icon && nav) {
                    icon.style.display = 'none'
                    nav.style.display = 'none'
                  }
                  // }
                  // if (left) {

                  //   left.style.display = 'none'
                  // }
                  observer.disconnect() // 如果找到了 div，断开 observer
                  return
                }
              }
            }
          }
        })

        // 使用配置对象开始观察目标节点
        observer.observe(iframeDoc, { attributes: true, childList: true, subtree: true })
        if (firstDiv) {
        }
      }

      // 然后设置 src 或执行其他会导致 iframe 重新加载内容的操作
      // iframe.src = 'your_iframe_source.html';

      // iframe.onload = function () {
      //   // 访问iframe的内容
      //   const iframeDocument = iframe.contentDocument || iframe.contentWindow.document;

      //   // 操作iframe中的DOM
      //   const targetElement = iframeDocument.getElementById('targetElementId');
      //   targetElement.style.display = 'none'; // 例如，隐藏一个元素
      // };
    }
    iframe.addEventListener('load', handleLoad)

    return () => {
      iframe.removeEventListener('load', handleLoad)
    }
  }, [])

  useEffect(() => {
    // Find the item in menuItems where key === dashboardKey
    const dashItem = menuItems.find((item) => item.key === dashboardKey);
    if (!dashItem?.router?.page?.url) {
      return;
    }
    let baseUrl = dashItem.router.page.url;
  
    // Append time range parameters
    // If baseUrl already contains '?', it means there are initial query parameters, 
    // so use '&' to concatenate additional parameters.
    let connector = baseUrl.includes('?') ? '&' : '?';
  
    if (storeTimeRange.rangeType) {
      // If rangeType exists, find it in timeRangeList
      const storeTimeRangeItem = timeRangeList.find(
        (x) => x.rangeType === storeTimeRange.rangeType
      );
      if (storeTimeRangeItem) {
        baseUrl += `${connector}from=${storeTimeRangeItem.from}&to=${storeTimeRangeItem.to}`;
      }
    } else {
      // Otherwise, use processedTimeRange
      baseUrl += `${connector}from=${Math.round(startTime / 1000)}&to=${Math.round(endTime / 1000)}`;
    }
  
    setSrc(baseUrl);
  }, [dashboardKey, menuItems, storeTimeRange, startTime, endTime]);
  
 

  return (
    <iframe
      id="iframe"
      ref={iframeRef}
      src={src}
      width="100%"
      height="100%"
      frameBorder={0}
      onLoad={handleLoad}
      key={src}
    ></iframe>
  )
}

export default IframeDashboard