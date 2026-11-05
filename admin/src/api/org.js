import request from '@/utils/request'

// 组织树（管理员返回全部，其他用户返回可管理子树）
export function getOrgTree() {
  return request({
    url: '/org/tree',
    method: 'get'
  })
}

export function getOrg(id) {
  return request({
    url: `/org/${id}`,
    method: 'get'
  })
}

export function createOrg(data) {
  return request({
    url: '/org',
    method: 'post',
    data
  })
}

export function updateOrg(id, data) {
  return request({
    url: `/org/${id}`,
    method: 'put',
    data
  })
}

export function deleteOrg(id) {
  return request({
    url: `/org/${id}`,
    method: 'delete'
  })
}

// 用户检索（负责人 / 成员选择器）
export function searchUsers(keyword) {
  return request({
    url: '/user/search',
    method: 'get',
    params: { keyword }
  })
}

// ========== 组织成员 ==========

export function getOrgMembers(id) {
  return request({
    url: `/org/${id}/members`,
    method: 'get'
  })
}

export function addOrgMember(id, userId) {
  return request({
    url: `/org/${id}/members`,
    method: 'post',
    data: { user_id: userId }
  })
}

export function removeOrgMember(id, userId) {
  return request({
    url: `/org/${id}/members/${userId}`,
    method: 'delete'
  })
}

// ========== 设备-组织绑定 ==========

export function getDeviceOrgs(id) {
  return request({
    url: `/device/${id}/orgs`,
    method: 'get'
  })
}

export function bindDeviceOrg(id, orgId) {
  return request({
    url: `/device/${id}/orgs`,
    method: 'post',
    data: { org_id: orgId }
  })
}

export function unbindDeviceOrg(id, orgId) {
  return request({
    url: `/device/${id}/orgs/${orgId}`,
    method: 'delete'
  })
}
