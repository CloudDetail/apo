/**
 * Copyright 2025 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import { useUserContext } from 'src/core/contexts/UserContext'
import { Button, Tabs } from 'antd'
import { AiOutlineSetting } from 'react-icons/ai'
import { useNavigate } from 'react-router-dom'
import { useTranslation } from 'react-i18next'

export default function DataGroupTabs({ children }) {
  const { t } = useTranslation('core/dataGroup')
  //@ts-ignore
  const headerHeight = import.meta.env.VITE_APP_CODE_VERSION === 'CE' ? 'var(--ce-app-head-height)' : 'var(--ee-app-head-height)';
  const { dataGroupList } = useUserContext()
  const navigate = useNavigate()
  const getTabItems = () => {
    return dataGroupList.map((dataGroup) => ({
      label: dataGroup.groupName,
      key: dataGroup.groupId,
      closable: false,
      children: children(dataGroup.groupId, `calc(100vh - ${headerHeight} - var(--service-page-tab-list-height))`),
    }))
  }
  return (
    <>
      {dataGroupList && dataGroupList.length > 0 ? (
        <Tabs
          defaultActiveKey="1"
          items={getTabItems()}
          animated={true}
          tabBarExtraContent={
            <Button
              color="primary"
              variant="outlined"
              className="ml-3"
              icon={<AiOutlineSetting />}
              onClick={() => {
                navigate('/system/data-group')
              }}
            >
              {t('groupManage')}
            </Button>
          }
        />
      ) : (
        <>{children()}</>
      )}
    </>
  )
}
