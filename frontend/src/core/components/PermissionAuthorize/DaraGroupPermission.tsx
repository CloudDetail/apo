import { Alert, Button, Table, Tag, Tooltip } from 'antd'
import { useEffect, useState } from 'react'
import styles from './index.module.scss'
import { getDataGroupsApi } from 'src/core/api/dataGroup'
import Paragraph from 'antd/es/typography/Paragraph'
import DatasourceTag from 'src/core/views/DataGroup/component/DatasourceTag'
import { SaveDataGroupParams } from 'src/core/types/dataGroup'
import { AiOutlineCloseCircle } from 'react-icons/ai'
import { useTranslation } from 'react-i18next'

interface DataGroupPermissionProps {
  id: string
  dataGroupList: any[]
  onChange: any
  type: 'team' | 'user'
  permissionSourceTeam: any[]
}

const DataGroupPermission = (props: DataGroupPermissionProps) => {
  const { t } = useTranslation('core/permission')
  const { id, dataGroupList = [], onChange, type, permissionSourceTeam } = props
  const [checkedKeys, setcheckedKeys] = useState([])
  const [data, setData] = useState<SaveDataGroupParams[]>([])
  const columns = [
    {
      title: 'groupId',
      dataIndex: 'groupId',
      key: 'groupId',
      hidden: true,
    },
    {
      title: t('dataGroupName'),
      dataIndex: 'groupName',
      width: 200,
      key: 'groupName',
    },
    {
      title: t('dataGroupDes'),
      width: 200,
      dataIndex: 'description',
      key: 'description',
    },
    {
      title: t('datasource'),
      dataIndex: 'datasourceList',
      key: 'datasourceList',
      render: (value) => {
        return (
          <Paragraph
            className="m-0"
            ellipsis={{
              expandable: true,
              rows: 3,
            }}
          >
            {value?.map((item) => <DatasourceTag type={item.type} datasource={item.datasource} />)}
          </Paragraph>
        )
      },
    },
  ]
  const getDataGroups = () => {
    getDataGroupsApi({
      currentPage: 1,
      pageSize: 1000,
    }).then((res) => {
      setData(res.dataGroupList)
    })
  }
  useEffect(() => {
    getDataGroups()
  }, [])
  const onSelectAll = (selected, selectedRows, changeRows) => {
    if (selected) {
      onChange([...dataGroupList, ...changeRows])
    } else {
      onChange([])
    }
  }
  const rowSelection = {
    onSelect: (record: SaveDataGroupParams) => {
      onSelectTableRow(record)
    },
    onSelectAll: onSelectAll,
    selectedRowKeys: checkedKeys,
  }
  const onSelectTableRow = ({ groupId, groupName }: SaveDataGroupParams, event?: any) => {
    if (event) {
      event.stopPropagation()
    }
    if (checkedKeys.includes(groupId)) {
      onChange(dataGroupList.filter((item) => !(item.groupId === groupId)))
    } else {
      onChange([
        ...dataGroupList,
        {
          groupId: groupId,
          groupName: groupName,
        },
      ])
    }
  }
  const deleteDataGroup = (e, groupId: string) => {
    e.preventDefault()
    const result = dataGroupList.filter((item) => item.groupId !== groupId)
    onChange(result)
  }
  useEffect(() => {
    setcheckedKeys(dataGroupList.map((group) => group.groupId))
  }, [dataGroupList])

  return (
    <div style={{ maxHeight: '60vh' }} className="flex flex-col" id={id}>
      <div className={styles.tagContainer}>
        <div className="flex-1">
          {dataGroupList.map((item) => (
            <Tag
              closable
              onClose={(e) => {
                deleteDataGroup(e, item.groupId)
              }}
              color="success"
            >
              {item.groupName}
            </Tag>
          ))}
          {/* {permissionSourceTeam?.map((item) => (
            <Tooltip title={t('tagTooltip')}>
              <Tag
                onClose={(e) => {
                  deleteDataGroup(e, item.groupId)
                }}
              >
                {item.groupName}
              </Tag>
            </Tooltip>
          ))} */}
        </div>

        {dataGroupList?.length > 0 && (
          <Button
            size="small"
            type="text"
            icon={<AiOutlineCloseCircle />}
            className="absolute right-2 flex-grow-0 flex-shrink-0"
            onClick={() => {
              onSelectAll(false, [], [])
            }}
          ></Button>
        )}
      </div>
      {/* <Alert
        message={t('dataGroupAlert')}
        type="warning"
        showIcon
        className="text-xs mb-2 mx-0"
      ></Alert> */}
      <Table<SaveDataGroupParams>
        rowSelection={{ type: 'checkbox', ...rowSelection }}
        dataSource={data}
        columns={columns}
        pagination={false}
        className="overflow-auto"
        scroll={{ y: 310 }}
        size="small"
        rowKey="groupId"
        onRow={(record) => {
          return {
            onClick: (event) => {
              onSelectTableRow(record, event)
            },
          }
        }}
      ></Table>
    </div>
  )
}
export default DataGroupPermission
