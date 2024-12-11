import React from 'react'
import AnormalListCom from './AnormalListCom'
import ErrorInstanceTable from './ErrorInstanceTable'
import FaultSiteLogsWrapper from './FaultSiteLogs'
import InstanceTable from './InstanceTable'
import MetricsDashboardCom from './MetricsDashboard'
import ServiceTable from './ServiceTable'
import SlowType from './SlowType'

const DataViewCom = ({ type, data, analyserNode }) => {
  const component = () => {
    switch (type) {
      case 'service':
        return <ServiceTable data={data ? [data] : []} />
      case 'alert':
        return (
          <div className="flex p-2">
            <AnormalListCom {...data} selectedNodeInfo={analyserNode} />
          </div>
        )
      case 'endpoint':
        return <InstanceTable data={data.data} />

      case 'error':
        return <ErrorInstanceTable data={data.instances} />
      case 'log':
        return <FaultSiteLogsWrapper {...data} selectedNodeInfo={analyserNode} />
      case 'dashboard':
        const namespaceList = []
        const podList = []
        data?.forEach((item) => {
          namespaceList.push(item.namespace)
          podList.push(item.podList)
        })
        return (
          <MetricsDashboardCom
            variable={{ namespaceList, podList, service: analyserNode?.service }}
          />
        )
      case 'slow':
        return <SlowType {...data} />
    }
  }
  return <>{component()}</>
}
export default React.memo(DataViewCom)
