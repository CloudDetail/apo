import { Flex, Form, Input, Button, Divider, Tooltip, Modal, Table, ConfigProvider, Popconfirm, Spin, Pagination } from "antd"
import { getUserListApi, removeUserApi } from "core/api/user";
import { showToast } from "core/utils/toast";
import { useEffect, useState } from "react";
import { RiDeleteBin5Line } from 'react-icons/ri'
import { MdOutlineModeEdit } from "react-icons/md";
import EditModal from "./componnets/EditModal";
import AddModal from "./componnets/AddModal";
import { BsPersonFillAdd } from "react-icons/bs";
import LoadingSpinner from 'src/core/components/Spinner'
import { useUserContext } from "src/core/contexts/UserContext";
import style from "./index.module.css"


export default function UserManage() {
    const [modalAddVisibility, setModalAddVisibility] = useState(false)
    const [userList, setUserList] = useState([])
    const [username, setUsername] = useState("")
    const [role, setRole] = useState("")
    const [corporation, setCorporation] = useState("")
    const [tableVisibility, setTableVisibility] = useState(true)
    const [modalEditVisibility, setModalEditVisibility] = useState(false)
    const [currentPage, setCurrentPage] = useState(1)
    const [pageSize, setPageSize] = useState(10)
    const [total, setTotal] = useState(0)
    const [selectedUser, setSelectedUser] = useState("")
    const [loading, setLoading] = useState(false)
    const { user, dispatchUser } = useUserContext()
    //移除用户
    async function removeUser(prop) {
        const params = {
            username: prop
        }
        try {
            await removeUserApi(params)
            if (userList.length <= 1) {
                await getUserList(undefined, "special")
            } else {
                await getUserList()
            }
            showToast({
                title: "移除用户成功",
                color: "success"
            })
        } catch (error) {
            const errorMessage = error.response?.data?.message || "移除失败"
            showToast({
                title: errorMessage,
                color: "danger"
            })
            console.log(error)
        }
    }

    //获取用户列表
    async function getUserList(signal = undefined, type = "normal", search = false) {
        if (loading) return
        setLoading(true)
        const params = type === "special" ?
            { currentPage: currentPage - 1, pageSize, username, role, corporation } :
            search ? { currentPage: 1, pageSize, username, role, corporation } : { currentPage, pageSize, username, role, corporation }

        try {
            const { users, currentPage, pageSize, total } = await getUserListApi(params, signal)
            setUserList(users)
            setCurrentPage(currentPage)
            setPageSize(pageSize)
            setTotal(total)
            setTableVisibility(true)
        } catch (error) {
            console.error(error)
            showToast({
                title: "获取用户列表失败",
                color: "danger"
            })
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
            title: '用户名',
            dataIndex: 'username',
            key: 'username',
            align: 'center',
            width: "16%"
        },
        // {
        //     title: '角色',
        //     dataIndex: 'role',
        //     key: 'role',
        //     align: 'center',
        //     width: "16%"
        // },
        {
            title: '组织',
            dataIndex: 'corporation',
            key: 'corporation',
            align: 'center',
            width: "16%"
        },
        {
            title: '手机',
            dataIndex: 'phone',
            key: 'phone',
            align: 'center',
            width: "16%"
        },
        {
            title: '邮箱',
            dataIndex: 'email',
            key: 'email',
            align: 'center',
            width: "16%"
        },
        {
            title: '操作',
            dataIndex: 'username',
            key: 'username',
            align: 'center',
            render: (prop) => {
                return user.user.username !== prop ?
                    (
                        <>
                            <Button onClick={() => {
                                setSelectedUser(prop)
                                setModalEditVisibility(true)
                            }} icon={<MdOutlineModeEdit />} type="text" className="mr-1">
                                编辑
                            </Button>
                            <Popconfirm
                                title={`确定要移除用户名为${prop}的用户吗`}
                                onConfirm={() => removeUser(prop)}
                            >
                                <Button type="text" icon={<RiDeleteBin5Line />} danger>
                                    删除
                                </Button>
                            </Popconfirm>
                        </>
                    ) : <>

                    </>
            },
            width: "16%"
        }
    ]

    //初始化列表数据
    useEffect(() => {
        const controller = new AbortController();
        const { signal } = controller; // 获取信号对象
        getUserList(signal)
        return () => {
            controller.abort
        }
    }, [currentPage, pageSize])

    useEffect(() => {
        const controller = new AbortController();
        const { signal } = controller; // 获取信号对象
        getUserList(signal, null, true)
        return () => {
            controller.abort
        }
    }, [username, role, corporation])

    return (
        <>
            <LoadingSpinner loading={loading} />
            <div className={style.userManageContainer}>
                <Flex className="mb-3 h-[40px]">
                    <Flex className="w-full justify-between">
                        <Flex className="w-full">
                            <Flex className="w-auto items-center justify-start mr-5">
                                <p className="text-md mr-2">用户名称:</p>
                                <Input placeholder="检索" className="w-2/3" value={username} onChange={(e) => setUsername(e.target.value)} />
                            </Flex>
                            <Flex className="w-auto items-center justify-start">
                                <p className="text-md mr-2">组织:</p>
                                <Input placeholder="检索" className="w-2/3" value={corporation} onChange={(e) => setCorporation(e.target.value)} />
                            </Flex>
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
                                headerBg: "#222631",
                            }
                        }
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
