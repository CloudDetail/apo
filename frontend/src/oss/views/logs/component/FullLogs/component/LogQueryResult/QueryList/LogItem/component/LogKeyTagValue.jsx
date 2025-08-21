/**
 * Copyright 2024 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import { useEffect, useState } from 'react'
import ReactJson from 'react-json-view'
import { Virtuoso } from 'react-virtuoso'
import LogTagDropDown from './LogTagDropdown'
import { useSelector } from 'react-redux'
import { theme } from 'antd'
import { useLogsContext } from 'src/core/contexts/LogsContext'
function isJSONString(str) {
  try {
    return typeof JSON.parse(str) === 'object' && JSON.parse(str) !== null
  } catch (e) {
    return false
  }
}
const determineTypeAndValue = (description, title) => {
  if (typeof description === 'string') {
    if (isJSONString(description)) {
      return { type: 'object', value: JSON.parse(description) }
    }
    if (description.length < 1000 && title !== 'content') {
      return { type: 'string', value: [description] }
    }
    return { type: 'longString', value: description.split('\n') }
  }
  if (typeof description === 'number' || typeof description === 'boolean') {
    return { type: typeof description, value: [String(description)] }
  }
  if (typeof description === 'object') {
    return { type: 'object', value: [description] }
  }
  return { type: null, value: [null] } // 默认情况
}

const formatValue = (value) => (typeof value === 'object' ? JSON.stringify(value) : value)

// 从query中提取全文检索关键词
const extractFullTextSearchKeywords = (query) => {
  if (!query) return []
  
  // 匹配 `content` LIKE '%keyword%' 模式
  const likePattern = /`content`\s+LIKE\s+'%([^%]+)%'/gi
  const keywords = []
  let match
  
  while ((match = likePattern.exec(query)) !== null) {
    keywords.push(match[1])
  }
  
  return keywords
}

// 检查文本中是否包含关键词
const hasKeywords = (text, keywords) => {
  if (!text || !keywords || keywords.length === 0) {
    return false
  }
  
  return keywords.some(keyword => {
    if (keyword && keyword.trim()) {
      const regex = new RegExp(keyword.replace(/[.*+?^${}()|[\]\\]/g, '\\$&'), 'gi')
      return regex.test(text)
    }
    return false
  })
}

// 统计匹配的内容数量（所有关键词的匹配次数总和）
const countMatches = (text, keywords) => {
  if (!text || !keywords || keywords.length === 0) {
    return 0
  }
  
  let totalMatches = 0
  
  keywords.forEach(keyword => {
    if (keyword && keyword.trim()) {
      const regex = new RegExp(keyword.replace(/[.*+?^${}()|[\]\\]/g, '\\$&'), 'gi')
      const matches = text.match(regex)
      if (matches) {
        totalMatches += matches.length
      }
    }
  })
  
  return totalMatches
}

// 统计多行文本中的匹配数量
const countMatchesInLines = (lines, keywords) => {
  if (!lines || !keywords || keywords.length === 0) {
    return 0
  }
  
  let totalMatches = 0
  lines.forEach(line => {
    totalMatches += countMatches(line, keywords)
  })
  
  return totalMatches
}

// 高亮文本函数
const highlightText = (text, keywords, theme) => {
  if (!text || !keywords || keywords.length === 0) {
    return text
  }
  
  // 根据主题定义高亮颜色
  const highlightColor = theme === 'dark' ? '#c3e88d' : '#f4bf75' // 使用base16-light的字符串颜色
  
  let highlightedText = text
  
  keywords.forEach(keyword => {
    if (keyword && keyword.trim()) {
      const regex = new RegExp(`(${keyword.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')})`, 'gi')
      highlightedText = highlightedText.replace(regex, `<span style="color: ${highlightColor}; font-weight: bold;">$1</span>`)
    }
  })
  
  return highlightedText
}

const LogKeyTagValue = ({ title, description, isHighlighted = false }) => {
  const { reactJsonTheme, theme: currentTheme } = useSelector((state) => state?.settingReducer || {})
  const { query } = useLogsContext()
  const [type, setType] = useState(null)
  const [value, setValue] = useState([null])
  const { useToken } = theme
  const { token } = useToken()
  
  // 提取全文检索关键词
  const fullTextKeywords = extractFullTextSearchKeywords(query)
  
  // 检查是否有匹配的关键词（用于显示提示信息）
  const [hasMatches, setHasMatches] = useState(false)
  const [matchesCount, setMatchesCount] = useState(0)
  
  useEffect(() => {
    if (title === 'content' && fullTextKeywords.length > 0) {
      let totalMatches = 0
      
      if (type === 'longString') {
        totalMatches = countMatchesInLines(value, fullTextKeywords)
      } else if (type === 'string') {
        totalMatches = countMatches(value[0], fullTextKeywords)
      }
      
      setMatchesCount(totalMatches)
      setHasMatches(totalMatches > 0)
    } else {
      setHasMatches(false)
      setMatchesCount(0)
    }
  }, [title, fullTextKeywords, value, type])
  useEffect(() => {
    const { type, value } = determineTypeAndValue(description, title)
    setType(type)
    setValue(value)
  }, [description, title])
  return (
    <div
      className="break-all  cursor-pointer w-full"
      style={{
        whiteSpace: 'break-spaces',
        wordBreak: 'break-all',
        overflow: 'hidden',
      }}
    >
      {type === 'object' ? (
        <div
          onClick={(e) => e.stopPropagation()} // 阻止事件冒泡
          style={{ width: '100%' }}
        >
          <ReactJson
            collapsed={1}
            src={value[0]}
            theme={reactJsonTheme}
            displayDataTypes={false}
            style={{ width: '100%' }}
            enableClipboard={true}
            name={false}
          />
        </div>
      ) : type === 'longString' ? (
        <div>
          <pre
            className="h-full w-full overflow-hidden text-xs leading-relaxed m-0"
            style={{
              whiteSpace: 'break-spaces',
              wordBreak: 'break-all',
              overflow: 'hidden',
              marginBottom: 3,
              color: token.colorTextSecondary,
              backgroundColor: token.colorBgContainer
            }}
          >
            {value?.length > 10 ? (
              <Virtuoso
                style={{ height: 400, width: '100%' }}
                overscan={1000}
                data={value}
                itemContent={(index, paragraph) => (
                  <div 
                    key={index} 
                    dangerouslySetInnerHTML={{ 
                      __html: title === 'content' && fullTextKeywords.length > 0 
                        ? highlightText(value[index], fullTextKeywords, currentTheme)
                        : value[index] 
                    }}
                  />
                )}
              />
            ) : (
              <div 
                dangerouslySetInnerHTML={{ 
                  __html: title === 'content' && fullTextKeywords.length > 0 
                    ? highlightText(value.join('\n'), fullTextKeywords, currentTheme)
                    : value.join('\n')
                }}
              />
            )}
                  </pre>
      </div>
      ) : (
        <div>
          <div 
                 className={`hover:underline ${
         isHighlighted
           ? 'font-semibold border rounded px-1'
           : ''
       }`}
       style={isHighlighted ? (() => {
         if (currentTheme === 'dark') {
           return {
           // 暗色主题的高亮颜色
             color: '#c3e88d',
             borderColor: '#c3e88d',
             backgroundColor: '#444444'
           }
         } else {
           // 亮色主题的高亮颜色 - 使用base16-light的字符串颜色
           return {
             color: '#f4bf75',
             borderColor: '#f4bf75',
             backgroundColor: '#faf8f0'
           }
         }
       })() : {}}
          >
            {title === 'content' && fullTextKeywords.length > 0 ? (
              <div dangerouslySetInnerHTML={{ __html: highlightText(value[0], fullTextKeywords, currentTheme) }} />
            ) : (
              value[0]
            )}
                  </div>
      </div>
      )}
    </div>
  )
}
export default LogKeyTagValue
