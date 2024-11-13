import { Collapse } from 'antd'
import React, { useState } from 'react'
import HighlightCode from 'src/core/components/HighlightCode/HighlightCode'
import { convertTime } from 'src/core/utils/time'

function ExceptionsTable(props: any) {
  const { exceptions, spanName } = props
  const [openExceptionsList, setOpenExceptionsList] = useState<number[]>([])

  const clickCollapsableSection = (index: number) => {
    if (openExceptionsList.includes(index)) {
      setOpenExceptionsList(openExceptionsList.filter((item) => item !== index))
    } else {
      setOpenExceptionsList([...openExceptionsList, index])
    }
  }
  const getItems = () => {
    return exceptions.map((item: any, index: number) => {
      return {
        key: index,
        label: item.message,
        children: (
          <>
            <div className="row">
              <div className={' col-0'}>Type：</div>
              <div className="flex-1 word-wrap">{item.type}</div>
            </div>
            <div className="row">
              <div className={' col-0'}>Time：</div>
              <div className="flex-1">{convertTime(item.timestamp, 'yyyy-mm-dd hh:mm:ss')}</div>
            </div>
            <div className="row">
              <div className={' col-0'}>Stack：</div>
              <div className="flex-1">
                {item.stack?.length > 0 ? (
                  <HighlightCode
                    rows={item.stack.split('\n')}
                    title={spanName + '-' + item.message}
                  />
                ) : (
                  '无'
                )}
              </div>
            </div>
          </>
        ),
      }
    })
  }
  return (
    <>
      <Collapse items={getItems()} defaultActiveKey={[0]}></Collapse>
    </>
  )
}
// const getStyles = (theme: GrafanaTheme2) => ({
//     exceptionLabel: css`
//     font-size: 14px;
//     font-weight: bold;
//     color: ${theme.colors.text.primary}
//     `
// })

export default ExceptionsTable
