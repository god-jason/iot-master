import { submitLoginToken } from '@/api/auth'
import { createDevice } from '@/api/device'
import { acceptFamilyInvite } from '@/api/family'
import { currentFamily } from './user'

export const scanCode = async (onSuccess) => {
  return new Promise((resolve, reject) => {
    uni.scanCode({
      success: async (res) => {
        const content = res.result.trim()
        const token = uni.getStorageSync('token')
        
        try {
          if (content.startsWith('noob-login:')) {
            const loginId = content.replace('noob-login:', '')
            await submitLoginToken(loginId, token)
            uni.showToast({ title: '登录成功', icon: 'success' })
          } else if (content.startsWith('noob-invite:')) {
            const inviteKey = content.replace('noob-invite:', '')
            await acceptFamilyInvite(inviteKey)
            uni.showToast({ title: '绑定家庭成功', icon: 'success' })
          } else {
            const familyId = currentFamily.value?.id
            if (!familyId) {
              uni.showToast({ title: '请先选择家庭', icon: 'none' })
              reject(new Error('请先选择家庭'))
              return
            }
            await createDevice({ sn: content, name: content, family_id: familyId })
            uni.showToast({ title: '绑定成功', icon: 'success' })
          }
          
          if (typeof onSuccess === 'function') {
            onSuccess()
          }
          
          resolve(content)
        } catch (error) {
          uni.showToast({ title: error.message || '操作失败', icon: 'none' })
          reject(error)
        }
      },
      fail: (error) => {
        uni.showToast({ title: '扫描失败', icon: 'none' })
        reject(error)
      }
    })
  })
}