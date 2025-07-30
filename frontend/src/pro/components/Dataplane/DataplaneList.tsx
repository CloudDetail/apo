import { useEffect, useState } from 'react'
import { getDataplaneTypeListApi } from 'src/pro/api/dataplane'
import DataplaneCard from './DataplaneCard'

interface DataplaneItem {
  typeName: string
  paramSpec: any
  desc: string
  capabilitySpec: string[]
}

const DataplaneList = ({ onClick }: { onClick: (type: string) => void }) => {
  const [dataplaneList, setDataplaneList] = useState<DataplaneItem[]>([])

  useEffect(() => {
    getDataplaneTypeListApi().then((res) => {
      setDataplaneList(res.dpTypes)
    })
  }, [])
  const defaultDataplane = [
    {
      typeName: 'Apo',
      desc: '从APO中获取指标和链路数据',
      paramSpec: {
        name: 'root',
        type: 'object',
        children: [
          {
            name: 'promAddress',
            type: 'string',
          },
          {
            name: 'promStorage',
            type: 'string',
          },
        ],
      },
      capabilitySpec: ['RedMetric', 'ServiceInstance', 'Topology', 'Log', 'Trace'],
    },
    {
      typeName: 'Datadog',
      desc: '从Datadog获取指标和链路数据',
      paramSpec: {
        name: 'root',
        type: 'object',
        children: [
          {
            name: 'site',
            type: 'string',
          },
          {
            name: 'apiKey',
            type: 'string',
          },
          {
            name: 'appKey',
            type: 'string',
          },
          {
            name: 'env',
            type: 'string',
          },
        ],
      },
      capabilitySpec: ['RedMetric', 'ServiceInstance', 'Topology', 'Log', 'Trace'],
    },
    {
      typeName: 'Prometheus',
      paramSpec: {},
      desc: 'Prometheus',
      capabilitySpec: ['Metrics'],
    },
    {
      typeName: 'Loki',
      paramSpec: {},
      desc: 'Loki',
      capabilitySpec: ['Log'],
    },
    {
      typeName: 'Tempo',
      paramSpec: {},
      desc: 'Tempo',
      capabilitySpec: ['Trace', 'RED Metrics'],
    },
    {
      typeName: 'OpenObserve',
      paramSpec: {},
      desc: 'OpenObserve',
      capabilitySpec: ['Metrics', 'Log', 'Trace', 'RED Metrics'],
    },
    {
      typeName: 'SigNoz',
      paramSpec: {},
      desc: 'SigNoz',
      capabilitySpec: ['Metrics', 'Log', 'Trace', 'RED Metrics'],
    },
  ]
  return (
    <div className="w-full p-1 h-full overflow-y-auto">
      <div className="grid gap-2 w-full grid-cols-[repeat(auto-fill,minmax(200px,1fr))]">
        {defaultDataplane.map((option) => (
          <div key={option.typeName} onClick={() => onClick(option)}>
            <DataplaneCard info={option} />
          </div>
        ))}
      </div>
    </div>
  )
}

export default DataplaneList
