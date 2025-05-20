/**
 * Copyright 2025 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import React, { useState, useEffect } from 'react';
import { Flex, Pagination } from 'antd';
import LoadingSpinner from 'src/core/components/Spinner';
import { useRoleActions } from './hooks/useRoleActions';
import DataGroupAuthorizeModal from 'src/core/components/PermissionAuthorize/DataGroupAuthorizeModal';
import { User } from 'src/core/types/user';
import style from 'src/core/views/UserManage/index.module.css';
import { SearchBar } from './components/SearchBar';
import { UserTable } from './components/UserTable';
import { AddUserModal } from './components/AddUserModal';
import { EditUserModal } from './components/EditUserModal';
import { useUserActions } from './hooks/useUserActions';
import { BasicCard } from 'src/core/components/Card/BasicCard';

interface UserSearchParams {
  username?: string;
  corporation?: string;
  currentPage: number;
  pageSize: number;
}

export default function UserManage() {
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

  // Get user list
  const handleFetchUsers = async (params = searchParams) => {
    setLoading(true);
    try {
      const result = await fetchUsers(params);
      if (result) {
        const { users, currentPage: newPage, pageSize: newSize, total: newTotal } = result;
        const usersReady = users.map((user: User) => ({
          ...user,
          key: user.userId
        }));
        setUserList(usersReady);
        setCurrentPage(newPage);
        setPageSize(newSize);
        setTotal(newTotal);
      }
    } catch (error) {
      console.error('Error fetch user list:', error);
    } finally {
      setLoading(false);
    }
  };

  const handleSearch = (type: 'username' | 'corporation', value: string) => {
    const newParams = {
      ...searchParams,
      [type]: value,
      currentPage: 1
    };
    setSearchParams(newParams);
    handleFetchUsers(newParams);
  };

  const handlePaginationChange = (page: number, size: number) => {
    const newParams = {
      ...searchParams,
      currentPage: page,
      pageSize: size
    };
    setSearchParams(newParams);
    handleFetchUsers(newParams);
  };

  const handleRemoveUser = async (userId: string | number) => {
    try {
      await removeUserById(userId);
      if (userList.length <= 1 && currentPage > 1) {
        handleFetchUsers({ ...searchParams, currentPage: currentPage - 1 });
      } else {
        handleFetchUsers(searchParams);
      }
    } catch (error) {
      console.error('Error delete user:', error);
    }
  };

  const toggleModal = (modalName: keyof typeof modalStates, visible: boolean) => {
    setModalStates(prev => ({ ...prev, [modalName]: visible }));
    if (!visible) setSelectedUser(null);
  };

  useEffect(() => {
    handleFetchUsers();
    fetchRoles();
  }, []);

  return (
    <BasicCard>
      <LoadingSpinner loading={loading || roleLoading} />

      <BasicCard.Header>
        <SearchBar
          username={searchParams.username}
          corporation={searchParams.corporation}
          onSearch={handleSearch}
          onAddUser={() => toggleModal('add', true)}
        />
      </BasicCard.Header>

      <BasicCard.Table>
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
          className="mt-4 absolute bottom-5 right-0"
          current={currentPage}
          pageSize={pageSize}
          total={total}
          onChange={handlePaginationChange}
        />
      </BasicCard.Table>

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
            console.error('Error update user:', error);
          }
        }}
        onResetPassword={async (values) => {
          if (!selectedUser) return;
          try {
            await resetPassword(selectedUser.userId, values);
            toggleModal('edit', false);
            handleFetchUsers();
          } catch (error) {
            console.error('Error reset password:', error);
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
    </BasicCard>
  );
}