<template>
  <view class="page-family-index">
    <DsNavBar title="我的家庭" />

    <view class="family-content">

      <DsLoading v-if="isLoading && familyList.length === 0" text="加载家庭中..." />

      <DsEmpty
        v-else-if="familyList.length === 0"
        text="还没有家庭"
        subText="创建一个家庭，把设备、房间和成员都组织起来。"
        icon="home"
      >
        <template #action>
          <DsButton text="立即创建家庭" type="primary" size="large" @click="openCreateDialog" />
        </template>
      </DsEmpty>

      <view v-else class="family-list">
        <view
          v-for="family in familyList"
          :key="family.id"
          class="family-row"
          :class="{ active: currentFamilyId === family.id }"
          @click="goToDetail(family.id)"
        >
          <view class="family-badge">
            <text>{{ getFamilyInitial(family.name) }}</text>
          </view>

          <view class="family-main">
            <view class="family-headline">
              <text class="family-name">{{ family.name || '未命名家庭' }}</text>
            </view>
          </view>

          <view class="family-actions">
            <view class="family-switch" :class="{ active: currentFamilyId === family.id }" @click.stop="selectFamily(family)">
              <text>{{ currentFamilyId === family.id ? '已选中' : '切换' }}</text>
            </view>
            <DsIcon name="next" :size="18" color="#c0cad5" />
          </view>
        </view>
      </view>
    </view>

    <view v-if="familyList.length > 0" class="family-footer">
      <DsButton text="创建家庭" type="primary" size="large" block @click="openCreateDialog" />
    </view>

    <view v-if="showCreateDialog" class="family-modal-mask" @click="closeCreateDialog">
      <view class="family-modal" @click.stop>
        <text class="modal-title">创建家庭</text>
        <text class="modal-subtitle">输入家庭名称后即可创建，创建成功会自动切换到新家庭。</text>

        <view class="modal-form">
          <DsInput v-model="form.name" placeholder="例如：我的家 / 父母家 / 办公室" />
        </view>

        <view class="modal-actions">
          <DsButton text="取消" type="ghost" size="large" @click="closeCreateDialog" />
          <DsButton text="确认创建" type="primary" size="large" :loading="creating" :disabled="creating" @click="handleCreate" />
        </view>
      </view>
    </view>
  </view>
</template>

<script setup>
import { computed, ref } from 'vue'
import { onPullDownRefresh, onShow } from '@dcloudio/uni-app'
import { createFamily } from '@/api/family'
import { familyList, currentFamily, loadFamilyData, setCurrentFamily } from '@/utils/user'
import DsNavBar from '@/components/DsNavBar.vue'
import DsEmpty from '@/components/DsEmpty.vue'
import DsIcon from '@/components/DsIcon.vue'
import DsTag from '@/components/DsTag.vue'
import DsButton from '@/components/DsButton.vue'
import DsInput from '@/components/DsInput.vue'
import DsLoading from '@/components/DsLoading.vue'

const isLoading = ref(false)
const showCreateDialog = ref(false)
const creating = ref(false)
const form = ref({
  name: ''
})

const currentFamilyId = computed(() => currentFamily.value?.id || '')
const currentFamilyName = computed(() => currentFamily.value?.name || '未选择家庭')

const getFamilyInitial = (name) => {
  if (!name) return '家'
  return String(name).trim().slice(0, 1)
}

const loadList = async () => {
  isLoading.value = true
  try {
    await loadFamilyData()
  } catch (error) {
    console.error('加载家庭列表失败:', error)
  } finally {
    isLoading.value = false
    uni.stopPullDownRefresh()
  }
}

const goToDetail = (id) => {
  uni.navigateTo({ url: `/pages/family/detail?id=${id}` })
}

const selectFamily = (family) => {
  if (!family?.id || currentFamily.value?.id === family.id) return
  setCurrentFamily(family)
  uni.showToast({ title: `已切换到${family.name}`, icon: 'none' })
}

const openCreateDialog = () => {
  showCreateDialog.value = true
}

const closeCreateDialog = () => {
  if (creating.value) return
  showCreateDialog.value = false
}

const handleCreate = async () => {
  const name = form.value.name.trim()
  if (!name) {
    uni.showToast({ title: '请输入家庭名称', icon: 'none' })
    return
  }

  creating.value = true
  try {
    const res = await createFamily({ name })
    form.value.name = ''
    showCreateDialog.value = false
    await loadFamilyData()

    const createdId = res.data?.id || res.id
    if (createdId) {
      const created = familyList.value.find(item => item.id === createdId)
      if (created) {
        setCurrentFamily(created)
      }
    }

    uni.showToast({ title: '家庭已创建', icon: 'success' })
  } catch (error) {
    console.error('创建家庭失败:', error)
    uni.showToast({ title: error.message || '创建失败', icon: 'none' })
  } finally {
    creating.value = false
  }
}

