/**
 * Copyright 2025 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */
import { FiDatabase } from 'react-icons/fi'
import { LuServer } from 'react-icons/lu'
import { PiComputerTowerBold } from 'react-icons/pi'
import { VscSymbolNamespace } from 'react-icons/vsc'
import { DatasourceType } from 'src/core/types/dataGroup'

const DatasourceIcon = ({ type }: { type: DatasourceType }) => {
  return (
    <span>
      {type === 'cluster' ? (
        <FiDatabase className="text-blue-400" size={14} />
      ) : type === 'namespace' ? (
        <VscSymbolNamespace className="text-lime-400" size={16} />
      ) : type === 'service' ? (
        <LuServer className="text-emerald-400" size={16} />
      ) : (
        <PiComputerTowerBold className="text-sky-400" size={16} />
      )}
    </span>
  )
}

export default DatasourceIcon
