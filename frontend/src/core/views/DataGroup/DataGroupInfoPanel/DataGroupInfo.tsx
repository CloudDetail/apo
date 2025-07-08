import { Button, Descriptions, DescriptionsProps, Modal, Popconfirm } from 'antd'
import Paragraph from 'antd/es/typography/Paragraph'

import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import { LuShieldCheck } from 'react-icons/lu'
import { MdOutlineAdd, MdOutlineEdit } from 'react-icons/md'
import { DataGroupPermissionInfo } from 'src/core/types/dataGroup'
import DatasourceTag from '../component/DatasourceTag'
import { RiDeleteBin5Line } from 'react-icons/ri'
import { refreshGroupDatasourceApiV2 } from 'src/core/api/dataGroup'
import { notify } from 'src/core/utils/notify'
const DataGroupInfo = ({
  info,
  datasources,
  openAddModal,
  handlePermission,
  openEditModal,
  deleteDataGroup,
}: {
  info: DataGroupPermissionInfo
  datasources: any[]
  openAddModal: () => void
  handlePermission: () => void
  openEditModal: (info: DataGroupPermissionInfo) => void
  deleteDataGroup: (info: DataGroupPermissionInfo) => void
}) => {
  const { t } = useTranslation('core/dataGroup')
  const { t: ct } = useTranslation('common')
  const [openRefreshModal, setOpenRefreshModal] = useState<boolean>(false)
  const [cleanList, setCleanList] = useState<any[]>([])
  const [protectedList, setProtectedList] = useState<any[]>([])

  const items: DescriptionsProps['items'] = [
    {
      key: '1',
      span: 3,
      label: t('dataGroupDes'),
      children: info?.description,
    },
    {
      key: '2',
      label: t('datasource'),
      span: 3,
      children: (
        <Paragraph
          className="m-0 items-center flex flex-wrap w-full"
          ellipsis={{
            expandable: true,
            rows: 1,
          }}
        >
          {datasources
            ?.sort((a, b) => {
              const typeOrder = ['system', 'cluster', 'namespace', 'service']
              const aIndex = typeOrder.indexOf(a.type)
              const bIndex = typeOrder.indexOf(b.type)
              return aIndex - bIndex
            })
            ?.map((item) => <DatasourceTag key={item.id} {...item} block={false} />)}
        </Paragraph>
      ),
    },
  ]
  const handleRefresh = (clean: boolean) => {
    refreshGroupDatasourceApiV2(info.groupId, clean).then((res: any) => {
      if (!clean && (res?.toBeDeleted?.length || 0) + (res?.protected?.length || 0) > 0) {
        setCleanList(res?.toBeDeleted || [])
        setProtectedList(res?.protected || [])
        setOpenRefreshModal(true)
      } else if (!clean) {
        notify({
          type: 'info',
          message: t('noInvalidDatasource'),
        })
      } else if (clean) {
        notify({
          type: 'success',
          message: t('clearSuccess'),
        })
      }
    })
  }
  return (
    <div className="flex flex-col justify-between h-full">
      <Descriptions
        title={info?.groupName}
        items={items}
        size="small"
        styles={{ label: { alignItems: 'center' } }}
        classNames={{ label: 'flex items-center' }}
      />
      <div className="w-full text-right mb-2 pr-2">
        {/* {info?.permissionType === 'edit' && (
          <Button
            type="text"
            size="small"
            icon={<LuRefreshCcw />}
            onClick={() => {
              handleRefresh(false)
            }}
            className="mr-2"
          >
            {t('clearInvalidDatasource')}
          </Button>
        )} */}

        {info?.permissionType === 'edit' && (
          <>
            <Button
              type="text"
              size="small"
              onClick={() => openEditModal(info)}
              icon={
                <MdOutlineEdit className="!text-[var(--ant-color-primary-text)] !hover:text-[var(--ant-color-primary-text-active)]" />
              }
            >
              <span className="text-[var(--ant-color-primary-text)] hover:text-[var(--ant-color-primary-text-active)]">
                {t('edit')}
              </span>
            </Button>
            <Popconfirm
              title={t('confirmDelete', {
                groupName: info.groupName,
              })}
              onConfirm={() => deleteDataGroup(info)}
              okText={ct('confirm')}
              cancelText={ct('cancel')}
            >
              <Button type="text" size="small" icon={<RiDeleteBin5Line />} danger>
                {ct('delete')}
              </Button>
            </Popconfirm>
            <Button
              size="small"
              color="primary"
              variant="outlined"
              icon={<LuShieldCheck />}
              onClick={handlePermission}
            >
              {t('authorize')}
            </Button>
          </>
        )}
        {info?.permissionType !== 'known' && (
          <Button
            type="link"
            size="small"
            icon={<MdOutlineAdd />}
            onClick={openAddModal}
            className="mr-2"
          >
            {t('addSubGroup')}
          </Button>
        )}
      </div>
      <Modal
        width={'50%'}
        open={openRefreshModal}
        onCancel={() => {
          setOpenRefreshModal(false)
        }}
        onOk={() => {
          if (cleanList?.length > 0) {
            handleRefresh(true)
          } else {
            setOpenRefreshModal(false)
          }
        }}
        title={t('clearInvalidDatasource')}
      >
        <div>
          {cleanList?.length > 0 && (
            <div>
              {t('toBeCleanedData')}
              <div>
                {cleanList
                  ?.sort((a, b) => {
                    const typeOrder = ['system', 'cluster', 'namespace', 'service']
                    const aIndex = typeOrder.indexOf(a.type)
                    const bIndex = typeOrder.indexOf(b.type)
                    return aIndex - bIndex
                  })
                  ?.map((item) => <DatasourceTag key={item.id} {...item} block={false} />)}
              </div>
            </div>
          )}

          {protectedList?.length > 0 && (
            <div>
              {t('protectedDataSources')}
              <div>
                {protectedList
                  ?.sort((a, b) => {
                    const typeOrder = ['system', 'cluster', 'namespace', 'service']
                    const aIndex = typeOrder.indexOf(a.type)
                    const bIndex = typeOrder.indexOf(b.type)
                    return aIndex - bIndex
                  })
                  ?.map((item) => <DatasourceTag key={item.id} {...item} block={false} />)}
              </div>
            </div>
          )}
        </div>
      </Modal>
    </div>
  )
}

export default DataGroupInfo
