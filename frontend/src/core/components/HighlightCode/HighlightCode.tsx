/**
 * Copyright 2024 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import React, { useCallback, useEffect, useRef, useState } from 'react'
import Highlighter from 'react-highlight-words'
import { saveAs } from 'file-saver'
import { VariableSizeList as List } from 'react-window'
import TextArea from 'antd/es/input/TextArea'
import { Button } from 'antd'
import { MdArrowDownward, MdArrowUpward } from 'react-icons/md'
import { ThemeStyle } from 'src/constants'
import { IoCloudDownloadOutline } from 'react-icons/io5'
import { useTranslation } from 'react-i18next'
export default function HighlightCode(props: any) {
  const { t } = useTranslation('common')
  const rowHeights = useRef<Record<number, number>>({})

  const { searchWord, title, rows } = props
  const [searchWords, setSearchWords] = useState<any>([])
  const [activeIndex, setActiveIndex] = useState<number>(-1)
  const [count, setCount] = useState<number>(0)
  const [wrapActive, setWrapActive] = useState<boolean>(false)
  const divRef = useRef<any>(null)
  const listRef = useRef<any>(null)
  const preRef = useRef<any>(null)
  const innerRef = useRef<any>(null)
  const [activeIndexMap, setActiveIndexMap] = useState<any>([])

  const wrapperRef = useRef(null)
  const theme = ThemeStyle['dark']
  useEffect(() => {
    // 定义点击事件处理程序
    function handleClickOutside(event) {
      if (wrapperRef.current && !wrapperRef.current.contains(event.target)) {
        setWrapActive(false)
      }
    }

    // 监听全局点击事件
    document.addEventListener('mousedown', handleClickOutside)
    return () => {
      // 在组件卸载时移除事件监听
      document.removeEventListener('mousedown', handleClickOutside)
    }
  }, [])
  useEffect(() => {
    divRef.current.style.pointerEvents = wrapActive ? 'auto' : 'none'
    preRef.current.style.outline = wrapActive ? 'rgb(110, 159, 255) solid 2px' : 'none'
  }, [wrapActive])
  useEffect(() => {
    if (searchWord) {
      setSearchWords([searchWord])
    }
  }, [searchWord])
  useEffect(() => {
    const activeIndexMap: any[] = []

    if (searchWords[0]) {
      const regex = new RegExp(searchWords[0]?.replace(/[.*+?^${}()|[\]\\]/g, '\\$&'), 'gi')
      // const matches = textToHighlight.match(regex);

      // setCount(matches ? matches.length : 0);
      rows.map((row: any, rowIndex: number) => {
        const matches = row?.match(regex) ?? []
        matches?.forEach((item: any, index: number) => {
          activeIndexMap.push(rowIndex + '-' + index)
        })
      })
      setActiveIndexMap(activeIndexMap)
      setActiveIndex(-1)
      setCount(activeIndexMap ? activeIndexMap.length : 0)
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [searchWords])

  const Highlight = (props: any, rowIndex: number) => {
    const { children, highlightIndex } = props
    return (
      <strong
        id={searchWords[0] + rowIndex + '-' + highlightIndex}
        style={{
          background:
            rowIndex + '-' + highlightIndex === activeIndexMap[activeIndex] ? 'orange' : 'yellow',
          color: theme.colors.background.primary,
        }}
      >
        {children}
      </strong>
    )
  }
  const changeActiveIndex = (action: string) => {
    let index = -1
    if (action === 'sub') {
      if (activeIndex > 0) {
        index = activeIndex - 1
      } else {
        index = count - 1
      }
    } else {
      if (activeIndex > count - 2) {
        index = 0
      } else {
        index = activeIndex + 1
      }
    }
    setActiveIndex(index)
    const rowIndex = Number(activeIndexMap[index].split(/-(.+)/)[0])
    listRef.current.scrollToItem(rowIndex, 'smart')
    requestAnimationFrame(() => {
      if (index > -1 && searchWords[0]) {
        const targetElementId = searchWords[0] + activeIndexMap[index]
        const targetElement = document.getElementById(targetElementId)
        if (listRef.current && divRef.current && targetElement) {
          const elementTop =
            targetElement.offsetTop + (targetElement.offsetParent as HTMLElement)?.offsetTop
          const elementCenter = elementTop
          const containerCenter = divRef.current.offsetHeight / 2
          let scrollTo = elementCenter - containerCenter
          listRef.current.scrollTo(scrollTo)
        }
      }
    })
  }
  const exportCode = () => {
    const blob = new Blob([rows.join('\n')], { type: 'text/plain;charset=utf-8' })
    saveAs(blob, `${title}.txt`)
  }
  const getRowHeight = useCallback((index: number) => {
    return rowHeights.current[index] || 100
  }, [])

  const setRowHeight = useCallback((index: number, size: number) => {
    listRef.current?.resetAfterIndex(0)
    rowHeights.current = { ...rowHeights.current, [index]: size }
  }, [])
  const Row = ({ index, style }: any) => {
    const rowRef = useRef<HTMLDivElement>(null)
    useEffect(() => {
      if (rowRef.current) {
        setRowHeight(index, (rowRef.current?.firstElementChild as HTMLDivElement).offsetHeight ?? 0)
      }
    }, [index, rowRef])

    return (
      <div style={style} ref={rowRef} className="text-xs text-[var(--ant-color-text)]">
        <Highlighter
          highlightClassName=""
          searchWords={searchWords}
          autoEscape={true}
          textToHighlight={rows[index]}
          activeIndex={activeIndex}
          activeClassName="test"
          className="text-wrap break-words"
          highlightTag={(e) => Highlight(e, index)}
        />
      </div>
    )
  }

  return (
    <div ref={wrapperRef}>
      <pre ref={preRef} onClick={() => setWrapActive(true)} style={{ margin: 5 }}>
        <div className="flex-center">
          <TextArea
            onChange={(e) => {
              setSearchWords([e.currentTarget.value])
            }}
            placeholder={t('search')}
          />
          <Button
            disabled={count === 0 || searchWords.length === 0}
            type="text"
            icon={<MdArrowUpward />}
            onClick={() => changeActiveIndex('sub')}
          ></Button>
          <Button
            icon={<MdArrowDownward />}
            disabled={count === 0 || searchWords.length === 0}
            type="text"
            onClick={() => changeActiveIndex('add')}
          ></Button>

          <Button icon={<IoCloudDownloadOutline />} type="link" onClick={exportCode}></Button>
        </div>
        <div
          className="overflow-hidden"
          style={{ height: 420, overflow: 'hidden', marginTop: 10, paddingTop: 10 }}
          ref={divRef}
        >
          <List
            height={400}
            itemCount={rows?.length}
            itemSize={getRowHeight}
            width="100%"
            style={{ maxHeight: 400, overflowY: 'auto', overflowX: 'hidden' }}
            ref={listRef}
            innerRef={innerRef}
            overscanCount={0}
          >
            {Row}
          </List>
        </div>
      </pre>
    </div>
  )
}
