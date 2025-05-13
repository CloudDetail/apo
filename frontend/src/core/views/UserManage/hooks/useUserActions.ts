/**
 * Copyright 2025 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import { useApiParams } from 'src/core/hooks/useApiParams'
import { useTranslation } from 'react-i18next'
import { notify } from 'src/core/utils/notify'
import { User } from 'src/core/types/user'
import * as userApi from 'src/core/api/user'
import { useRoleActions } from './useRoleActions'

interface UserSearchParams {
  username?: string
  corporation?: string
  currentPage: number
  pageSize: number
}

interface UpdateUserData {
  corporation?: string
  email?: string
  phone?: string
  roleId?: number | string
}

interface PasswordData {
  newPassword: string
  confirmPassword: string
}

export const useUserActions = () => {
  const { handleRevokeUserRole } = useRoleActions()

  const { t } = useTranslation('core/userManage')

  const api = {
    getList: useApiParams(userApi.getUserListApi),
    remove: useApiParams(userApi.removeUserApi),
    create: useApiParams(userApi.createUserApi),
    updateCorporation: useApiParams(userApi.updateCorporationApi),
    updateEmail: useApiParams(userApi.updateEmailApi),
    updatePhone: useApiParams(userApi.updatePhoneApi),
    resetPassword: useApiParams(userApi.updatePasswordWithNoOldPwdApi),
  }

  const fetchUsers = async (params: UserSearchParams) => {
    const result = await api.getList.sendRequest(params, { useURLSearchParams: false })
    if (result) {
      const { users, ...pagination } = result
      return {
        users: users.map((user: User) => ({
          ...user,
          role: user.roleList?.[0]?.roleName,
        })),
        ...pagination,
      }
    }
    return null
  }

  const removeUserById = async (userId: string | number) => {
    await api.remove.sendRequest(
      { userId },
      {
        useURLSearchParams: false,
        onSuccess: () => {
          notify({ message: t('index.deleteSuccess'), type: 'success' })
        },
      },
    )
  }

  const createNewUser = async (userData: Record<string, any>) => {
    const userDataReady = {
      ...userData,
      roleList: [userData.roleId],
    }
    await api.create.sendRequest(userDataReady, {
      onSuccess: () => {
        notify({ message: t('index.addSuccess'), type: 'success' })
      },
    })
  }

  const updateUser = async (user: User, updates: UpdateUserData) => {
    const { corporation, email, phone, roleId } = updates
    const tasks = []

    if (roleId !== user.roleList[0].roleId) {
      tasks.push(handleRevokeUserRole(user.userId, roleId))
    }

    if (corporation !== user.corporation) {
      tasks.push(api.updateCorporation.sendRequest({ userId: user.userId, corporation }))
    }
    if (email !== user.email) {
      tasks.push(api.updateEmail.sendRequest({ username: user.username, email }))
    }
    if (phone !== user.phone) {
      tasks.push(api.updatePhone.sendRequest({ username: user.username, phone }))
    }

    if (tasks.length > 0) {
      await Promise.all(tasks)
      notify({ message: t('index.updateSuccess'), type: 'success' })
    }
  }

  const resetPassword = async (userId: string | number, passwordData: PasswordData) => {
    await api.resetPassword.sendRequest({
      userId,
      ...passwordData,
    })
    notify({ message: t('index.updateSuccess'), type: 'success' })
  }

  return {
    fetchUsers,
    removeUserById,
    createNewUser,
    updateUser,
    resetPassword,
  }
}
