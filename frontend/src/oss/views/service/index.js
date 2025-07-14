/**
 * Copyright 2025 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import ServiceTable from './ServiceTable'
import { ChartsProvider } from 'src/core/contexts/ChartsContext'
export default function ServiceView() {
  return (
    <>
      {/* <DataGroupTabs>
        {(groupId, height) => (
          <div className="overflow-hidden">
            <ChartsProvider>
              <ServiceTable groupId={groupId} height={height} />
            </ChartsProvider>
          </div>
        )}
      </DataGroupTabs> */}
      <div className="overflow-hidden">
        <ChartsProvider>
          <ServiceTable />
        </ChartsProvider>
      </div>
    </>
  )
}
