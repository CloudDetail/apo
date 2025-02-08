import { GetTeamParams, SaveTeamParams, TeamParams } from '../types/team'
import { get, headers, post } from '../utils/request'

export function getTeamsApi(params: GetTeamParams) {
  return get('/api/team', params)
}

export function addTeamApi(params: SaveTeamParams) {
  return post('/api/team/create', params)
}
export function updateTeamApi(params: SaveTeamParams) {
  return post('/api/team/update', params)
}
export function deleteTeamApi(teamId: string) {
  return post('/api/team/delete', { teamId }, headers.formUrlencoded)
}
