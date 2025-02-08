/**
 * Copyright 2025 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import { Pagination } from './common'

export interface GetTeamParams extends Pagination {
  teamName?: string
  featureList?: string[]
  dataGroupList?: string[]
}
export interface TeamParams {
  teamId: string
}
export interface SaveTeamParams {
  teamId?: string
  teamName: string
  description?: string
  userList: string[]
}
