<template>
  <view class="page-device-edit">
    <DsNavBar title="编辑设备" />
    <DsCard>
      <DsForm>
        <DsFormItem label="设备名称">
          <DsInput v-model="form.name" placeholder="请输入设备名称" />
        </DsFormItem>
        
        <DsFormItem label="设备编号">
          <DsInput v-model="form.sn" placeholder="请输入设备编号" disabled />
        </DsFormItem>
        
        <DsFormItem label="所属房间">
          <DsComboBox :options="roomOptions" :value="form.room_id" label="请选择房间" @change="form.room_id = $event" />
        </DsFormItem>
      </DsForm>
    </DsCard>
    
    <view class="submit-area">
      <DsButton text="保存修改" type="primary" size="large" @click="submit" />
    </view>
  </view>
</template>

<script setup>
import { ref, reactive, computed } from 'vue'
import { onLoad } from '@dcloudio/uni-app'
import { getDeviceDetail, updateDevice } from '@/api/device'
import { getRoomList } from '@/api/family'
import { currentFamily, loadFamilyData } from '@/utils/user'
import DsNavBar from '@/components/DsNavBar.vue'
import DsCard from '@/components/DsCard.vue'
import DsForm from '@/components/DsForm.vue'
import DsFormItem from '@/components/DsFormItem.vue'
import DsInput from '@/components/DsInput.vue'
import DsComboBox from '@/components/DsComboBox.vue'
import DsButton from '@/components/DsButton.vue'
const deviceId = ref('')
const form = reactive({
  name: '',
  sn: '',
  room_id: ''
})

const roomOptions = ref([])

const submit = async () => {
  if (!form.name.trim()) {
    uni.showToast({ title: '请输入设备名称', icon: 'none' })
    return
  }

  uni.showLoading({ title: '保存中...' })

  try {
    const data = {
      name: form.name.trim(),
      room_id: form.room_id || null
    }
    await updateDevice(deviceId.value, data)
    
    uni.hideLoading()
    uni.showToast({ title: '保存成功', icon: 'success' })
    setTimeout(() => {
      uni.navigateBack()
    }, 1500)
  } catch (error) {
    uni.hideLoading()
    uni.showToast({ title: error.message || '保存失败', icon: 'none' })
  }
}

onLoad(async (options) => {
  if (options.id) {
    deviceId.value = options.id

    try {
      const res = await getDeviceDetail(options.id)
      const device = res.data || res
      form.name = device.name || ''
      form.sn = device.sn || ''
      form.room_id = device.room_id || ''
    } catch (error) {
      console.error('获取设备详情失败:', error)
    }

    const familyId = currentFamily.value?.id || (await loadFamilyData(), currentFamily.value?.id)
    if (familyId) {
      try {
        const roomRes = await getRoomList(familyId)
        const rooms = roomRes.data || []
        roomOptions.value = rooms.map(r => ({
          value: r.id,
          label: r.name || r.id
        }))
      } catch (error) {
        console.error('获取房间列表失败:', error)
      }
    }
  }
})
</script>

<style lang="scss">
@import '@/components/variables.scss';

.container {
  min-height: 100vh;
  background: $bg-secondary;
  padding-bottom: 160rpx;
}

.form-textarea {
  width: 100%;
  height: 160rpx;
  font-size: $font-size-base;
  color: $text-primary;
}

.submit-area {
  position: fixed;
  bottom: 0;
  left: 0;
  right: 0;
  padding: $spacing-base;
  background: $bg-primary;
  box-shadow: 0 -2rpx 10rpx rgba(0, 0, 0, 0.05);
}
</style>
