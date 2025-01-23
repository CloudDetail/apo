import ServiceTable from './ServiceTable'
import DataGroupTabs from 'src/core/components/DataGroupTabs'

export default function ServiceView() {
  return (
    <>
      <DataGroupTabs>{(groupId) => <ServiceTable groupId={groupId} />}</DataGroupTabs>
    </>
  )
}
