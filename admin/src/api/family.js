import request from '@/utils/request'

// 获取我的家庭列表（用户拥有 + 用户加入的）
export function getFamilies(params = {}) {
  return request({
    url: '/family',
    method: 'get',
    params
  })
}

// 获取家庭详情
export function getFamily(id) {
  return request({
    url: `/family/${id}`,
    method: 'get'
  })
}

// 创建家庭
export function createFamily(data) {
  return request({
    url: '/family',
    method: 'post',
    data
  })
}

// 更新家庭
export function updateFamily(id, data) {
  return request({
    url: `/family/${id}`,
    method: 'put',
    data
  })
}

// 删除家庭
export function deleteFamily(id) {
  return request({
    url: `/family/${id}`,
    method: 'delete'
  })
}

// 生成家庭邀请码
export function createInvite(id, roomId) {
  const params = roomId ? { room_id: roomId } : {}
  return request({
    url: `/family/${id}/invite`,
    method: 'get',
    params
  })
}

// 搜索家庭成员
export function searchMembers(data) {
  return request({
    url: '/member/search',
    method: 'post',
    data
  })
}

// 邀请用户加入家庭
export function inviteMember(data) {
  return request({
    url: '/member/invite',
    method: 'post',
    data
  })
}

// 接受邀请码加入家庭
export function acceptInvite(inviteKey) {
  return request({
    url: '/member/accept',
    method: 'post',
    data: { invite_key: inviteKey }
  })
}

// 更新成员角色（管理员开关）
export function updateMember(id, admin) {
  return request({
    url: `/member/${id}`,
    method: 'put',
    data: { admin }
  })
}

// 移除成员
export function removeMember(id) {
  return request({
    url: `/member/${id}`,
    method: 'delete'
  })
}
