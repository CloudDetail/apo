/**
 * Copyright 2025 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import { useUserContext } from 'src/core/contexts/UserContext'
import ServiceTable from './ServiceTable'
import DataGroupTabs from 'src/core/components/DataGroupTabs'
import { useEffect } from 'react'
import { ChartsProvider } from 'src/core/contexts/ChartsContext'
export default function ServiceView() {
  const { getUserDataGroup } = useUserContext()
  useEffect(() => {
    getUserDataGroup()
  }, [])
  return (
    <>
      <DataGroupTabs>
        {(groupId, height) => (
          <div className="overflow-hidden">
            <ChartsProvider>
              <ServiceTable groupId={groupId} height={height} />
            </ChartsProvider>
          </div>
        )}
      </DataGroupTabs>
    </>
  )
}
