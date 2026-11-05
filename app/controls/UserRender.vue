<template>
  <UserButton
    v-if="control.type === 'button'"
    :control="control"
    :value="value"
    @change="(e) => emit('change', e)"
    @action="(e) => emit('action', e)"
  />
  <UserSwitch
    v-else-if="control.type === 'switch'"
    :control="control"
    :value="value"
    @change="(e) => emit('change', e)"
    @action="(e) => emit('action', e)"
  />
  <UserSlider
    v-else-if="control.type === 'slider'"
    :control="control"
    :value="value"
    @change="(e) => emit('change', e)"
    @action="(e) => emit('action', e)"
  />
  <UserBrightness
    v-else-if="control.type === 'brightness'"
    :control="control"
    :value="value"
    @change="(e) => emit('change', e)"
    @action="(e) => emit('action', e)"
  />
  <UserSelect
    v-else-if="control.type === 'mode'"
    :control="control"
    :value="value"
    @change="(e) => emit('change', e)"
    @action="(e) => emit('action', e)"
  />
  <UserColorPicker
    v-else-if="control.type === 'color'"
    :control="control"
    :value="value"
    @change="(e) => emit('change', e)"
    @action="(e) => emit('action', e)"
  />
  <UserColorTemperature
    v-else-if="control.type === 'color-temperature'"
    :control="control"
    :value="value"
    @change="(e) => emit('change', e)"
    @action="(e) => emit('action', e)"
  />
  <UserColorHue
    v-else-if="control.type === 'color-hue'"
    :control="control"
    :value="value"
    @change="(e) => emit('change', e)"
    @action="(e) => emit('action', e)"
  />
  <UserNumber
    v-else-if="control.type === 'number'"
    :control="control"
    :value="value"
    @change="(e) => emit('change', e)"
    @action="(e) => emit('action', e)"
  />
  <UserText
    v-else-if="control.type === 'text'"
    :control="control"
    :value="value"
    @change="(e) => emit('change', e)"
    @action="(e) => emit('action', e)"
  />
  <UserVideo
    v-else-if="control.type === 'video'"
    :control="control"
    :value="value"
    @change="(e) => emit('change', e)"
    @action="(e) => emit('action', e)"
  />
  <UserImage
    v-else-if="control.type === 'image'"
    :control="control"
    :value="value"
    @change="(e) => emit('change', e)"
    @action="(e) => emit('action', e)"
  />
  <view v-else class="unknown-control">未知控件</view>
</template>

<script setup>
import UserButton from './UserButton.vue'
import UserSwitch from './UserSwitch.vue'
import UserSlider from './UserSlider.vue'
import UserBrightness from './UserBrightness.vue'
import UserSelect from './UserSelect.vue'
import UserColorPicker from './UserColorPicker.vue'
import UserColorTemperature from './UserColorTemperature.vue'
import UserColorHue from './UserColorHue.vue'
import UserNumber from './UserNumber.vue'
import UserText from './UserText.vue'
import UserVideo from './UserVideo.vue'
import UserImage from './UserImage.vue'

// 定义组件选项，启用虚拟主机模式，解决事件不响应的问题
defineOptions({
  options: {
    virtualHost: true
  }
})

const props = defineProps({
  control: {
    type: Object,
    required: true
  },
  value: {
    type: [String, Number, Boolean, Object, Array],
    default: ''
  }
})

const emit = defineEmits(['change', 'action'])
</script>

<style lang="scss" scoped>
.unknown-control {
  padding: 24rpx;
  color: #999;
  font-size: 28rpx;
  text-align: center;
}
</style>
