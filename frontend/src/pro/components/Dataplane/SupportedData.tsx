import { useEffect, useState } from 'react'
import { getCapabilitiesApi } from 'src/pro/api/dataplane'
import StatusIndicator from '../StatusIndicator'
import { SettingsIcon } from 'lucide-react'

export default function SupportedData() {
  const [supportedData, setSupportedData] = useState<any[]>([])
  const getSupportedData = () => {
    getCapabilitiesApi().then((res) => {
      setSupportedData(
        res?.totalSupported?.map((item) => {
          return {
            status: item.enabled ? 'success' : 'warning',
            label: item.type,
            key: item.type,
          }
        }),
      )
    })
  }
  useEffect(() => {
    getSupportedData()
  }, [])
  return (
    <div className="items-center flex gap-2 w-full justify-between rounded-md border border-[var(--ant-color-border-secondary)] ">
      <div className="text-sm font-medium px-4 py-2 border-r border-[var(--ant-color-border-secondary)]">
        {' '}
        Data Coverage Overview
      </div>
      <div className="flex-1 justify-end flex">
        <StatusIndicator items={supportedData} width="500px" />
      </div>
      <SettingsIcon className="mr-2 w-4 h-4 flex-0 cursor-pointer" />
    </div>
  )
}
