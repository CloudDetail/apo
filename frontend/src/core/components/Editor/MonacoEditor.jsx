/**
 * Copyright 2024 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import Editor from '@monaco-editor/react'
import { promLanguageDefinition } from 'monaco-promql'
import { Resizable } from 'react-resizable'
import 'react-resizable/css/styles.css'
import React, { useEffect, useState } from 'react'
import { useSelector } from 'react-redux'

const MonacoEditorWrapper = ({ defaultValue, handleEditorChange = null, readOnly = false }) => {
  const { monacoPromqlConfig } = useSelector((state) => state.settingReducer)
  const { theme } = useSelector((state) => state.settingReducer)

  const [editorHeight, setEditorHeight] = useState(50);

  const handleResize = (e, { size }) => {
    setEditorHeight(size.height);
  };

  const handleEditorDidMount = async (editor, monaco) => {
    const languageId = promLanguageDefinition.id
    // 注册 PromQL 语言
    monaco.languages.register({ id: languageId })
    if (monacoPromqlConfig) {
      // 注册语言定义和补全服务
      monaco.languages.setMonarchTokensProvider(languageId, monacoPromqlConfig.language)
      monaco.languages.setLanguageConfiguration(
        languageId,
        monacoPromqlConfig.languageConfiguration,
      )
      monaco.languages.registerCompletionItemProvider(
        languageId,
        monacoPromqlConfig.completionItemProvider,
      )
    }
  }

  return (
    <Resizable height={editorHeight} width={Infinity} axis="y" resizeHandles={['s']} onResize={handleResize}>
      <div style={{ height: `${editorHeight}px`, position: 'relative' }}>
        <Editor
          width="100%"
          height="100%"
          theme={theme === 'light' ? 'vs' : 'vs-dark'}
          language="promql"
          options={{
            readOnly: readOnly,
            lineNumbers: 'off', // 取消行号
            minimap: { enabled: false }, // 取消右侧迷你地图
            wordWrap: 'on',
            scrollbar: {
              vertical: 'hidden', // 隐藏垂直滚动条
              horizontal: 'hidden', // 隐藏水平滚动条
            },
            glyphMargin: false, // 去掉行号左边的边距
            disableLayerHinting: true,
            hideCursorInOverviewRuler: true,
            overviewRulerBorder: false,
            lineDecorationsWidth: 0, // 去掉装订线
            folding: false, // 去掉代码折叠的装订线
            quickSuggestions: true, // 启用快速建议
            suggestOnTriggerCharacters: true, // 触发字符时自动补全
            autoClosingBrackets: 'always', // 自动关闭括号
            acceptSuggestionOnEnter: 'on', // 按回车键接受建议
          }}
          value={defaultValue}
          onMount={handleEditorDidMount}
          onChange={handleEditorChange}
        />
      </div>
    </Resizable>
  )
}

export default MonacoEditorWrapper
