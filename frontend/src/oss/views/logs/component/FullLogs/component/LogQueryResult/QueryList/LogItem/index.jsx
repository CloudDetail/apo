/**
 * Copyright 2024 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import React, { useEffect, useState } from 'react'
import { AiFillCaretDown, AiFillCaretRight } from 'react-icons/ai'
import LogItemFold from './component/LogItemFold'
import LogItemDetail from './component/LogItemDetail'
import { Button } from 'antd'
import { useLogsContext } from 'src/core/contexts/LogsContext'
import { convertTime } from 'src/core/utils/time'
import { useTranslation } from 'react-i18next' // 引入i18n

const LogItem = (props) => {
  const { log, openContextModal } = props
  const { tableInfo } = useLogsContext()
  const { t } = useTranslation('oss/fullLogs')

  return (
    <div className="flex overflow-hidden px-2 w-full">
      {/* icon 和 时间 */}
      <div className="flex-grow-0 flex-shrink-0  w-[230px]">
        <div className="items-center pl-2 j">
          <div className="flex-shrink-0 flex-grow-0 flex items-center">
            <span>{convertTime(log?.timestamp, 'yyyy-mm-dd hh:mm:ss.SSS')}</span>
          </div>
          {openContextModal && !tableInfo.timeField && (
            <Button
              color="primary"
              variant="filled"
              size="small"
              onClick={() => openContextModal(log)}
              className="text-xs"
            >
              {t('queryList.logItem.viewContextText')}
            </Button>
          )}
        </div>
      </div>
      {/* 具体日志 */}
      <div className="flex-1 overflow-hidden">
        {
          <>
            <LogItemFold tags={log.tags} />
            <LogItemDetail log={log} contentVisibility={!tableInfo?.timeField} />
          </>
        }
      </div>
    </div>
  )
}
export default LogItem
