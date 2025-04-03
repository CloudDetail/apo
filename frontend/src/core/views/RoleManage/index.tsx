/**
 * Copyright 2024 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import {
  Flex,
  Form,
  Input,
  Button,
  Divider,
  Tooltip,
  Modal,
  Table,
  ConfigProvider,
  Popconfirm,
  Spin,
  Pagination,
  Dropdown,
  Space,
  message,
} from 'antd'
import { DownOutlined } from '@ant-design/icons'
import { getUserListApi, removeUserApi } from 'core/api/user'
import { showToast } from 'core/utils/toast'
import { useEffect, useState } from 'react'
import { RiDeleteBin5Line } from 'react-icons/ri'
import { MdOutlineModeEdit } from 'react-icons/md'
import EditModal from './components/EditModal'
import AddModal from './components/AddModal'
import { BsPersonFillAdd } from 'react-icons/bs'
import LoadingSpinner from 'src/core/components/Spinner'
import { useUserContext } from 'src/core/contexts/UserContext'
import style from '../UserManage/index.module.css'
import { useTranslation } from 'react-i18next'
import { LuShieldCheck } from 'react-icons/lu'
import DataGroupAuthorizeModal from 'src/core/components/PermissionAuthorize/DataGroupAuthorizeModal'
import { deleteRoleApi, getAllRolesApi, revokeUserRoleApi } from 'src/core/api/role'

export default function UserManage() {
  const { t } = useTranslation('core/roleManage')
  const [modalAddVisibility, setModalAddVisibility] = useState(false)
  // const [userList, setUserList] = useState([])
  const [username, setUsername] = useState('')
  const [role, setRole] = useState('')
  const [corporation, setCorporation] = useState('')
  const [tableVisibility, setTableVisibility] = useState(true)
  const [modalEditVisibility, setModalEditVisibility] = useState(false)
  const [currentPage, setCurrentPage] = useState(1)
  const [pageSize, setPageSize] = useState(10)
  const [total, setTotal] = useState(0)
  const [selectedRole, setSelectedRole] = useState(null)
  const [loading, setLoading] = useState(false)
  const [roleList, setRoleList] = useState([])

  const [authorizeModalVisibility, setAuthorizeModalVisibility] = useState(false)

  async function deleteRole(prop) {
    const params = {
      roleId: prop,
    }
    try {
      await deleteRoleApi(params)
      console.log('deleting role: ', params)
      await fetchRoles()
      showToast({
        title: t('index.deleteSuccess'),
        color: 'success',
      })
    } catch (error) {
      console.log(error)
    }
  }

  async function fetchRoles() {
    try {
      const roles = await getAllRolesApi(); // 等待 API 返回数据
      setRoleList(roles)
      console.log('roles: ', roles)
    } catch (error) {
      console.error("Failed to fetch roles: ", error); // 捕获并处理错误
    }
  }

  useEffect(() => {
    fetchRoles(); // 调用异步函数
  }, []); // 空依赖数组，确保只在组件挂载时调用一次

  //改变分页器
  function paginationChange(page, pageSize) {
    setPageSize(pageSize)
    setCurrentPage(page)
  }

  //用户列表列定义
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
      dataIndex: 'userId',
      key: 'userId',
      align: 'center',
      render: (_, role) => {
        return role.roleName !== 'admin' ? (
          <>
            <Button
              onClick={() => {
                setSelectedRole(role)
                setModalEditVisibility(true)
              }}
              icon={<MdOutlineModeEdit />}
              type="text"
              className="mr-1"
            >
              {t('index.edit')}
            </Button>
            <Popconfirm
              title={t('index.confirmDelete', { name: username })}
              onConfirm={() => deleteRole(role.roleId)}
            >
              <Button type="text" icon={<RiDeleteBin5Line />} danger className="mr-1">
                {t('index.delete')}
              </Button>
            </Popconfirm>
          </>
        ) : (
          <></>
        )
      },
      width: 400,
    },
  ]

  const closeAuthorizeModal = () => {
    setAuthorizeModalVisibility(false)
    setSelectedRole(null)
  }
  return (
    <>
      <LoadingSpinner loading={loading} />
      <div className={style.userManageContainer}>
        <Flex className="mb-3 h-[40px]">
          <Flex className="w-full justify-between">
            <Flex className="w-full">
              <Flex className="w-auto items-center justify-start mr-5">
                <p className="text-md mr-2">{t('index.roleName')}：</p>
                <Input
                  placeholder={t('index.search')}
                  className="w-2/3"
                  value={username}
                  onChange={(e) => setUsername(e.target.value)}
                />
              </Flex>
              <Flex className="w-auto items-center justify-start">
                <p className="text-md mr-2">{t('index.description')}：</p>
                <Input
                  placeholder={t('index.search')}
                  className="w-2/3"
                  value={corporation}
                  onChange={(e) => setCorporation(e.target.value)}
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
                <span className="text-xs">{t('index.addRole')}</span>
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
              dataSource={roleList}
              columns={columns}
              pagination={false}
              loading={!tableVisibility}
              scroll={{ y: 'calc(100vh - 220px)' }}
            />
            <Pagination
              className="mt-4 absolute bottom-0 right-0"
              align="end"
              current={currentPage}
              pageSize={pageSize}
              total={total}
              pageSizeOptions={[10, 30, 50]}
              showSizeChanger
              onChange={paginationChange}
              showQuickJumper
            />
          </Flex>
        </ConfigProvider>
      </div>
      <EditModal
        selectedRole={selectedRole}
        modalEditVisibility={modalEditVisibility}
        setModalEditVisibility={setModalEditVisibility}
        getRoleList={fetchRoles}
      />
      <AddModal
        modalAddVisibility={modalAddVisibility}
        setModalAddVisibility={setModalAddVisibility}
        getRoleList={fetchRoles}
      />
      {/* <DataGroupAuthorizeModal
        open={authorizeModalVisibility}
        closeModal={closeAuthorizeModal}
        subjectId={selectedUser?.userId}
        subjectName={selectedUser?.username}
        type="user"
        refresh={refresh}
      /> */}
    </>
  )
}
