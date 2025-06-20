/**
 * Copyright 2024 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import React, { useEffect, useState } from 'react'
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
  const { tableInfo } = useLogsContext()
  const [nullFieldVisibility, setNullFieldVisibility] = useState(false)
  const { t } = useTranslation('oss/fullLogs')
  const { theme } = useSelector((state) => state.settingReducer)

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
            <LogItemDetail log={log} contentVisibility={!tableInfo?.timeField} nullFieldVisibility={nullFieldVisibility} />
          </>
        }
      </div>
    </div>
  )
}
export default LogItem
