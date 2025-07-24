/**
 * Copyright 2025 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import { Button, Flex, Popconfirm, Table } from 'antd'
import { useEffect, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { LuShieldCheck } from 'react-icons/lu'
import { MdOutlineEdit } from 'react-icons/md'
import { RiDeleteBin5Line } from 'react-icons/ri'
import { deleteTeamApi, getTeamsApi } from 'src/core/api/team'
import InfoModal from './InfoModal'
import { notify } from 'src/core/utils/notify'
import DataGroupAuthorizeModal from 'src/core/components/PermissionAuthorize/DataGroupAuthorizeModal'
import { BasicCard } from 'src/core/components/Card/BasicCard'

function TeamPage() {
  const { t } = useTranslation('core/team')
  const { t: ct } = useTranslation('common')
  const [data, setData] = useState([])
  const [currentPage, setCurrentPage] = useState(1)
  const [pageSize, setPageSize] = useState(10)
  const [total, setTotal] = useState(0)
  const [teamInfo, setTeamInfo] = useState(null)
  const [infoModalVisible, setInfoModalVisible] = useState(false)
  const [permissionModalVisible, setPermissionModalVisible] = useState(false)
  const columns = [
    {
      title: 'teamId',
      dataIndex: 'teamId',
      key: 'teamId',
      hidden: true,
    },
    {
      title: t('teamName'),
      dataIndex: 'teamName',
      key: 'teamName',
      width: 400,
    },
    {
      title: t('description'),
      dataIndex: 'description',
      key: 'description',
    },
    {
      title: t('userList'),
      dataIndex: 'userList',
      key: 'userList',
      width: 100,
      render: (value) => <span>{value?.length || 0}</span>,
    },
    {
      title: ct('operation'),
      dataIndex: 'operation',
      key: 'operation',
      width: 350,
      render: (_, record) => {
        return (
          <Flex align="center" justify="space-evenly">
            <Button
              type="text"
              onClick={() => {
                setInfoModalVisible(true)
                setTeamInfo(record)
              }}
              icon={
                <MdOutlineEdit className="!text-[var(--ant-color-primary-text)] !hover:text-[var(--ant-color-primary-text-active)]" />
              }
            >
              <span className="text-[var(--ant-color-primary-text)] hover:text-[var(--ant-color-primary-text-active)]">
                {ct('edit')}
              </span>
            </Button>
            <Popconfirm
              title={t('confirmDelete', {
                teamName: record.teamName,
              })}
              onConfirm={() => deleteTeam(record.teamId)}
              okText={ct('confirm')}
              cancelText={ct('cancel')}
            >
              <Button type="text" icon={<RiDeleteBin5Line />} danger>
                {ct('delete')}
              </Button>
            </Popconfirm>
            {import.meta.env.VITE_APP_CODE_VERSION !== 'CE' && (
              <Button
                color="primary"
                variant="outlined"
                icon={<LuShieldCheck />}
                onClick={() => {
                  setPermissionModalVisible(true)
                  setTeamInfo(record)
                }}
              >
                {t('authorize')}
              </Button>
            )}
          </Flex>
        )
      },
    },
  ]
  const getTeams = () => {
    getTeamsApi({
      currentPage: currentPage,
      pageSize: pageSize,
    }).then((res) => {
      setTotal(res.total || 0)
      setData(res.teamList || [])
    })
  }
  useEffect(() => {
    getTeams()
  }, [pageSize, currentPage])
  const closeInfoModal = () => {
    setInfoModalVisible(false)
    setTeamInfo(null)
  }
  const closePermissionModal = () => {
    setPermissionModalVisible(false)
    setTeamInfo(null)
  }
  const refresh = () => {
    closeInfoModal()
    closePermissionModal()
    getTeams()
  }
  const deleteTeam = (teamId: string) => {
    deleteTeamApi(teamId)
      .then((res) => {
        notify({
          type: 'success',
          message: t('deleteSuccess'),
        })
      })
      .finally(() => {
        getTeams()
      })
  }
  const closeAuthorizeModal = () => {
    setPermissionModalVisible(false)
    setTeamInfo(null)
  }
  const changePagination = (pagination) => {
    setPageSize(pagination.pageSize)
    setCurrentPage(pagination.current)
  }
  return (
    <BasicCard>
      <BasicCard.Header>
        <div className="w-full flex justify-between mt-2">
          {/* <DataGroupFilter /> */}
          <div></div>
          <Button type="primary" onClick={() => setInfoModalVisible(true)}>
            {ct('add')}
          </Button>
        </div>
      </BasicCard.Header>

      <BasicCard.Table>
        <Table
          dataSource={data}
          columns={columns}
          pagination={{ current: currentPage, pageSize: pageSize, total: total }}
          onChange={changePagination}
          scroll={{ y: 'calc(100vh - 240px)' }}
          className="overflow-auto"
        ></Table>
      </BasicCard.Table>

      <InfoModal
        open={infoModalVisible}
        closeModal={closeInfoModal}
        teamInfo={teamInfo}
        refresh={refresh}
      />
      <DataGroupAuthorizeModal
        open={permissionModalVisible}
        closeModal={closeAuthorizeModal}
        subjectId={teamInfo?.teamId}
        subjectName={teamInfo?.teamName}
        type="team"
        refresh={refresh}
      />
    </BasicCard>
  )
}
export default TeamPage
