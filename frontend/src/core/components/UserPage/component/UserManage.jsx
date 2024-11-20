import { Flex, Form, Input, Button, Divider, Tooltip, Modal, Table, ConfigProvider, Popconfirm, Spin, Pagination } from "antd"
import { UserOutlined, LockOutlined } from "@ant-design/icons"
import { createUserApi, getUserListApi, removeUserApi } from "src/core/api/user";
import { showToast } from "src/core/utils/toast";
import { IoPersonAdd } from "react-icons/io5";
import { useEffect, useState } from "react";
import '../index.css'

export default function UserManage() {
    const [form] = Form.useForm()
    const [modalVisibility, setModalVisibility] = useState(false)
    const [userList, setUserList] = useState([])
    const [username, setUsername] = useState("")
    const [role, setRole] = useState("")
    const [corporation, setCorporation] = useState("")
    const [tableVisibility, setTableVisibility] = useState(true)
    const [currentPage, setCurrentPage] = useState(1)
    const [pageSize, setPageSize] = useState(10)
    const [total, setTotal] = useState(0)

    //创建用户
    async function createUser() {
        form.validateFields()
            .then(async (values) => {
                console.log(values)
                if (values.password !== values.confirmPassword) {
                    showToast({
                        title: '两次密码不一致,请重新输入',
                        color: 'danger'
                    })
                    return
                }
                try {
                    await createUserApi(values)
                    showToast({
                        title: '新用户添加成功',
                        color: 'success'
                    })
                    await getUserList()
                } catch (error) {
                    showToast({
                        title: error.response ? error.response.data.message : "未知错误",
                        color: 'danger'
                    })
                }
                form.resetFields()
            })
    }

    //移除用户
    async function removeUser(prop) {
        const params = {
            username: prop
        }
        try {
            await removeUserApi(params)
            await getUserList()
        } catch (error) {
            showToast({
                title: "移除用户失败",
                message: error,
                color: "danger"
            })
        }
    }

    //获取用户列表
    async function getUserList() {
        let loadingTimer
        loadingTimer = setTimeout(() => {
            setTableVisibility(false)
        }, 500)

        const params = {
            currentPage,
            pageSize,
            username,
            role,
            corporation
        }

        try {
            const { users, currentPage, pageSize, total } = await getUserListApi(params)
            clearTimeout(loadingTimer)
            setUserList(users)
            setCurrentPage(currentPage)
            setPageSize(pageSize)
            setTotal(total)
            setTableVisibility(true)
        } catch (error) {
            showToast({
                title: "获取用户列表失败",
                color: "danger   "
            })
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
            width: 120
        },
        {
            title: '角色',
            dataIndex: 'role',
            key: 'role',
            align: 'center',
            width: 120
        },
        {
            title: '组织',
            dataIndex: 'corporation',
            key: 'corporation',
            align: 'center',
            width: 120
        },
        {
            title: '手机',
            dataIndex: 'phone',
            key: 'phone',
            align: 'center',
            width: 120
        },
        {
            title: '邮箱',
            dataIndex: 'email',
            key: 'email',
            align: 'center',
            width: 120
        },
        {
            title: '操作',
            dataIndex: 'username',
            key: 'username',
            align: 'center',
            render: (prop) => {
                return (
                    <Popconfirm
                        title={`确定要移除用户名为${prop}的用户吗`}
                        onConfirm={() => removeUser(prop)}
                    >
                        <Button type="link" >
                            移除用户
                        </Button>
                    </Popconfirm>
                )
            },
            width: 120
        }
    ]


    //初始化列表数据
    useEffect(() => {
        getUserList()
    }, [])

    useEffect(() => {
        getUserList()
    }, [username, role, corporation, currentPage, pageSize])

    return (
        <Flex vertical className="w-full pl-5">
            <Flex className="mb-10">
                <Divider orientation="left">用户管理</Divider>
            </Flex>
            <Flex className="mb-10">
                <Flex className="w-full justify-between">
                    <Flex className="w-full">
                        <Flex className="w-auto items-center justify-start mr-5">
                            <p className="text-md mr-2">用户名称:</p>
                            <Input placeholder="检索" className="w-52" value={username} onChange={(e) => setUsername(e.target.value)} />
                        </Flex>
                        <Flex className="w-auto items-center justify-start mr-5">
                            <p className="text-md mr-2">角色:</p>
                            <Input placeholder="检索" className="w-40" value={role} onChange={(e) => setRole(e.target.value)} />
                        </Flex>
                        <Flex className="w-auto items-center justify-start">
                            <p className="text-md mr-2">组织:</p>
                            <Input placeholder="检索" className="w-40" value={corporation} onChange={(e) => setCorporation(e.target.value)} />
                        </Flex>
                    </Flex>
                    <Flex className="w-full justify-end">
                        <Button onClick={() => setModalVisibility(true)}><Tooltip title='新增用户'><IoPersonAdd className="w-6 h-6" /></Tooltip></Button>
                    </Flex>
                </Flex>
            </Flex>
            <ConfigProvider
                theme={{
                    components: {
                        Table: {
                            headerBg: "#222631",
                            bodySortBg: "#222631"
                        }
                    }
                }}
            >
                {

                    tableVisibility ?
                        (<Flex vertical className={tableVisibility ? "w-11/12 flex" : "w-11/12 hidden"}>
                            <Table
                                dataSource={userList}
                                columns={columns}
                                className="w-full"
                                scroll={{ y: 800 }}
                                pagination={false}
                            />
                            <Pagination
                                className="mt-4"
                                align="end"
                                current={currentPage}
                                pageSize={pageSize}
                                total={total}
                                pageSizeOptions={[10, 20, 40]}
                                showSizeChanger
                                onChange={paginationChange}
                                showQuickJumper
                            />
                        </Flex>) :
                        (<Spin />)
                }

            </ConfigProvider>
            <Modal
                open={modalVisibility}
                onCancel={() => setModalVisibility(false)}
                maskClosable={false}
                title="新增用户"
                okText="新增"
                onOk={createUser}
            >
                <Flex vertical className="w-full mt-4 mb-4">
                    <Flex vertical className="w-full justify-center items-center">
                        <Form form={form} requiredMark={false}>
                            <Form.Item
                                label="用户名&#12288;"
                                name="username"
                                rules={[
                                    { required: true, message: '请输入用户名' }
                                ]}
                            >
                                <div className="flex justify-start items-start">
                                    <Input prefix={<UserOutlined />} placeholder="请输入用户名" className="w-80" />
                                </div>
                            </Form.Item>
                            <Form.Item
                                label="密&#12288;&#12288;码"
                                name="password"
                                rules={[
                                    { required: true, message: '请输入密码' }
                                ]}
                            >
                                <div className="flex justify-start items-start">
                                    <Input.Password prefix={<LockOutlined />} placeholder="请输入密码" className="w-80" />
                                </div>
                            </Form.Item>
                            <Form.Item
                                label="重复密码"
                                name="confirmPassword"
                                rules={[
                                    { required: true, message: '请再次输入密码' }
                                ]}
                            >
                                <Input.Password prefix={<LockOutlined />} placeholder="请再次输入密码" className="w-80" />
                            </Form.Item>
                        </Form>
                    </Flex>
                </Flex>
            </Modal>
        </Flex>
    )
}