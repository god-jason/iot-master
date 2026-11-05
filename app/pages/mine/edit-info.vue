﻿<template>
	<view class="page-mine-edit-info">
		<DsNavBar title="编辑资料" />

		<DsCard>

			<DsFormItem label="用户名">
				<DsInput v-model="username" disabled placeholder="请输入用户名" />
			</DsFormItem>
			<DsFormItem label="昵称">
				<DsInput v-model="nickname" placeholder="请输入昵称" />
			</DsFormItem>
			<DsFormItem label="手机号">
				<DsInput v-model="phone" placeholder="请输入手机号" />
			</DsFormItem>
			<DsFormItem label="邮箱">
				<DsInput v-model="email" placeholder="请输入邮箱" />
			</DsFormItem>

		</DsCard>

		<DsButton text="保存" type="primary" size="large" @click="saveInfo" />
	</view>
</template>

<script setup>
	import {
		ref,
		onMounted
	} from 'vue'
	import {
		getCurrentUser
	} from '@/api/auth'
	import {
		updateUserInfo
	} from '@/api/user'
	import DsNavBar from '@/components/DsNavBar.vue'
	import DsFormItem from '@/components/DsFormItem.vue'
	import DsInput from '@/components/DsInput.vue'
	import DsButton from '@/components/DsButton.vue'
	import DsCard from '@/components/DsCard.vue'
	const username = ref('')
	const nickname = ref('')
	const phone = ref('')
	const email = ref('')
	const userId = ref('')

	const saveInfo = async () => {
		try {
			await updateUserInfo({
				nickname: nickname.value,
				cellphone: phone.value,
				email: email.value
			})
			uni.showToast({
				title: '保存成功',
				icon: 'success'
			})
			setTimeout(() => {
				uni.navigateBack()
			}, 1500)
		} catch (error) {
			uni.showToast({
				title: '保存失败',
				icon: 'none'
			})
		}
	}

	const loadInfo = async () => {
		try {
			const res = await getCurrentUser()
			const data = res.data || res
			userId.value = data.id || ''
			username.value = data.username || ''
			nickname.value = data.nickname || ''
			phone.value = data.cellphone || data.phone || ''
			email.value = data.email || ''
		} catch (error) {
			username.value = '用户'
			nickname.value = ''
			phone.value = ''
			email.value = ''
		}
	}

	onMounted(() => {
		loadInfo()
	})
</script>

<style lang="scss">
	@import '@/components/variables.scss';

	page {
		background: $bg-secondary;
	}

	.edit-container {}
</style>