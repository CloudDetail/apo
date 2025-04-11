import React from 'react';
import { Table, Button, Popconfirm } from 'antd';
import { MdOutlineModeEdit } from 'react-icons/md';
import { RiDeleteBin5Line } from 'react-icons/ri';
import { LuShieldCheck } from 'react-icons/lu';
import { useTranslation } from 'react-i18next';
import { User } from 'src/core/types/user';

interface UserTableProps {
  userList: User[];
  loading: boolean;
  onEdit: (user: User) => void;
  onDelete: (userId: string | number) => Promise<void>;
  onAuthorize: (user: User) => void;
}

export const UserTable: React.FC<UserTableProps> = ({
  userList,
  loading,
  onEdit,
  onDelete,
  onAuthorize,
}) => {
  const { t } = useTranslation('core/userManage');

  const columns = [
    {
      title: t('index.userName'),
      dataIndex: 'username',
      key: 'username',
      align: 'center',
    },
    {
      title: t('index.role'),
      dataIndex: 'role',
      key: 'role',
      align: 'center',
    },
    {
      title: t('index.corporation'),
      dataIndex: 'corporation',
      key: 'corporation',
      align: 'center',
    },
    {
      title: t('index.phone'),
      dataIndex: 'phone',
      key: 'phone',
      align: 'center',
    },
    {
      title: t('index.email'),
      dataIndex: 'email',
      key: 'email',
      align: 'center',
    },
    {
      title: t('index.operation'),
      dataIndex: 'userId',
      key: 'userId',
      align: 'center',
      render: (userId: string | number, record: User) => {
        const { username } = record;
        return username !== 'admin' && (
          <>
            <Button
              onClick={() => onEdit(record)}
              icon={<MdOutlineModeEdit />}
              type="text"
              className="mr-1"
            >
              {t('index.edit')}
            </Button>
            <Popconfirm
              title={t('index.confirmDelete', { name: username })}
              onConfirm={() => onDelete(userId)}
            >
              <Button type="text" icon={<RiDeleteBin5Line />} danger className="mr-1">
                {t('index.delete')}
              </Button>
            </Popconfirm>
            <Button
              color="primary"
              variant="outlined"
              icon={<LuShieldCheck />}
              onClick={() => onAuthorize(record)}
            >
              {t('index.dataGroup')}
            </Button>
          </>
        )
      },
      width: 400,
    },
  ];

  return (
    <Table
      dataSource={userList}
      columns={columns}
      pagination={false}
      loading={loading}
      scroll={{ y: 'calc(100vh - 220px)' }}
    />
  );
};