import { Button, Descriptions, DescriptionsProps, Popconfirm } from 'antd'
import Paragraph from 'antd/es/typography/Paragraph'

import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import { LuRefreshCcw, LuShieldCheck } from 'react-icons/lu'
import { MdOutlineAdd, MdOutlineEdit } from 'react-icons/md'
import { DataGroupPermissionInfo } from 'src/core/types/dataGroup'
import DatasourceTag from '../component/DatasourceTag'
import { RiDeleteBin5Line } from 'react-icons/ri'

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
        {info?.permissionType === 'edit' && (
          <Button
            type="text"
            size="small"
            icon={<LuRefreshCcw />}
            onClick={() => {
              // setOpenRefreshModal(true)
              // handleRefresh(false)
            }}
            className="mr-2"
          >
            {t('refresh')}
          </Button>
        )}
        {info?.permissionType !== 'known' && (
          <Button
            color="cyan"
            variant="link"
            size="small"
            icon={<MdOutlineAdd />}
            onClick={openAddModal}
            className="mr-2"
          >
            {t('add')}
          </Button>
        )}
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
      </div>
      {/* <Modal open={true} onCancel={() => {}} footer={null}>
        <div>
          <div>
            <div>123</div>
          </div>
        </div>
      </Modal> */}
    </div>
  )
}

export default DataGroupInfo
