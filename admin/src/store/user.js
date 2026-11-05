import { defineStore } from 'pinia'
import Cookies from 'js-cookie'

export const useUserStore = defineStore('user', {
  state: () => ({
    id: null,
    username: '',
    email: '',
    cellphone: '',
    avatar: '',
    token: Cookies.get('token') || null,
    admin: false,
    orgId: 0
  }),
  getters: {
    isLoggedIn: (state) => !!state.token,
    isAdmin: (state) => state.admin,
  },
  actions: {
    setUser(user) {
      this.id = user.id
      this.username = user.username
      this.email = user.email || ''
      this.cellphone = user.cellphone || ''
      this.avatar = user.avatar || ''
      this.admin = user.admin || false
      this.orgId = user.org_id || 0
    },
    setToken(token) {
      this.token = token
      Cookies.set('token', token, { expires: 7 })
    },
    updateUserInfo(info) {
      this.email = info.email || ''
      this.cellphone = info.cellphone || ''
      this.avatar = info.avatar || ''
    },
    logout() {
      this.id = null
      this.username = ''
      this.email = ''
      this.cellphone = ''
      this.avatar = ''
      this.token = null
      Cookies.remove('token')
    }
  }
})
