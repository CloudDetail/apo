import { CButton } from '@coreui/react'
import { Divider, Popover } from 'antd'
import React from 'react'
import { FaCircle } from 'react-icons/fa'
import ReactJson from 'react-json-view'
import { StatusColorMap } from 'src/constants'
import { convertTime } from 'src/utils/time'
function isJSONString(str) {
  try {
    JSON.parse(str)
    return true
  } catch (e) {
    return false
  }
}
function StatusInfo({ status, alertReason = [], title = null }) {
  console.log(title, alertReason.slice(0, 3))
  return (
    <Popover
      content={
        ['critical', 'warning'].includes(status) && alertReason.length > 0 ? (
          <div className="max-w-[400px]">
            <div className="max-h-[300px] overflow-y-auto text-xs">
              {alertReason.slice(0, 3).map((item, index) => (
                <div key={index}>
                  {index > 0 && <Divider />}
                  <div>
                    <span className="text-[#a1a1a1]">告警对象：</span>
                    {item.alertObject}
                  </div>
                  <div>
                    <span className="text-[#a1a1a1]">告警时间：</span>
                    {convertTime(item.timestamp, 'yyyy-mm-dd hh:mm:ss')}
                  </div>
                  <div>
                    <span className="text-[#a1a1a1]">告警原因：</span>
                    {item.alertReason}
                  </div>
                  <div>
                    <span className="text-[#a1a1a1]">细节：</span>
                    {isJSONString(item.alertMessage) ? (
                      <ReactJson
                        src={JSON.parse(item.alertMessage)}
                        theme="brewer"
                        collapsed={false}
                        displayDataTypes={false}
                        style={{ width: '100%' }}
                        enableClipboard={false}
                      />
                    ) : (
                      item.alertMessage
                    )}
                  </div>
                </div>
              ))}
            </div>
            {alertReason.length > 3 && (
              <div className="text-[#a1a1a1] text-center pt-2">更多请查看服务详情的告警事件</div>
              // <CButton color="info" variant="ghost" size="sm" onClick={clickMore}>
              //   更多
              // </CButton>
            )}
            {/* {alertReason.length === 0 && '暂未检测到告警原因'} */}
          </div>
        ) : null
      }
      title={
        ['critical', 'warning'].includes(status) && alertReason.length > 0
          ? title + '告警原因'
          : null
      }
    >
      <div className="p-2 w-full justify-center flex items-center h-full">
        <FaCircle color={StatusColorMap[status]} />
      </div>
    </Popover>
  )
}
export default StatusInfo
