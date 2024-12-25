/**
 * Copyright 2024 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import React, { Component } from 'react'
import ErrorPage from 'src/core/assets/errorPage.svg'
import { Image } from 'antd'
class ErrorBoundary extends Component {
  constructor(props) {
    super(props)
    this.state = { hasError: false }
  }

  static getDerivedStateFromError(error) {
    // 当捕获到错误时，更新 state
    return { hasError: true }
  }

  componentDidCatch(error, errorInfo) {
    // 可以在此处记录错误信息到日志服务
    console.error('Error caught by ErrorBoundary:', error, errorInfo)
  }

  render() {
    if (this.state.hasError) {
      // 自定义错误页面
      return (
        <div className="w-screen h-screen flex  items-center justify-center flex-col">
          <Image src={ErrorPage} width={'30%'} />
          <div>页面遇到未知错误</div>
        </div>
      )
    }
    return this.props.children
  }
}

export default ErrorBoundary
