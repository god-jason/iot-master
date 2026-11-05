<script>
	import { initUserData } from '@/utils/user'
	import { navigateToLogin, isRedirectingToLogin } from '@/utils/navigation.js'
	import { initOEM } from '@/utils/oem'
	export default {
		onLaunch: async function() {
			console.log('App Launch')
			// 初始化 OEM 定制化配置（先加载缓存，再异步拉取最新）
			initOEM()
			const token = uni.getStorageSync('token')
			if (token) {
				try {
					await initUserData()
				} catch (error) {
					console.error('初始化用户信息失败:', error)
					navigateToLogin()
				}
			} else {
				if (!isRedirectingToLogin()) {
					setTimeout(() => {
						uni.reLaunch({ url: '/pages/login/index' })
					}, 50)
				}
			}
		},
		onShow: function() {
			console.log('App Show')
		},
		onHide: function() {
			console.log('App Hide')
		}
	}
</script>

<style lang="scss">
	@import '@/iconfont/iconfont.css';
	@import '@/components/variables.scss';

	page {
		background: $bg-secondary;
	}
</style>
