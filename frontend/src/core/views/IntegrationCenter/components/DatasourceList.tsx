/**
 * Copyright 2025 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import { Col, Row, Image, Radio, RadioChangeEvent } from 'antd'
import { useEffect, useState } from 'react'
import { DatasourceItemData, IntegrationType } from '../types'
import { useIntergrationContext } from 'src/core/contexts/IntergrationContext'
import styles from './datasourceList.module.scss'
import DatasourceItem from './DatasourceItem'
interface LogoListProps {
  list: DatasourceItemData[]
  type: IntegrationType
}

const DatasourceList = ({ list, type }: LogoListProps) => {
  const openConfigDrawer = useIntergrationContext((ctx) => ctx.openConfigDrawer)
  const [value, setValue] = useState()

  const clickRadio = (key) => {
    openConfigDrawer(type, key)
    if (!value) {
      // onClick()
    }
    setValue(key)
  }
  return (
    <>
      <div className={styles.container}>
        {list.map((item, index) => (
          <div className={styles.item} key={index} onClick={() => clickRadio(item.key)}>
            <DatasourceItem {...item} />
            <Radio checked={value === item.key} className="absolute right-3 top-3"></Radio>
          </div>
        ))}
      </div>
    </>
  )
}
export default DatasourceList
