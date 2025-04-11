import React, { useState, useEffect } from 'react';
import { Table, Button, Input, Flex, ConfigProvider, Popconfirm, Space, Dropdown, Form, Pagination, Divider, Select, Radio } from 'antd';
import { DownOutlined } from '@ant-design/icons';
import { MdOutlineModeEdit } from 'react-icons/md';
import { RiDeleteBin5Line } from 'react-icons/ri';
import { LuShieldCheck } from 'react-icons/lu';
import { BsPersonFillAdd } from 'react-icons/bs';
import { useTranslation } from 'react-i18next';
import { useUserContext } from 'src/core/contexts/UserContext';
import LoadingSpinner from 'src/core/components/Spinner';
import { getUserListApi, removeUserApi, createUserApi, updateCorporationApi, updateEmailApi, updatePhoneApi, updatePasswordWithNoOldPwdApi } from 'src/core/api/user';
import { getAllRolesApi, revokeUserRoleApi } from 'src/core/api/role'
import { useRoleActions } from './hooks/useRoleActions';
import { useApiParams } from 'src/core/hooks/useApiParams';
import FormModal from 'src/core/components/Modal/FormModal';
import { showToast } from 'src/core/utils/toast';
import DataGroupAuthorizeModal from 'src/core/components/PermissionAuthorize/DataGroupAuthorizeModal';
import { User } from 'src/core/types/user';
import style from 'src/core/views/UserManage/index.module.css';
import { SearchBar } from './components/SearchBar';
import { UserTable } from './components/UserTable';
import { AddUserModal } from './components/AddUserModal';
import { EditUserModal } from './components/EditUserModal';
import { useUserActions } from './hooks/useUserActions';

// 用户搜索参数类型
interface UserSearchParams {
  username?: string;
  corporation?: string;
  currentPage: number;
  pageSize: number;
}

export default function UserManage() {
  const { t } = useTranslation('core/userManage');
  const [userList, setUserList] = useState<User[]>([]);
  const [selectedUser, setSelectedUser] = useState<User | null>(null);
  const [loading, setLoading] = useState(false);
  const [currentPage, setCurrentPage] = useState(1);
  const [pageSize, setPageSize] = useState(10);
  const [total, setTotal] = useState(0);
  const [searchParams, setSearchParams] = useState<UserSearchParams>({
    username: '',
    corporation: '',
    currentPage: 1,
    pageSize: 10
  });
  const [modalStates, setModalStates] = useState({
    add: false,
    edit: false,
    authorize: false
  });

  const {
    fetchUsers,
    removeUserById,
    createNewUser,
    updateUser,
    resetPassword,
  } = useUserActions();

  const {
    loading: roleLoading,
    fetchRoles,
    roleOptions
  } = useRoleActions();

  // 获取用户列表
  const handleFetchUsers = async (params = searchParams) => {
    setLoading(true);
    try {
      const result = await fetchUsers(params);
      if (result) {
        const { users, currentPage: newPage, pageSize: newSize, total: newTotal } = result;
        setUserList(users);
        setCurrentPage(newPage);
        setPageSize(newSize);
        setTotal(newTotal);
      }
    } catch (error) {
      console.error('获取用户列表失败:', error);
    } finally {
      setLoading(false);
    }
  };

  // 处理搜索
  const handleSearch = (type: 'username' | 'corporation', value: string) => {
    const newParams = {
      ...searchParams,
      [type]: value,
      currentPage: 1
    };
    setSearchParams(newParams);
    handleFetchUsers(newParams);
  };

  // 处理分页变化
  const handlePaginationChange = (page: number, size: number) => {
    const newParams = {
      ...searchParams,
      currentPage: page,
      pageSize: size
    };
    setSearchParams(newParams);
    handleFetchUsers(newParams);
  };

  // 删除用户
  const handleRemoveUser = async (userId: string | number) => {
    try {
      await removeUserById(userId);
      if (userList.length <= 1 && currentPage > 1) {
        handleFetchUsers({ ...searchParams, currentPage: currentPage - 1 });
      } else {
        handleFetchUsers(searchParams);
      }
    } catch (error) {
      console.error('删除用户失败:', error);
    }
  };

  // 处理模态框
  const toggleModal = (modalName: keyof typeof modalStates, visible: boolean) => {
    setModalStates(prev => ({ ...prev, [modalName]: visible }));
    if (!visible) setSelectedUser(null);
  };

  useEffect(() => {
    handleFetchUsers();
    fetchRoles();
  }, []);

  return (
    <>
      <LoadingSpinner loading={loading || roleLoading} />
      <div className={style.userManageContainer}>
        <SearchBar
          username={searchParams.username}
          corporation={searchParams.corporation}
          onSearch={handleSearch}
          onAddUser={() => toggleModal('add', true)}
        />

        <ConfigProvider theme={{ components: { Table: { headerBg: '#222631' } } }}>
          <Flex vertical className="w-full flex-1 pb-4 justify-between">
            <UserTable
              userList={userList}
              loading={loading}
              onEdit={(user) => {
                setSelectedUser(user);
                toggleModal('edit', true);
              }}
              onDelete={handleRemoveUser}
              onAuthorize={(user) => {
                setSelectedUser(user);
                toggleModal('authorize', true);
              }}
            />
            <Pagination
              className="mt-4 absolute bottom-0 right-0"
              current={currentPage}
              pageSize={pageSize}
              total={total}
              pageSizeOptions={['10', '30', '50']}
              showSizeChanger
              onChange={handlePaginationChange}
              showQuickJumper
            />
          </Flex>
        </ConfigProvider>
      </div>

      <AddUserModal
        visible={modalStates.add}
        loading={loading}
        roleItems={roleOptions}
        onCancel={() => toggleModal('add', false)}
        onFinish={async (values) => {
          try {
            await createNewUser(values);
            toggleModal('add', false);
            handleFetchUsers();
          } catch (error) {
            console.error('添加用户失败:', error);
          }
        }}
      />

      <EditUserModal
        visible={modalStates.edit}
        user={selectedUser}
        roleItems={roleOptions}
        onCancel={() => toggleModal('edit', false)}
        onFinish={async (values) => {
          if (!selectedUser) return;
          try {
            await updateUser(selectedUser, values);
            toggleModal('edit', false);
            handleFetchUsers();
          } catch (error) {
            console.error('更新用户失败:', error);
          }
        }}
        onResetPassword={async (values) => {
          if (!selectedUser) return;
          try {
            await resetPassword(selectedUser.userId, values);
            toggleModal('edit', false);
            handleFetchUsers();
          } catch (error) {
            console.error('重置密码失败:', error);
          }
        }}
      />

      <DataGroupAuthorizeModal
        open={modalStates.authorize}
        closeModal={() => toggleModal('authorize', false)}
        subjectId={selectedUser?.userId}
        subjectName={selectedUser?.username}
        type="user"
        refresh={() => toggleModal('authorize', false)}
      />
    </>
  );
}