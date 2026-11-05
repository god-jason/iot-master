import { getCurrentUser } from '@/api/auth'
import { getFamilyList } from '@/api/family'
import { ref } from 'vue'

export const currentUser = ref(null)
export const familyList = ref([])
export const currentFamily = ref(null)

export const initUserData = async () => {
    const res = await getCurrentUser()
    const user = res.data || res
    currentUser.value = user

    await loadFamilyData()
}

export const loadFamilyData = async () => {
    try {
        const res = await getFamilyList()
        familyList.value = res.data || []

        if (familyList.value.length === 0) {
            currentFamily.value = null
            uni.removeStorageSync('current_family_id')
            return
        }

        const savedFamilyId = uni.getStorageSync('current_family_id')
        const found = familyList.value.find(f => f.id === savedFamilyId)

        if (found) {
            currentFamily.value = found
        } else {
            const first = familyList.value[0]
            currentFamily.value = first
            uni.setStorageSync('current_family_id', first.id)
        }
    } catch (error) {
        console.error('加载家庭数据失败:', error)
        familyList.value = []
        currentFamily.value = null
    }
}

export const setCurrentFamily = (family) => {
    currentFamily.value = family
    uni.setStorageSync('current_family_id', family.id)
}

export const getUserInfo = () => {
    if (currentUser.value) {
        return currentUser.value
    }
    const stored = uni.getStorageSync('user')
    if (stored) {
        try {
            return JSON.parse(stored)
        } catch (e) {
            return null
        }
    }
    return null
}