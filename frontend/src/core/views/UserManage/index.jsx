/**
 * Copyright 2024 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import {
  Flex,
  Input,
  Button,
  Table,
  ConfigProvider,
  Popconfirm,
  Pagination,
  Select,
  Tag,
} from 'antd'
import { getUserListApi, removeUserApi, getRoleListApi } from 'core/api/user'
import { showToast } from 'core/utils/toast'
import { useEffect, useState } from 'react'
import { RiDeleteBin5Line } from 'react-icons/ri'
import { MdOutlineModeEdit } from 'react-icons/md'
import EditModal from './componnets/EditModal'
import AddModal from './componnets/AddModal'
import { BsPersonFillAdd } from 'react-icons/bs'
import LoadingSpinner from 'src/core/components/Spinner'
import { useUserContext } from 'src/core/contexts/UserContext'
import style from './index.module.css'

export default function UserManage() {
  const [modalAddVisibility, setModalAddVisibility] = useState(false)
  const [userList, setUserList] = useState(null)
  const [username, setUsername] = useState('')
  const [role, setRole] = useState([])
  const [roleOptions, setRoleOptions] = useState(null)
  const [corporation, setCorporation] = useState('')
  const [tableVisibility, setTableVisibility] = useState(true)
  const [modalEditVisibility, setModalEditVisibility] = useState(false)
  const [currentPage, setCurrentPage] = useState(1)
  const [pageSize, setPageSize] = useState(10)
  const [total, setTotal] = useState(0)
  const [selectedUser, setSelectedUser] = useState(null)
  const [loading, setLoading] = useState(false)
  const { user } = useUserContext()

  //移除用户
  function removeUser(userId) {
    removeUserApi({ userId })
      .then(() => getUserList(userList.length <= 1 ? 'special' : undefined))
      .then(() => showToast({ title: '移除用户成功', color: 'success' }))
      .catch((error) => console.log(error))
  }

  //获取用户列表
  function getUserList(type = 'normal', search = false, signal) {
    const params = {
      currentPage: type === 'special' ? currentPage - 1 : search ? 1 : currentPage,
      pageSize,
      username,
      roleList: role,
      corporation,
    }

    return getUserListApi(params, signal)
      .then(({ users, currentPage, pageSize, total }) => {
        setUserList(users)
        setCurrentPage(currentPage)
        setPageSize(pageSize)
        setTotal(total)
        setTableVisibility(true)
      })
      .catch((error) => {
        console.error(error)
        showToast({ title: '获取用户列表失败', color: 'danger' })
      })
  }

  //获取角色列表
  function getRoleList() {
    return getRoleListApi()
      .then((response) => {
        const options = response.map((option) => ({
          value: option.roleId,
          label: option.roleName,
        }))
        setRoleOptions(options)
      })
      .catch((error) => console.log(error))
  }

  //改变分页器
  function paginationChange(page, pageSize) {
    setPageSize(pageSize)
    setCurrentPage(page)
  }

  const TagsColor = {
    admin: 'magenta',
    manager: 'orange',
    viewer: 'lime',
  }

  //用户列表列定义
  const columns = [
    {
      title: '用户名',
      dataIndex: 'username',
      key: 'username',
      align: 'center',
      width: '16%',
    },
    {
      title: '角色',
      dataIndex: 'roleList',
      key: 'roleList',
      align: 'center',
      width: '16%',
      render: (roleList) => {
        return (
          <div className="flex justify-center items-center flex-wrap">
            {roleList?.map((role) => (
              <div className="ml-1 mt-1">
                <Tag color={TagsColor[role.roleName]}>{role.roleName}</Tag>
              </div>
            ))}
          </div>
        )
      },
    },
    {
      title: '组织',
      dataIndex: 'corporation',
      key: 'corporation',
      align: 'center',
      width: '16%',
    },
    {
      title: '手机',
      dataIndex: 'phone',
      key: 'phone',
      align: 'center',
      width: '16%',
    },
    {
      title: '邮箱',
      dataIndex: 'email',
      key: 'email',
      align: 'center',
      width: '16%',
    },
    {
      title: '操作',
      dataIndex: 'userId',
      key: 'userId',
      align: 'center',
      render: (userId, record) => {
        const { username } = record
        return user.userId !== userId ? (
          <>
            <Button
              onClick={() => {
                setSelectedUser(record)
                setModalEditVisibility(true)
              }}
              icon={<MdOutlineModeEdit />}
              type="text"
              className="mr-1"
            >
              编辑
            </Button>
            <Popconfirm
              title={`确定要移除用户名为${username}的用户吗`}
              onConfirm={() => removeUser(userId)}
            >
              <Button type="text" icon={<RiDeleteBin5Line />} danger>
                删除
              </Button>
            </Popconfirm>
          </>
        ) : null
      },
      width: '16%',
    },
  ]

  //初始化列表数据
  useEffect(() => {
    const controller = new AbortController()
    setLoading(true)
    Promise.all([getUserList(undefined, false, controller.signal), getRoleList()]).finally(() =>
      setLoading(false),
    )
    return () => controller.abort()
  }, [currentPage, pageSize])

  useEffect(() => {
    const controller = new AbortController()
    setLoading(true)
    getUserList(null, true, controller.signal).finally(() => setLoading(false))
    return () => controller.abort()
  }, [username, role, corporation])

  return (
    <>
      <LoadingSpinner loading={loading} />
      <div className={style.userManageContainer}>
        <Flex className="mb-3 h-[40px]">
          <Flex className="w-full justify-between">
            <Flex className="w-full">
              {[
                { label: '用户名称:', value: username, onChange: setUsername, component: Input },
                {
                  label: '角色:',
                  value: role,
                  onChange: setRole,
                  roleOptions: roleOptions,
                  mode: 'multiple',
                  allowClear: true,
                  maxTagCount: 1,
                  maxTagPlaceholder: (omittedValues) => `+${omittedValues.length} 更多`,
                  component: Select,
                },
                { label: '组织:', value: corporation, onChange: setCorporation, component: Input },
              ].map(
                (
                  {
                    label,
                    value,
                    onChange,
                    roleOptions,
                    mode,
                    allowClear,
                    maxTagCount,
                    maxTagPlaceholder,
                    component: Component,
                  },
                  index,
                ) => (
                  <Flex key={index} className="whitespace-nowrap items-center justify-start mr-5">
                    <p className="text-md mr-2">{label}</p>
                    <Component
                      placeholder="检索"
                      className="min-w-48 max-w-60"
                      value={value}
                      options={roleOptions}
                      mode={mode}
                      maxTagCount={maxTagCount}
                      maxTagPlaceholder={maxTagPlaceholder}
                      allowClear={allowClear}
                      onChange={(e) => onChange(e.target ? e.target.value : e)}
                    />
                  </Flex>
                ),
              )}
            </Flex>
            <Flex className="w-full justify-end items-center">
              <Button
                type="primary"
                icon={<BsPersonFillAdd size={20} />}
                onClick={() => setModalAddVisibility(true)}
                className="flex-grow-0 flex-shrink-0"
              >
                <span className="text-xs">新增用户</span>
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
        selectedUser={selectedUser}
        modalEditVisibility={modalEditVisibility}
        setModalEditVisibility={setModalEditVisibility}
        getUserList={getUserList}
      />
      <AddModal
        modalAddVisibility={modalAddVisibility}
        setModalAddVisibility={setModalAddVisibility}
        getUserList={getUserList}
      />
    </>
  )
}
