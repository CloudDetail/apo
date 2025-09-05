/**
 * Copyright 2024 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import React, { useEffect, useState, useMemo } from 'react'
import { AiFillCaretDown, AiFillCaretRight } from 'react-icons/ai'
import LogItemFold from './component/LogItemFold'
import LogItemDetail from './component/LogItemDetail'
import { Button, Tag } from 'antd'
import { useLogsContext } from 'src/core/contexts/LogsContext'
import { convertTime } from 'src/core/utils/time'
import { useTranslation } from 'react-i18next' // 引入i18n
import { useSelector } from 'react-redux'

const LogItem = (props) => {
  const { log, openContextModal } = props
  const { tableInfo, query } = useLogsContext()
  const [nullFieldVisibility, setNullFieldVisibility] = useState(false)
  const { t } = useTranslation('oss/fullLogs')
  const { theme } = useSelector((state) => state.settingReducer)
  
  // 全文检索匹配提示逻辑
  const fullTextSearchInfo = useMemo(() => {
    // 从 query 中提取全文检索关键词
    const extractFullTextSearchKeywords = (query) => {
      if (!query) return []
      const likePattern = /`content`\s+LIKE\s+'%([^%]+)%'/gi
      const keywords = []
      let match
      while ((match = likePattern.exec(query)) !== null) {
        keywords.push(match[1])
      }
      return keywords
    }
    
    // 统计匹配数量
    const countMatches = (text, keywords) => {
      if (!text || !keywords || keywords.length === 0) return 0
      let totalMatches = 0
      keywords.forEach(keyword => {
        if (keyword && keyword.trim()) {
          const regex = new RegExp(keyword.replace(/[.*+?^${}()|[\]\\]/g, '\\$&'), 'gi')
          const matches = text.match(regex)
          if (matches) totalMatches += matches.length
        }
      })
      return totalMatches
    }
    
    const keywords = extractFullTextSearchKeywords(query)
    const content = log?.content || ''
    const matchesCount = keywords.length > 0 ? countMatches(content, keywords) : 0
    
    return {
      keywords,
      matchesCount,
      hasMatches: matchesCount > 0
    }
  }, [query, log?.content])

  return (
    <div className="flex flex-col overflow-hidden px-2 w-full">
      {/* icon 和 时间 */}
      <div className="flex-grow-0 flex-shrink-0 w-full">
        <div className="flex items-center gap-2 pb-2 j">
          <div className="flex-shrink-0 flex-grow-0 flex items-center">
            {/* <span className='w-1 h-5 bg-[var(--ant-color-primary)] rounded-md'></span> */}
            <span className={`${theme === 'dark' ? 'text-[var(--ant-color-text)] font-semibold bg-[var(--ant-color-primary-bg-hover)]' : 'text-[var(--ant-color-text)]'} tracking-wide border-l-4 px-2 border-[var(--ant-color-primary)] bg-[var(--ant-color-primary-bg)]`}>
              {convertTime(log?.timestamp, 'yyyy-mm-dd hh:mm:ss.SSS')}
            </span>
          </div>
          {openContextModal && !tableInfo.timeField && (
            <Button
              color="primary"
              variant="outlined"
              size="small"
              onClick={() => openContextModal(log)}
              className="text-xs"
            >
              {t('queryList.logItem.viewContextText')}
            </Button>
          )}
          <Button
            color="primary"
            variant="outlined"
            size="small"
            onClick={() => setNullFieldVisibility(!nullFieldVisibility)}
            className="text-xs"
          >
            {nullFieldVisibility && t("logQueryResult.hideNull")}
            {!nullFieldVisibility && t("logQueryResult.displayNull")}
          </Button>
        </div>
      </div>
      {/* 具体日志 */}
      <div className="flex-1 overflow-hidden w-full">
        {
          <>
            <LogItemFold tags={log.tags} />
            {/* 全文检索匹配提示 */}
            {fullTextSearchInfo.hasMatches && (
              <div 
                className="px-2 py-1 text-xs rounded flex items-center" 
                style={(() => {
                  if (theme === 'dark') {
                    return {
                    // 暗色主题的全文检索提示颜色
                      backgroundColor: '#c3e88d20',
                      borderLeft: '3px solid #c3e88d',
                      color: '#c3e88d',
                      display: 'inline-block'
                    }
                  } else {
                    // 亮色主题的全文检索提示颜色 - 使用base16-light的字符串颜色
                    return {
                      backgroundColor: '#f4bf7520',
                      borderLeft: '3px solid #f4bf75',
                      color: '#f4bf75',
                      display: 'inline-block'
                    }
                  }
                })()}
              >
                {localStorage.getItem('i18nextLng') === 'en' 
                  ? `Found ${fullTextSearchInfo.matchesCount} matches in content(Keywords: ${fullTextSearchInfo.keywords.join(', ')})`
                  : `在content中找到 ${fullTextSearchInfo.matchesCount} 处匹配（关键词：${fullTextSearchInfo.keywords.join(', ')}）`
                }
              </div>
            )}
            <LogItemDetail log={log} contentVisibility={!tableInfo?.timeField} nullFieldVisibility={nullFieldVisibility} />
          </>
        }
      </div>
    </div>
  )
}
export default LogItem
