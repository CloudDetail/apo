import React from 'react';
import { Table, Button, Popconfirm } from 'antd';
import { MdOutlineModeEdit } from 'react-icons/md';
import { RiDeleteBin5Line } from 'react-icons/ri';
import { LuShieldCheck } from 'react-icons/lu';
import { useTranslation } from 'react-i18next';
import { Role } from 'src/core/types/role';

interface RoleTableProps {
  roleList: Role[];
  onEdit: (role: Role) => void;
  onDelete: (roleId: string | number) => void;
  onConfigPermission: (role: Role) => void;
}

export const RoleTable: React.FC<RoleTableProps> = ({
  roleList,
  onEdit,
  onDelete,
  onConfigPermission,
}) => {
  const { t } = useTranslation('core/roleManage');

  const columns = [
    {
      title: t('index.roleName'),
      dataIndex: 'roleName',
      key: 'roleName',
      align: 'center',
    },
    {
      title: t('index.description'),
      dataIndex: 'description',
      key: 'description',
      align: 'center',
    },
    {
      title: t('index.operation'),
      key: 'operation',
      align: 'center',
      render: (_: any, record: Role) => (
        <>
          <Button
            onClick={() => onEdit(record)}
            icon={<MdOutlineModeEdit />}
            type="text"
            className="mr-2"
          >
            {t('index.edit')}
          </Button>
          {record.roleName !== 'admin' && (
            <Popconfirm
              title={t('index.confirmDelete', { name: record.roleName })}
              onConfirm={() => onDelete(record.roleId)}
              >
              <Button type="text" icon={<RiDeleteBin5Line />} danger className="mr-1">
                {t('index.delete')}
              </Button>
            </Popconfirm>
          )}
          <Button
            color="primary"
            variant="outlined"
            icon={<LuShieldCheck />}
            onClick={() => onConfigPermission(record)}
          >
            {t('index.configPermission')}
          </Button>
        </>
      ),
    },
  ];

  return (
    <Table
      dataSource={roleList}
      columns={columns}
      rowKey="roleId"
    />
  );
};