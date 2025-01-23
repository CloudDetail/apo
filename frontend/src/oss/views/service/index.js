import { useUserContext } from 'src/core/contexts/UserContext'
import ServiceTable from './ServiceTable'
import DataGroupTabs from 'src/core/components/DataGroupTabs'
import { useEffect } from 'react'
export default function ServiceView() {
  const { getUserDataGroup } = useUserContext()
  useEffect(() => {
    getUserDataGroup()
  }, [])
  return (
    <>
      <DataGroupTabs>
        {(groupId) => (
          <div style={{ height: 'calc(100vh - 120px)' }} className="overflow-hidden">
            <ServiceTable groupId={groupId} />
          </div>
        )}
      </DataGroupTabs>
    </>
  )
}
