import React, { useState, useEffect } from 'react';
import { Table, Button, Popconfirm, Form, Input } from 'antd';
import { MdOutlineModeEdit } from 'react-icons/md';
import { RiDeleteBin5Line } from 'react-icons/ri';
import { useTranslation } from 'react-i18next';
import LoadingSpinner from 'src/core/components/Spinner';
import { getAllRolesApi, createRoleApi, updateRoleApi, deleteRoleApi } from 'src/core/api/role';
import FormModal from 'src/core/components/Modal/FormModal';
import CommonModal from 'src/core/components/Modal/CommonModal';
import PermissionTree from 'src/core/components/PermissionTree';
import { useApiParams } from 'src/core/hooks/useApiParams';
import { showToast } from 'src/core/utils/toast';
import { Role } from 'src/core/types/role';
import { LuShieldCheck } from 'react-icons/lu';

export default function RoleManage() {
  const { t } = useTranslation('core/roleManage');
  // Todo: 改成使用useRole的钩子
  const [roleList, setRoleList] = useState<Role[]>([]);
  const [selectedRole, setSelectedRole] = useState<Role | null>(null);
  const [loading, setLoading] = useState(true);
  const [addModalVisible, setAddModalVisible] = useState(false);
  const [editModalVisible, setEditModalVisible] = useState(false);
  const [permissionModalVisible, setPermissionModalVisible] = useState(false);

  // 使用 useApiParams 钩子
  const { sendRequest: fetchRolesRequest } = useApiParams(getAllRolesApi);
  const { sendRequest: addRoleRequest, loading: addLoading } = useApiParams(createRoleApi);
  const { sendRequest: updateRoleRequest, loading: updateLoading } = useApiParams(updateRoleApi);
  const { sendRequest: removeRoleRequest } = useApiParams(deleteRoleApi);

  // 获取角色列表
  const fetchRoles = async () => {
    setLoading(true);
    try {
      const roles = await fetchRolesRequest({}, { useURLSearchParams: false });
      setRoleList(roles || []);
      return roles;
    } catch (error) {
      console.error('获取角色列表失败:', error);
      return [];
    } finally {
      setLoading(false);
    }
  };

  // 添加角色
  const handleAddRole = async (values: { roleName: string, description: string,  permissionList}) => {
    // Todo: 添加角色的时候，权限的设置应该时有效的（现在是无效的）
    await addRoleRequest(
      {
        roleName: values.roleName,
        description: values.description,
        permissionList: values.permissionList
      },
      {
        onSuccess: () => {
          showToast({
            title: t('index.addSuccess'),
            color: 'success',
          });
          fetchRoles();
          setAddModalVisible(false);
        },
        onError: (error) => {
          console.error('添加角色失败:', error);
        }
      }
    );
  };

  // 更新角色
  const handleEditRole = async (values: { roleName: string, description: string }) => {
    if (!selectedRole) return;

    await updateRoleRequest(
      {
        roleId: selectedRole.roleId,
        roleName: values.roleName,
        description: values.description
      },
      {
        onSuccess: () => {
          showToast({
            title: t('index.updateSuccess'),
            color: 'success',
          });
          fetchRoles();
          setEditModalVisible(false);
        },
        onError: (error) => {
          console.error('更新角色失败:', error);
        }
      }
    );
  };

  // 删除角色
  const removeRole = async (roleId: string | number) => {
    await removeRoleRequest(
      { roleId },
      {
        useURLSearchParams: false,
        onSuccess: () => {
          showToast({
            title: t('index.deleteSuccess'),
            color: 'success',
          });
          fetchRoles();
        },
        onError: (error) => {
          console.error('删除角色失败:', error);
        }
      }
    );
  };

  // 处理权限保存
  const handleSavePermissions = async (checkedKeys: React.Key[]) => {
    if (!selectedRole) return;

    await updateRoleRequest(
      {
        roleId: selectedRole.roleId,
        roleName: selectedRole.roleName,
        permissionList: checkedKeys
      },
      {
        onSuccess: () => {
          showToast({
            title: t('index.permissionUpdateSuccess'),
            color: 'success',
          });
          setPermissionModalVisible(false);
        },
        onError: (error) => {
          console.error('保存权限失败:', error);
        }
      }
    );
  };

  // 初始化加载
  useEffect(() => {
    fetchRoles();
  }, []);

  // 打开编辑模态框
  const showEditModal = (role: Role) => {
    setSelectedRole(role);
    setEditModalVisible(true);
  };

  // 打开权限模态框
  const showPermissionModal = (role: Role) => {
    setSelectedRole(role);
    setPermissionModalVisible(true);
  };

  // 角色列表列定义
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
        record.roleName !== 'admin' ? (
          <>
            <Button
              onClick={() => showEditModal(record)}
              icon={<MdOutlineModeEdit />}
              type="text"
              className="mr-2"
            >
              {t('index.edit')}
            </Button>
            <Popconfirm
              title={t('index.confirmDelete', { name: record.roleName })}
              onConfirm={() => removeRole(record.roleId)}
            >
              <Button type="text" icon={<RiDeleteBin5Line />} danger className="mr-2">
                {t('index.delete')}
              </Button>
            </Popconfirm>
            <Button
              color="primary"
              variant="outlined"
              icon={<LuShieldCheck />}
              onClick={() => showPermissionModal(record)}
            >
              {t('index.configPermission')}
            </Button>
          </>
        ) : (
          <Button
            type="primary"
            onClick={() => showPermissionModal(record)}
          >
            {t('index.viewPermission')}
          </Button>
        )
      ),
    },
  ];

  return (
    <>
      <LoadingSpinner loading={loading} />
      <div className="p-4">
        <Button
          type="primary"
          onClick={() => setAddModalVisible(true)}
          className="mb-4"
        >
          {t('index.addRole')}
        </Button>

        <Table
          dataSource={roleList}
          columns={columns}
          rowKey="roleId"
        />

        {/* 添加角色模态框 */}
        <FormModal
          title={t('index.addRole')}
          open={addModalVisible}
          onCancel={() => setAddModalVisible(false)}
          confirmLoading={addLoading}
        >
          <FormModal.Section
            onFinish={handleAddRole}
          >
          <Form.Item
            name="roleName"
            label={t('index.roleName')}
            rules={[{ required: true, message: t('index.roleNameRequired') }]}
          >
            <Input />
          </Form.Item>
          <Form.Item
            name="description"
            label={t('index.description')}
          >
            <Input />
          </Form.Item>
          <Form.Item
            label={t('addModal.permissions')}
            name="permissionList"
          >
          {selectedRole && (
            <PermissionTree
              subjectId={selectedRole.roleId}
              subjectType="role"
              onSave={handleSavePermissions}
              readOnly={selectedRole.roleName === 'admin'}
            />
          )}
          </Form.Item>
          </FormModal.Section>
        </FormModal>

        {/* 编辑角色模态框 */}

        <FormModal
          title={t('editModal.title')}
          open={editModalVisible}
          onCancel={() => setEditModalVisible(false)}
          confirmLoading={updateLoading}
        >
          <FormModal.Section
            onFinish={handleEditRole}
            initialValues={selectedRole ? { roleName: selectedRole.roleName, description: selectedRole.description } : {}}
          >
          <Form.Item
            name="roleName"
            label={t('index.roleName')}
            rules={[{ required: true, message: t('index.roleNameRequired') }]}
          >
            <Input />
          </Form.Item>
          <Form.Item
            name="description"
            label={t('index.description')}
          >
            <Input />
          </Form.Item>
          </FormModal.Section>
        </FormModal>

        {/* 权限配置模态框 */}
        <CommonModal
          title={t('index.configPermission')}
          open={permissionModalVisible}
          onCancel={() => setPermissionModalVisible(false)}
          width={800}
          footer={null}
        >
          {selectedRole && (
            <PermissionTree
              subjectId={selectedRole.roleId}
              subjectType="role"
              onSave={handleSavePermissions}
              readOnly={selectedRole.roleName === 'admin'}
            />
          )}
        </CommonModal>
      </div>
    </>
  );
}
