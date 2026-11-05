import request from '@/utils/request'

export const getFamilyList = () => {
  return request({
    url: '/family',
    method: 'GET'
  })
}

export const getFamilyDetail = (id) => {
  return request({
    url: `/family/${id}`,
    method: 'GET'
  })
}

export const createFamily = (data) => {
  return request({
    url: '/family',
    method: 'POST',
    data
  })
}

export const updateFamily = (id, data) => {
  return request({
    url: `/family/${id}`,
    method: 'PUT',
    data
  })
}

export const deleteFamily = (id) => {
  return request({
    url: `/family/${id}`,
    method: 'DELETE'
  })
}

export const getRoomList = (familyId) => {
  return request({
    url: '/room/search',
    method: 'POST',
    data: {
      page: 1,
      page_size: 100,
      filter: [{ field: 'family_id', operator: 'eq', value: familyId }]
    }
  })
}

export const createRoom = (data) => {
  return request({
    url: '/room',
    method: 'POST',
    data
  })
}

export const updateRoom = (id, data) => {
  return request({
    url: `/room/${id}`,
    method: 'PUT',
    data
  })
}

export const deleteRoom = (id) => {
  return request({
    url: `/room/${id}`,
    method: 'DELETE'
  })
}

export const getMemberList = (familyId) => {
  return request({
    url: '/member/search',
    method: 'POST',
    data: {
      page: 1,
      page_size: 100,
      filter: [{ field: 'family_id', operator: 'eq', value: familyId }]
    }
  })
}

export const inviteMember = (data) => {
  return request({
    url: '/member/invite',
    method: 'POST',
    data
  })
}

export const updateMemberRole = (id, admin) => {
  return request({
    url: `/member/${id}`,
    method: 'PUT',
    data: { admin }
  })
}

export const removeMember = (id) => {
  return request({
    url: `/member/${id}`,
    method: 'DELETE'
  })
}

export const createFamilyInvite = (familyId) => {
  return request({
    url: `/family/${familyId}/invite`,
    method: 'GET'
  })
}

export const acceptFamilyInvite = (inviteKey) => {
  return request({
    url: '/member/accept',
    method: 'POST',
    data: { invite_key: inviteKey }
  })
}
