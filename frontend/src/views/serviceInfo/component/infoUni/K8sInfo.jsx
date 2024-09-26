import { CAccordionBody, CCol, CRow } from '@coreui/react'
import React, { useState, useEffect } from 'react'
import { useSelector } from 'react-redux'
import { useDebounce } from 'react-use'
import { getK8sEventApi } from 'src/api/serviceInfo'
import Empty from 'src/components/Empty/Empty'
import LoadingSpinner from 'src/components/Spinner'
import { usePropsContext } from 'src/contexts/PropsContext'
import { selectProcessedTimeRange } from 'src/store/reducers/timeRangeReducer'
function K8sInfo(props) {
  const { handlePanelStatus } = props
  const [data, setData] = useState({})
  const { serviceName } = usePropsContext()
  const [colList, setColList] = useState([])
  const [loading, setLoading] = useState(false)
  const { startTime, endTime } = useSelector(selectProcessedTimeRange)
  const mockColList = [
    {
      name: '应用变更失败',
      status: 'success',
      value: 0,
      weekValue: 0,
      monthValue: 0,
    },
    {
      name: '应用扩缩容',
      status: 'success',
      value: 2,
      weekValue: 664,
      monthValue: 881,
    },
    {
      name: '应用扩缩绒到达上下限',
      status: 'success',
      value: 2,
      weekValue: 642,
      monthValue: 848,
    },
    {
      name: '离群摘除',
      status: 'success',
      value: 0,
      weekValue: 0,
      monthValue: 0,
    },
    {
      name: 'pod启动失败',
      status: 'error',
      value: 21,
      weekValue: 3700,
      monthValue: 15447,
    },
    {
      name: '镜像拉取失败',
      status: 'success',
      value: 3,
      weekValue: 98,
      monthValue: 98,
    },
    {
      name: 'POD被驱逐',
      status: 'success',
      value: 0,
      weekValue: 0,
      monthValue: 0,
    },
    {
      name: 'POD OOM',
      status: 'error',
      value: 24,
      weekValue: 4243,
      monthValue: 18242,
    },
    {
      name: 'k8s集群资源不足',
      status: 'success',
      value: 0,
      weekValue: 0,
      monthValue: 0,
    },
    {
      name: 'K8s节点 OOM',
      status: 'success',
      value: 0,
      weekValue: 0,
      monthValue: 0,
    },
    {
      name: 'K8s节点重启',
      status: 'success',
      value: 0,
      weekValue: 0,
      monthValue: 0,
    },
    {
      name: 'K8s节点FD不足',
      status: 'success',
      value: 0,
      weekValue: 0,
      monthValue: 0,
    },
  ]
  const getData = () => {
    if (startTime && endTime) {
      setLoading(true)
      getK8sEventApi({
        startTime: startTime,
        endTime: endTime,
        service: serviceName,
      })
        .then((res) => {
          setData(res.data ?? {})
          setLoading(false)
          handlePanelStatus(res.status)
        })
        .catch((error) => {
          setData({})
          handlePanelStatus('unknown')
          setLoading(false)
        })
    }
  }
  // useEffect(() => {
  //   getData()
  // }, [serviceName, startTime, endTime])
  //防抖避免跳转使用旧时间
  useDebounce(
    () => {
      getData()
    },
    300, // 延迟时间 300ms
    [startTime, endTime, serviceName],
  )
  return (
    <>
      <CAccordionBody>
        <CRow xs={{ cols: 6 }}>
          {Object.keys(data).map((key) => {
            const item = data[key]
            return (
              <CCol key={key} className="text-center py-3.5">
                <div className="text-sm mb-2">{item.displayName}</div>
                <div
                  className="text-lg mb-2 text-[#467ffc] font-bold"
                  style={{ color: item.severity === 'Warning' ? '#dc2625' : '#467ffc' }}
                >
                  {item.counts.current ?? 0}
                </div>
                <div className="text-xs mb-1" style={{ color: 'rgba(248, 249, 250, 0.45)' }}>
                  次数(7天):{item.counts.lastWeek}
                </div>
                <div className="text-xs" style={{ color: 'rgba(248, 249, 250, 0.45)' }}>
                  次数(30天):{item.counts.lastMonth}
                </div>
              </CCol>
            )
          })}
          <LoadingSpinner loading={loading} />
          {!loading && (!data || Object.keys(data).length === 0) && <Empty />}
        </CRow>
      </CAccordionBody>
    </>
  )
}

export default K8sInfo
