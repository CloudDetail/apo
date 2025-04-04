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
} from 'antd'
import { getUserListApi, removeUserApi } from 'core/api/user'
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
import { useTranslation } from 'react-i18next'
import { LuShieldCheck } from 'react-icons/lu'
import DataGroupAuthorizeModal from 'src/core/components/PermissionAuthorize/DataGroupAuthorizeModal'
import CustomCard from 'src/core/components/Card/CustomCard'

export default function UserManage() {
  const { t } = useTranslation('core/userManage')
  const [modalAddVisibility, setModalAddVisibility] = useState(false)
  const [userList, setUserList] = useState([])
  const [username, setUsername] = useState('')
  const [role, setRole] = useState('')
  const [corporation, setCorporation] = useState('')
  const [tableVisibility, setTableVisibility] = useState(true)
  const [modalEditVisibility, setModalEditVisibility] = useState(false)
  const [currentPage, setCurrentPage] = useState(1)
  const [pageSize, setPageSize] = useState(10)
  const [total, setTotal] = useState(0)
  const [selectedUser, setSelectedUser] = useState(null)
  const [loading, setLoading] = useState(false)
  const { user, dispatch } = useUserContext()

  const [authorizeModalVisibility, setAuthorizeModalVisibility] = useState(false)

  //移除用户
  async function removeUser(prop) {
    const params = {
      userId: prop,
    }
    try {
      await removeUserApi(params)
      if (userList.length <= 1) {
        await getUserList(undefined, 'special')
      } else {
        await getUserList()
      }
      showToast({
        title: t('index.deleteSuccess'),
        color: 'success',
      })
    } catch (error) {
      console.log(error)
    }
  }

  //获取用户列表
  async function getUserList(signal = undefined, type = 'normal', search = false) {
    if (loading) return
    setLoading(true)
    const params =
      type === 'special'
        ? { currentPage: currentPage - 1, pageSize, username, role, corporation }
        : search
          ? { currentPage: 1, pageSize, username, role, corporation }
          : { currentPage, pageSize, username, role, corporation }

    try {
      const { users, currentPage, pageSize, total } = await getUserListApi(params, signal)
      setUserList(users)
      setCurrentPage(currentPage)
      setPageSize(pageSize)
      setTotal(total)
      setTableVisibility(true)
    } catch (error) {
      console.error(error)
    } finally {
      setLoading(false)
    }
  }

  //改变分页器
  function paginationChange(page, pageSize) {
    setPageSize(pageSize)
    setCurrentPage(page)
  }

  //用户列表列定义
  const columns = [
    {
      title: t('index.userName'),
      dataIndex: 'username',
      key: 'username',
      align: 'center',
    },
    // {
    //     title: '角色',
    //     dataIndex: 'role',
    //     key: 'role',
    //     align: 'center',
    //     width: "16%"
    // },
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
              {t('index.edit')}
            </Button>
            <Popconfirm
              title={t('index.confirmDelete', { name: username })}
              onConfirm={() => removeUser(userId)}
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
                setAuthorizeModalVisibility(true)
                setSelectedUser(record)
              }}
            >
              {t('index.dataGroup')}
            </Button>
          </>
        ) : (
          <></>
        )
      },
      width: 400,
    },
  ]

  //初始化列表数据
  useEffect(() => {
    const controller = new AbortController()
    const { signal } = controller // 获取信号对象
    getUserList(signal)
    return () => {
      controller.abort
    }
  }, [currentPage, pageSize])

  useEffect(() => {
    const controller = new AbortController()
    const { signal } = controller // 获取信号对象
    getUserList(signal, null, true)
    return () => {
      controller.abort
    }
  }, [username, role, corporation])

  const closeAuthorizeModal = () => {
    setAuthorizeModalVisibility(false)
    setSelectedUser(null)
  }
  const refresh = () => {
    getUserList()
    closeAuthorizeModal()
  }
  return (
    <>
      <CustomCard style={{ backgroundColor: 'inherit' }}>
        <LoadingSpinner loading={loading} />
        <Flex className="mb-3 h-[40px]">
          <Flex className="w-full justify-between">
            <Flex className="w-full">
              <Flex className="w-auto items-center justify-start mr-5">
                <p className="text-md mr-2">{t('index.userName')}：</p>
                <Input
                  placeholder={t('index.search')}
                  className="w-2/3"
                  value={username}
                  onChange={(e) => setUsername(e.target.value)}
                />
              </Flex>
              <Flex className="w-auto items-center justify-start">
                <p className="text-md mr-2">{t('index.corporation')}：</p>
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
      </CustomCard>
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
      <DataGroupAuthorizeModal
        open={authorizeModalVisibility}
        closeModal={closeAuthorizeModal}
        subjectId={selectedUser?.userId}
        subjectName={selectedUser?.username}
        type="user"
        refresh={refresh}
      />
    </>
  )
}