onShow(() => {
  loadList()
})

onPullDownRefresh(() => {
  loadList()
})
</script>

<style lang="scss">
@import '@/components/variables.scss';

page {
  background: linear-gradient(180deg, #f5f8fb 0%, #eef3f8 100%);
}

.family-page {
  min-height: 100vh;
  background: linear-gradient(180deg, #f5f8fb 0%, #eef3f8 100%);
}

.family-content {
  padding: $spacing-base;
  padding-bottom: 220rpx;
}

.current-family-panel {
  margin-bottom: $spacing-base;
  padding: 36rpx 40rpx;
  border-radius: $radius-xl;
  background: linear-gradient(135deg, $primary-color 0%, $primary-dark 100%);
  box-shadow: 0 12rpx 28rpx rgba(91, 163, 232, 0.24);
}

.panel-eyebrow {
  display: block;
  font-size: $font-size-xs;
  color: rgba(255, 255, 255, 0.72);
  margin-bottom: 10rpx;
}

.panel-title {
  display: block;
  font-size: 40rpx;
  font-weight: 700;
  color: #fff;
  margin-bottom: 12rpx;
}

.panel-subtitle {
  display: block;
  font-size: $font-size-sm;
  line-height: 1.7;
  color: rgba(255, 255, 255, 0.88);
}

.family-list {
  display: flex;
  flex-direction: column;
  gap: $spacing-sm;
}

.family-row {
  display: flex;
  align-items: center;
  gap: $spacing-base;
  padding: 28rpx 30rpx;
  border-radius: $radius-xl;
  background: $bg-primary;
  box-shadow: 0 10rpx 26rpx rgba(15, 23, 42, 0.05);

  &.active {
    box-shadow: 0 12rpx 30rpx rgba(91, 163, 232, 0.18);
    border: 2rpx solid rgba(91, 163, 232, 0.28);
  }
}

.family-badge {
  width: 96rpx;
  height: 96rpx;
  border-radius: 28rpx;
  background: linear-gradient(135deg, rgba(91, 163, 232, 0.14) 0%, rgba(91, 163, 232, 0.3) 100%);
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;

  text {
    font-size: 34rpx;
    font-weight: 700;
    color: $primary-color;
  }
}

.family-main {
  flex: 1;
  min-width: 0;
}

.family-headline {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: $spacing-sm;
  margin-bottom: 10rpx;
}

.family-name {
  font-size: 34rpx;
  font-weight: 600;
  color: $text-primary;
}

.family-desc {
  display: block;
  font-size: $font-size-sm;
  color: $text-secondary;
  line-height: 1.7;
}

.family-actions {
  display: flex;
  align-items: center;
  gap: $spacing-sm;
  flex-shrink: 0;
}

.family-switch {
  min-width: 88rpx;
  height: 52rpx;
  padding: 0 18rpx;
  border-radius: 999rpx;
  background: #edf3f8;
  display: flex;
  align-items: center;
  justify-content: center;

  text {
    font-size: $font-size-xs;
    color: $text-secondary;
  }

  &.active {
    background: rgba(91, 163, 232, 0.14);

    text {
      color: $primary-color;
      font-weight: 600;
    }
  }
}

.family-footer {
  position: fixed;
  left: 0;
  right: 0;
  bottom: 0;
  padding: $spacing-base;
  padding-bottom: calc(#{$spacing-base} + env(safe-area-inset-bottom));
  background: linear-gradient(180deg, rgba(245, 248, 251, 0) 0%, rgba(245, 248, 251, 0.96) 24%, #f5f8fb 100%);
}

.family-modal-mask {
  position: fixed;
  inset: 0;
  background: rgba(15, 23, 42, 0.42);
  display: flex;
  align-items: center;
  justify-content: center;
  padding: $spacing-base;
  z-index: 1000;
}

.family-modal {
  width: 100%;
  max-width: 640rpx;
  padding: 40rpx;
  border-radius: $radius-xl;
  background: $bg-primary;
  box-shadow: 0 18rpx 40rpx rgba(15, 23, 42, 0.14);
}

.modal-title {
  display: block;
  font-size: 34rpx;
  font-weight: 700;
  color: $text-primary;
  margin-bottom: 12rpx;
}

.modal-subtitle {
  display: block;
  font-size: $font-size-sm;
  color: $text-secondary;
  line-height: 1.7;
  margin-bottom: $spacing-lg;
}

.modal-form {
  margin-bottom: $spacing-lg;
}

.modal-actions {
  display: flex;
  flex-direction: column;
  gap: $spacing-sm;
}
</style>
