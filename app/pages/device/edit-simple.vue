<template>
  <view class="page-device-edit-simple">
    <DsNavBar title="编辑设备" />
    <DsLoading v-if="loading" text="加载中..." />

    <DsCard v-if="!loading">
      <DsFormItem label="设备名称">
        <DsInput
          v-model="form.name"
          placeholder="请输入设备名称"
          maxlength="32"
        />
      </DsFormItem>

      <DsFormItem label="所在房间" v-if="rooms.length > 0">
        <picker
          :range="rooms"
          range-key="name"
          :value="roomIndex"
          @change="onRoomChange"
        >
          <view class="picker-value">
            <text v-if="form.room_id">{{ selectedRoomName }}</text>
            <text v-else class="placeholder">请选择房间</text>
            <DsIcon name="next" :size="16" color="#999" />
          </view>
        </picker>
      </DsFormItem>

      <DsFormItem label="所在房间" v-else>
        <text class="placeholder">暂无房间，请先创建</text>
      </DsFormItem>
    </DsCard>

    <view class="footer-bar" v-if="!loading">
      <DsButton text="保存" type="primary" size="large" :loading="saving" @click="handleSave" />
    </view>
  </view>
</template>

<script setup>
import { ref, computed } from 'vue'
import { onLoad } from '@dcloudio/uni-app'
import { getDeviceDetail, updateDevice } from '@/api/device'
import { getRoomList } from '@/api/family'
import { currentFamily, loadFamilyData } from '@/utils/user'
import DsNavBar from '@/components/DsNavBar.vue'
import DsCard from '@/components/DsCard.vue'
import DsButton from '@/components/DsButton.vue'
import DsFormItem from '@/components/DsFormItem.vue'
import DsIcon from '@/components/DsIcon.vue'
import DsLoading from '@/components/DsLoading.vue'
import DsInput from '@/components/DsInput.vue'

const loading = ref(true)
const saving = ref(false)
const deviceId = ref('')

const form = ref({
  name: '',
  room_id: ''
})

const rooms = ref([])

const roomIndex = computed(() => {
  const idx = rooms.value.findIndex(r => r.id === form.value.room_id)
  return idx >= 0 ? idx : 0
})

const selectedRoomName = computed(() => {
  const room = rooms.value.find(r => r.id === form.value.room_id)
  return room?.name || ''
})

const onRoomChange = (e) => {
  const idx = e.detail.value
  const room = rooms.value[idx]
  form.value.room_id = room?.id || ''
}

const loadDevice = async (id) => {
  try {
    const res = await getDeviceDetail(id)
    const data = res.data || res
    form.value.name = data.name || ''
    form.value.room_id = data.room_id || ''
  } catch (e) {
    console.error(e)
    uni.showToast({ title: '加载失败', icon: 'none' })
  }
}

const loadRooms = async () => {
  if (!currentFamily.value?.id) return
  try {
    const res = await getRoomList(currentFamily.value.id)
    rooms.value = res.data || []
  } catch (e) {
    console.error(e)
  }
}

const handleSave = async () => {
  if (!form.value.name.trim()) {
    uni.showToast({ title: '请输入设备名称', icon: 'none' })
    return
  }

  saving.value = true
  try {
    await updateDevice(deviceId.value, {
      name: form.value.name.trim(),
      room_id: form.value.room_id
    })
    uni.showToast({ title: '保存成功', icon: 'success' })
    setTimeout(() => uni.navigateBack(), 1500)
  } catch (e) {
    console.error(e)
    uni.showToast({ title: '保存失败', icon: 'none' })
  } finally {
    saving.value = false
  }
}

onLoad(async (options) => {
  deviceId.value = options?.id || ''
  loading.value = true
  if (!currentFamily.value?.id && uni.getStorageSync('token')) {
    await loadFamilyData()
  }
  await Promise.all([loadDevice(deviceId.value), loadRooms()])
  loading.value = false
})
</script>

<style lang="scss">
@import '@/components/variables.scss';

page {
  background: $bg-secondary;
}

.edit-container {
  padding-bottom: 160rpx;
}

.form-input {
  flex: 1;
  font-size: $font-size-base;
  color: $text-primary;
  text-align: right;
}

.picker-value {
  display: flex;
  align-items: center;
  gap: 8rpx;
  font-size: $font-size-base;
  color: $text-primary;

  .placeholder {
    color: $text-placeholder;
  }
}

.footer-bar {
  position: fixed;
  bottom: 0;
  left: 0;
  right: 0;
  background: $bg-primary;
  padding: $spacing-base $spacing-lg;
  padding-bottom: calc($spacing-base + env(safe-area-inset-bottom));
  box-shadow: 0 -4rpx 20rpx rgba(0, 0, 0, 0.05);
  z-index: 99;
}
</style>
