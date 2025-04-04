/**
 * Copyright 2025 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import { Col, List, Row, Typography } from 'antd'
import { alertItems } from '../../constant'
import DatasourceItem from '../../components/DatasourceItem'
import Search from 'antd/es/input/Search'
import { useSearchParams } from 'react-router-dom'
import { useAlertIntegrationContext } from 'src/core/contexts/AlertIntegrationContext'
import { AlertKey } from 'src/core/types/alertIntegration'
import { useTranslation } from 'react-i18next'

const AlertsDatasourceList = () => {
  const { t } = useTranslation('core/alertsIntegration')
  const [searchParams, setSearchParams] = useSearchParams()
  const setConfigDrawerVisible = useAlertIntegrationContext((ctx) => ctx.setConfigDrawerVisible)
  const openAddDrawer = (sourceType: AlertKey) => {
    setSearchParams({ sourceId: 'add', sourceType: sourceType })
    setConfigDrawerVisible(true)
  }
  return (
    <div className="pr-2">
      {/* <Search  className="mb-3" /> */}
      <Typography>
        <Typography.Title level={5}>{t('addAlertsIntegration')}</Typography.Title>
      </Typography>
      <Row gutter={[0, 18]} justify="space-between">
        {alertItems?.map((item) => (
          <Col key={item.key} onClick={() => openAddDrawer(item.key)}>
            <DatasourceItem {...item} />
          </Col>
        ))}
      </Row>
    </div>
  )
}
export default AlertsDatasourceList
