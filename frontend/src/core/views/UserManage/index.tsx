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
import { useRoles } from 'src/core/hooks/useRoles';
import { useApiParams } from 'src/core/hooks/useApiParams';
import FormModal from 'src/core/components/Modal/FormModal';
import { showToast } from 'src/core/utils/toast';
import DataGroupAuthorizeModal from 'src/core/components/PermissionAuthorize/DataGroupAuthorizeModal';
import { User } from 'src/core/types/user';
import style from 'src/core/views/UserManage/index.module.css';

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
  const [updateLoading, setUpdateLoading] = useState(false);
  const [modalAddVisibility, setModalAddVisibility] = useState(false);
  const [roleEditVisibility, setRoleModalVisibility] = useState(false);
  const [modalEditVisibility, setModalEditVisibility] = useState(false);
  const [authorizeModalVisibility, setAuthorizeModalVisibility] = useState(false);
  const { user } = useUserContext();

  // 使用 useApiParams 钩子
  const { sendRequest: getUserList } = useApiParams(getUserListApi);
  const { sendRequest: revokeUserRole } = useApiParams(revokeUserRoleApi)
  const { sendRequest: removeUser } = useApiParams(removeUserApi);
  const { sendRequest: createUser, loading: addLoading } = useApiParams(createUserApi);
  const { sendRequest: updateUserCorporation } = useApiParams(updateCorporationApi);
  const { sendRequest: updateUserEmail } = useApiParams(updateEmailApi);
  const { sendRequest: updateUserPhone } = useApiParams(updatePhoneApi);
  const { sendRequest: resetUerPassword } = useApiParams(updatePasswordWithNoOldPwdApi);

  const { roleList, selectedRole, selectRole } = useRoles();

  // 获取用户列表
  const fetchUsers = async (page = currentPage, size = pageSize, search = searchParams) => {
    if (loading) return;

    setLoading(true);
    const params = {
      currentPage: page,
      pageSize: size,
      ...search
    };

    try {
      const result = await getUserList(params, { useURLSearchParams: false });

      if (result) {
        const { users, currentPage: newPage, pageSize: newSize, total: newTotal } = result;

        const formattedUsers = users.map((user: User) => ({
          ...user,
          role: user.roleList?.[0]?.roleName
        }));

        setUserList(formattedUsers);
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

  // 删除用户
  const handleRemoveUser = async (userId: string | number) => {
    await removeUser(
      { userId },
      {
        useURLSearchParams: false,
        onSuccess: () => {
          showToast({
            title: t('index.deleteSuccess'),
            color: 'success',
          });

          // 重新获取用户列表
          if (userList.length <= 1) {
            fetchUsers(1);
          } else {
            fetchUsers();
          }
        },
        onError: (error) => {
          console.error('删除用户失败:', error);
        }
      }
    );
  };

  // 改变用户角色
  const handleRevokeUserRole = async (userId: string | number, roleId: string | number) => {
    await revokeUserRole(
      {
        userId,
        roleList: [roleId]
      },
      {
        onSuccess: () => {
          showToast({
            title: t('index.revokeSuccess'),
            color: 'success',
          });

          // 重新获取用户列表
          fetchUsers();
        },
        onError: (error) => {
          console.error('撤销用户角色失败:', error);
        }
      }
    );
  };

  // 添加用户
  const handleAddUser = async (values: any) => {
    await createUser(
      values,
      {
        onSuccess: () => {
          showToast({
            title: t('index.addSuccess'),
            color: 'success',
          });
          setModalAddVisibility(false);
          fetchUsers();
        },
        onError: (error) => {
          console.error('添加用户失败:', error);
        }
      }
    );
  };

  // 更新密码
  const handleResetPassword = async (values: any) => {
    if (!selectedUser) return;

    await resetUerPassword({
      userId: selectedUser.userId,
      newPassword: values.newPassword,
      confirmPassword: values.confirmPassword
    });

    showToast({
      title: t('index.updateSuccess'),
      color: 'success',
    });
    setModalEditVisibility(false);
    fetchUsers();
  }

  // 更新用户
  const handleEditUser = async (values: any) => {
    console.log('handleEditUser')
    if (!selectedUser) return;

    console.log('handleEditUser123')

    try {
      // 更新公司信息
      if (values.corporation !== selectedUser.corporation) {
        await updateUserCorporation({
          userId: selectedUser.userId,
          corporation: values.corporation
        });
      }

      // 更新邮箱
      if (values.email !== selectedUser.email) {
        await updateUserEmail({
          username: selectedUser.username,
          email: values.email
        });
      }

      // 更新电话
      if (values.phone !== selectedUser.phone) {
        await updateUserPhone({
          username: selectedUser.username,
          phone: values.phone
        });
      }

      showToast({
        title: t('index.updateSuccess'),
        color: 'success',
      });
      setModalEditVisibility(false);
      fetchUsers();
    } catch (error) {
      console.error('更新用户失败:', error);
    }
  };

  // 角色下拉菜单项
  const roleItems = roleList.map((role) => ({
    label: role.roleName,
    key: role.roleId,
    value: role.roleId
  }));

  // 初始化加载
  useEffect(() => {
    fetchUsers();
  }, []);

  // 处理分页变化
  const handlePaginationChange = (page: number, size: number) => {
    setCurrentPage(page);
    setPageSize(size);
    fetchUsers(page, size);
  };

  // 用户名搜索变更
  const handleUsernameChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const username = e.target.value;
    const newParams = { ...searchParams, username };
    setSearchParams(newParams);
    fetchUsers(1, pageSize, newParams);
  };

  // 公司搜索变更
  const handleCorporationChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const corporation = e.target.value;
    const newParams = { ...searchParams, corporation };
    setSearchParams(newParams);
    fetchUsers(1, pageSize, newParams);
  };

  // 关闭授权模态框
  const closeAuthorizeModal = () => {
    setAuthorizeModalVisibility(false);
    setSelectedUser(null);
  };

  // 刷新数据
  const refresh = () => {
    fetchUsers();
    closeAuthorizeModal();
  };

  // 用户列表列定义
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
              onClick={() => {
                setSelectedUser(record);
                setModalEditVisibility(true);
              }}
              icon={<MdOutlineModeEdit />}
              type="text"
              className="mr-1"
            >
              {t('index.edit')}
            </Button>
            <Popconfirm
              title={t('index.confirmDelete', { name: username })}
              onConfirm={() => handleRemoveUser(userId)}
            >
              <Button type="text" icon={<RiDeleteBin5Line />} danger className="mr-1">
                {t('index.delete')}
              </Button>
            </Popconfirm>
            <Button
              color="primary"
              variant="outlined"
              icon={<LuShieldCheck />}
              onClick={() => {
                setAuthorizeModalVisibility(true);
                setSelectedUser(record);
              }}
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
    <>
      <LoadingSpinner loading={loading} />
      <div className={style.userManageContainer}>
        <Flex className="mb-3 h-[40px]">
          <Flex className="w-full justify-between">
            <Flex className="w-full">
              <Flex className="w-auto items-center justify-start mr-5">
                <p className="text-md mr-2">{t('index.userName')}：</p>
                <Input
                  placeholder={t('index.search')}
                  className="w-2/3"
                  value={searchParams.username}
                  onChange={handleUsernameChange}
                />
              </Flex>
              <Flex className="w-auto items-center justify-start">
                <p className="text-md mr-2">{t('index.corporation')}：</p>
                <Input
                  placeholder={t('index.search')}
                  className="w-2/3"
                  value={searchParams.corporation}
                  onChange={handleCorporationChange}
                />
              </Flex>
            </Flex>
            <Flex className="w-full justify-end items-center">
              <Button
                type="primary"
                icon={<BsPersonFillAdd size={20} />}
                onClick={() => setModalAddVisibility(true)}
                className="flex-grow-0 flex-shrink-0"
              >
                <span className="text-xs">{t('index.addUser')}</span>
              </Button>
            </Flex>
          </Flex>
        </Flex>

        <ConfigProvider
          theme={{
            components: {
              Table: {
                headerBg: '#222631',
              },
            },
          }}
        >
          <Flex vertical className="w-full flex-1 pb-4 justify-between">
            <Table
              dataSource={userList}
              columns={columns}
              pagination={false}
              loading={loading}
              scroll={{ y: 'calc(100vh - 220px)' }}
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

      {/* 添加用户模态框 */}
      <FormModal
        title={t('index.addUser')}
        open={modalAddVisibility}
        onCancel={() => setModalAddVisibility(false)}
        confirmLoading={addLoading}
        footer={null}
      >
        <FormModal.Section
          onFinish={handleAddUser}
        >
        <Form.Item
          name="username"
          label={t('index.userName')}
          rules={[{ required: true, message: t('index.userNameRequired') }]}
        >
          <Input />
        </Form.Item>
        <Form.Item
          name="password"
          label={t('addModal.password')}
          rules={[{ required: true, message: t('index.passwordRequired') }]}
        >
          <Input.Password />
        </Form.Item>
        <Form.Item
          label={t('editModal.confirmPassword')}
          name="confirmPassword"
          rules={[
            { required: true, message: t('editModal.confirmPasswordRequired') },
            ({ getFieldValue }) => ({
              validator(_, value) {
                if (!value || getFieldValue('password') === value) {
                  return Promise.resolve()
                }
                return Promise.reject(new Error(t('editModal.confirmPasswordMismatch')))
              },
            }),
          ]}
        >
            <Input.Password placeholder={t('editModal.confirmPasswordPlaceholder')} />
        </Form.Item>
        <Form.Item
          name="corporation"
          label={t('index.corporation')}
        >
          <Input />
        </Form.Item>
        <Form.Item
          name="email"
          label={t('index.email')}
          rules={[{ type: 'email', message: t('index.emailInvalid') }]}
        >
          <Input />
        </Form.Item>
        <Form.Item
          name="phone"
          label={t('index.phone')}
        >
          <Input />
        </Form.Item>
        <Form.Item
          name="roleId"
          label={t('index.role')}
          rules={[{ required: true, message: t('addModal.roleRequired') }]}
        >
          <Select
            placeholder={t('addModal.selectRole')}
            options={roleItems}
          />
        </Form.Item>
        </FormModal.Section>
      </FormModal>

      {/* 编辑用户模态框 */}
      <FormModal
        title={t('index.editUser')}
        open={modalEditVisibility}
        onCancel={() => setModalEditVisibility(false)}
        footer={null}
      >
        <FormModal.Section
          onFinish={handleEditUser}
          initialValues={selectedUser}
        >
          <Flex gap={16} className="mb-6">
            <Form.Item
              name="username"
              label={t('index.userName')}
              rules={[{ required: true, message: t('index.userNameRequired') }]}
              style={{ marginBottom: 0, flex: 1 }}
            >
              <Input disabled />
            </Form.Item>
            <Form.Item
              name="role"
              label='角色'
              rules={[{ required: true, message: '请选择角色' }]}
              style={{ marginBottom: 0, flex: 1 }}
            >
              <Select
                options={roleItems}
                onChange={(value) => handleRevokeUserRole(selectedUser?.userId || '', value)}
              />
            </Form.Item>
          </Flex>
          <Form.Item
            name="corporation"
            label={t('index.corporation')}
          >
            <Input />
          </Form.Item>
          <Form.Item
            name="email"
            label={t('index.email')}
            rules={[{ type: 'email', message: t('index.emailInvalid') }]}
          >
            <Input placeholder={t('editModal.emailPlaceholder')} />
          </Form.Item>
          <Form.Item
            name="phone"
            label={t('index.phone')}
            rules={[{ pattern: /^1[3-9]\d{9}$/, message: t('editModal.phoneInvalid') }]}
          >
            <Input placeholder={t('editModal.phonePlaceholder')} />
          </Form.Item>
        </FormModal.Section>
        <Divider />
        <FormModal.Section
          onFinish={handleResetPassword}
        >
          <Form.Item
            label={t('editModal.newPassword')}
            name="newPassword"
            rules={[
              {
                required: true,
                message: t('editModal.newPasswordPlaceholder'),
              },
              {
                pattern: /^(?=.*[a-z])(?=.*[A-Z])(?=.*[!@#$%^&*()\-_+=<>?/{}[\]|:;.,~]).{9,}$/,
                message: t('editModal.newPasswordPattern'),
              },
            ]}
          >
            <Input.Password placeholder={t('editModal.newPasswordPlaceholder')} />
          </Form.Item>
          <Form.Item
            label={t('editModal.confirmPassword')}
            name="confirmPassword"
            rules={[
              { required: true, message: t('editModal.confirmPasswordRequired') },
              ({ getFieldValue }) => ({
                validator(_, value) {
                  if (!value || getFieldValue('newPassword') === value) {
                    return Promise.resolve()
                  }
                  return Promise.reject(new Error(t('editModal.confirmPasswordMismatch')))
                },
              }),
            ]}
          >
            <Input.Password placeholder={t('editModal.confirmPasswordPlaceholder')} />
          </Form.Item>
        </FormModal.Section>
      </FormModal>

      {/* 编辑权限模态框 */}
      <FormModal
        title= {`设置 大雄 的角色`}
        open={roleEditVisibility}
        onCancel={() => setRoleModalVisibility(false)}
        footer={null}
      >
        <FormModal.Section
          onFinish={handleRevokeUserRole}
        >
          <Form.Item name='role'>
          <Radio.Group
            className="flex flex-col gap-2 m-2"
            onChange={() => {}}
            // value={1}
            options={[
              { value: 1, label: 'admin' },
              { value: 2, label: 'manager' },
              { value: 3, label: 'viewer' },
            ]}
          />
          </Form.Item>
        </FormModal.Section>
      </FormModal>

      {/* 数据组授权模态框 */}
      <DataGroupAuthorizeModal
        open={authorizeModalVisibility}
        closeModal={closeAuthorizeModal}
        subjectId={selectedUser?.userId}
        subjectName={selectedUser?.username}
        type="user"
        refresh={refresh}
      />
    </>
  );
}