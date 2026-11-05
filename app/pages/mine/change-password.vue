<template>
  <view class="page-mine-change-password">
    <DsNavBar title="修改密码" />
    <DsCard>
      <DsFormItem label="原密码">
        <DsInput v-model="oldPassword" type="password" placeholder="请输入原密码" />
      </DsFormItem>
      <DsFormItem label="新密码">
        <DsInput v-model="newPassword" type="password" placeholder="请输入新密码" />
      </DsFormItem>
      <DsFormItem label="确认密码">
        <DsInput v-model="confirmPassword" type="password" placeholder="请再次输入新密码" />
      </DsFormItem>
      <DsButton text="修改密码" type="primary" size="large" @click="changePassword" />
    </DsCard>
  </view>
</template>

<script setup>
import { ref } from 'vue'
import { changePassword as changePasswordApi } from '@/api/user'
import DsNavBar from '@/components/DsNavBar.vue'
import DsFormItem from '@/components/DsFormItem.vue'
import DsInput from '@/components/DsInput.vue'
import DsButton from '@/components/DsButton.vue'
import DsCard from '@/components/DsCard.vue'

const oldPassword = ref('')
const newPassword = ref('')
const confirmPassword = ref('')

const changePassword = async () => {
  if (!oldPassword.value) {
    uni.showToast({ title: '请输入原密码', icon: 'none' })
    return
  }
  if (!newPassword.value) {
    uni.showToast({ title: '请输入新密码', icon: 'none' })
    return
  }
  if (newPassword.value.length < 6) {
    uni.showToast({ title: '密码长度不能少于6位', icon: 'none' })
    return
  }
  if (newPassword.value !== confirmPassword.value) {
    uni.showToast({ title: '两次输入的密码不一致', icon: 'none' })
    return
  }

  try {
    await changePasswordApi({
      oldPassword: oldPassword.value,
      newPassword: newPassword.value
    })
    uni.showToast({ title: '修改成功', icon: 'success' })
    setTimeout(() => {
      uni.navigateBack()
    }, 1500)
  } catch (error) {
    uni.showToast({ title: error.message || '修改失败', icon: 'none' })
  }
}
</script>

<style lang="scss">
@import '@/components/variables.scss';

page {
  background: #F0F7FF;
}

.change-container {
}
</style>
