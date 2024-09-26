import { CAccordionBody, CTab, CTabContent, CTabList, CTabPanel, CTabs } from '@coreui/react'
import React, { useEffect, useMemo, useState } from 'react'
import { useSelector } from 'react-redux'
import { getServiceAlertEventApi } from 'src/api/serviceInfo'
import { RuleGroupMap } from 'src/constants'
import { usePropsContext } from 'src/contexts/PropsContext'
import { selectProcessedTimeRange } from 'src/store/reducers/timeRangeReducer'
import AlertInfoTable from './AlertInfoTable'
import Empty from 'src/components/Empty/Empty'
import { useDebounce } from 'react-use'

export default function AlertInfoTabs(props) {
  const { handlePanelStatus } = props
  const { serviceName, endpoint } = usePropsContext()
  const [loading, setLoading] = useState(true)
  const { startTime, endTime } = useSelector(selectProcessedTimeRange)

  const [tabList, setTabList] = useState([])
  const prepareData = (result) => {
    let tabList = []
    Object.keys(result.events).forEach((group) => {
      let num = 0
      const groupMap = result.events[group]

      const data = Object.keys(groupMap).map((item) => {
        num += groupMap[item].length
        return {
          name: item,
          list: groupMap[item],
        }
      })
      tabList.push({
        key: group,
        name: RuleGroupMap[group],
        data,
        num,
      })
    })
    setTabList(tabList)
  }
  const getAlertEvents = () => {
    if (startTime && endTime) {
      setLoading(true)
      getServiceAlertEventApi({
        startTime,
        endTime,
        service: serviceName,
        endpoint: endpoint,
      })
        .then((res) => {
          setLoading(false)
          prepareData(res)
          handlePanelStatus(res.status)
        })
        .catch((error) => {
          setTabList([])
          handlePanelStatus('unknown')
          setLoading(false)
        })
    }
  }
  // useEffect(() => {
  //   getAlertEvents()
  // }, [startTime, endTime, serviceName])

  //防抖避免跳转使用旧时间
  useDebounce(
    () => {
      getAlertEvents()
    },
    300, // 延迟时间 300ms
    [serviceName, startTime, endTime, endpoint],
  )
  return (
    <CAccordionBody>
      {tabList?.length > 0 && (
        <CTabs activeItemKey={tabList[0].key}>
          <CTabList variant="tabs">
            {tabList.map((tab) => (
              <CTab itemKey={tab.key}>{tab.name}</CTab>
            ))}
          </CTabList>
          <CTabContent>
            {tabList.map((tab) => (
              <CTabPanel className="p-3 text-xs" itemKey={tab.key}>
                <AlertInfoTable data={tab.data} />
              </CTabPanel>
            ))}
          </CTabContent>
        </CTabs>
      )}
      {tabList?.length === 0 && !loading && <Empty context="无告警事件" />}
    </CAccordionBody>
  )
}
