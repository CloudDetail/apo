import { CButton } from '@coreui/react'
import { Divider, Popover } from 'antd'
import React from 'react'
import { FaCircle } from 'react-icons/fa'
import ReactJson from 'react-json-view'
import { StatusColorMap } from 'src/constants'
import { convertTime } from 'src/core/utils/time'
import { useTranslation } from 'react-i18next'

function isJSONString(str) {
  try {
    JSON.parse(str)
    return true
  } catch (e) {
    return false
  }
}

function StatusInfo({ status, alertReason = [], title = null }) {
  const { t } = useTranslation('oss/service')
  console.log('StatusInfo:', title)

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
                    <span className="text-[#a1a1a1]">{t('StatusInfo.alertObjectText')}：</span>
                    {item.alertObject}
                  </div>
                  <div>
                    <span className="text-[#a1a1a1]">{t('StatusInfo.alertTimeText')}：</span>
                    {convertTime(item.timestamp, 'yyyy-mm-dd hh:mm:ss')}
                  </div>
                  <div>
                    <span className="text-[#a1a1a1]">{t('StatusInfo.alertReasonText')}：</span>
                    {item.alertReason}
                  </div>
                  <div>
                    <span className="text-[#a1a1a1]">{t('StatusInfo.detailsText')}：</span>
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
              <div className="text-[#a1a1a1] text-center pt-2">
                {t('StatusInfo.moreDetailsText')}
              </div>
            )}
            {/* {alertReason.length === 0 && t('StatusInfo.noReasonText')} */}
          </div>
        ) : null
      }
      title={
        ['critical', 'warning'].includes(status) && alertReason.length > 0
          ? title + t('StatusInfo.alertReasonText')
          : null
      }
    >
      <div className="p-2 w-full justify-center flex items-center h-full">
        <div>
          <FaCircle color={StatusColorMap[status]} />
        </div>
      </div>
    </Popover>
  )
}
export default StatusInfo
