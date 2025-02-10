import React, { useEffect, useState } from 'react'
// import Exceptions from './Exceptions'
import { ThemeColor, ThemeStyle } from 'src/constants'
import { getErrorTraceLogApi } from 'src/pro/api/cause'
import { Card, Tabs, TabsProps } from 'antd'
import { formatUnit } from 'src/core/utils/time'
import ErrorLog from './ErrorLog'
import ExceptionTab from './ExceptionTab'
// import { Card, Empty } from 'antd'
// import LoadingSpinner from 'src/core/components/Spinner'
// import { FaCircle } from 'react-icons/fa'
// import { formatUnit } from 'src/core/utils/time'

function ErrorNodeInfo(props: any) {
  const theme = ThemeStyle['dark']
  const { traceId, node, component } = props
  const [nodeColor, setNodeColor] = useState<string>('')
  const [activeKey, setActiveKey] = useState<string>('log')
  const [logMap, setLogMap] = useState<any>()
  const [openIdList, setOpenIdList] = useState<number[]>([])
  const [logCount, setLogCount] = useState<number>(0)
  const [exceptionsCount, setExceptionsCount] = useState<number>(0)
  const [loading, setLoading] = useState<boolean>(true)
  const items: TabsProps['items'] = [
    {
      key: 'log',
      label: 'Log',
      children: <ErrorLog logMap={logMap} logCount={logCount} loading={loading} node={node} />,
    },
    {
      key: 'exceptions',
      label: 'Exceptions',
      children: <ExceptionTab loading={loading} node={node} />,
    },
  ]

  const clickCollapse = (id: number) => {
    if (openIdList.indexOf(id) > -1) {
      setOpenIdList(openIdList.filter((item) => item !== id))
    } else {
      setOpenIdList(openIdList.concat([id]))
    }
  }

  useEffect(() => {
    setLoading(true)
    let color = ThemeColor.gray
    if (node?.isError) {
      color = ThemeColor.red
    }
    // if (node?.isProfiled) {
    //   color = ThemeColor.deepRed;
    // }
    setNodeColor(color)
    if (node && (node.isProfiled || node.hasOwnProperty('logs'))) {
      // 只有isProfiled为true才有log
      // 一个是报告中已有数据的，一个是弹窗里的 需要api获取
      if (node.logs) {
        setLogMap(node.logs)
        getLogCount(node.logs)
        setLoading(false)
      } else {
        getNodeLogInfo()
      }
    } else {
      setLogMap(null)
      setLogCount(0)
      setOpenIdList([])
      setLoading(false)
    }
    getExceptionsCount()

    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [node])
  const getNodeLogInfo = () => {
    setLogMap(null)
    getErrorTraceLogApi(traceId, node.spanId)
      .then((res) => {
        if (res.success && res.data) {
          setLogMap(res.data)
          getLogCount(res.data)
        }
      })
      .catch(() => {
        setLoading(false)
      })
      .finally(() => {
        setLoading(false)
      })
  }
  const getExceptionsCount = () => {
    let count = 0
    node?.errorSpans?.map((span: any) => {
      span.exceptions.map((exception: any) => {
        count++
      })
    })
    setExceptionsCount(count)
  }
  const getLogCount = (logMap: any) => {
    let count = 0
    let openId: any = null
    logMap.map((span: any) => {
      if (span?.id !== undefined && span?.name && span?.values.length > 0) {
        count++
        if (openId === null) {
          openId = span.id
        }
      }
    })
    setLogCount(count)
    setOpenIdList([openId])
  }
  // useEffect(() => {
  //   if (logMap && logMap.length > 0 && logMap[0]?.id !== undefined && logMap[0]?.name) {
  //     setOpenId(logMap[0].id);
  //   }
  // }, [logMap]);
  return (
    <>
      <Card
        className="error-node-info-card"
        title={
          node && node.hasOwnProperty('report_name') ? (
            <>{component ? component : node.report_name}</>
          ) : (
            <div>
              {/* <FaCircle style={{ color: nodeColor, marginRight: '10px' }} /> */}
              <>{node?.id && (component ? component(node?.id) : node.id)}</>
            </div>
          )
        }
      >
        {node && !node.hasOwnProperty('logs') && (
          <>
            <div>URL: {node?.url}</div>
            <div>isProfiled: {node?.isProfiled.toString()}</div>
            <div>本次调用: {formatUnit(node?.totalTime, 'ns')}</div>
          </>
        )}

        {node && (node.isError || node.isProfiled || node.hasOwnProperty('logs')) && (
          <div>
            <Tabs items={items} activeKey={activeKey} onChange={(key) => setActiveKey(key)} />
          </div>
        )}
      </Card>
    </>
  )
}
export default ErrorNodeInfo
