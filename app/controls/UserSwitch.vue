<template>
  <view 
    class="control-switch"
  >
    <Icon 
      v-if="iconStyle.icon" 
      :name="iconStyle.icon" 
      :color="iconStyle.color"
      :size="iconStyle.font_size"
      class="switch-icon" 
    />
    <text class="switch-label">{{ control.name }}</text>
    <view 
      class="switch-wrapper"
      :class="{ active: currentValue }"
      :style="switchStyle"
      @tap="toggle"
    >
      <view class="switch-track">
        <view class="switch-thumb"></view>
      </view>
    </view>
  </view>
</template>

<script setup>
import { ref, computed, watch } from 'vue'
import { toBool } from './User.js'
import Icon from './Icon.vue'

const props = defineProps({
  control: {
    type: Object,
    required: true
  },
  value: {
    type: [String, Boolean, Number],
    default: false
  }
})

const emit = defineEmits(['change', 'action'])

const iconStyle = computed(() => {
  const style = props.control.styles?.icon || {}
  return {
    icon: style.icon || '',
    color: style.color || '#999',
    font_size: style.font_size || '40rpx'
  }
})

const currentValue = ref(toBool(props.value))

watch(() => props.value, (val) => {
  currentValue.value = toBool(val)
})

const switchStyle = computed(() => {
  const style = props.control.styles?.default || {}
  const activeColor = style.color || '#5BA3E8'
  return {
    '--active-color': activeColor
  }
})

const toggle = () => {
  currentValue.value = !currentValue.value
  const bindKey = props.control.binds?.value
  emit('change', {
    type: 'toggle',
    value: currentValue.value,
    controlId: props.control.id,
    bindKey: bindKey
  })
  if (props.control.events?.toggle?.key) {
    emit('action', {
      type: 'toggle',
      controlId: props.control.id,
      eventKey: 'toggle'
    })
  }
}
</script>

<style lang="scss" scoped>
.control-switch {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 24rpx;
  width: 100%;
  height: 100%;

  .switch-icon {
    font-size: 40rpx;
  }

  .switch-label {
    font-size: 28rpx;
    color: #333;
    flex: 1;
  }

  .switch-wrapper {
    display: flex;
    align-items: center;
    flex-shrink: 0;

    .switch-track {
      width: 140rpx;
      height: 76rpx;
      background: #e0e0e0;
      border-radius: 38rpx;
      position: relative;
      transition: background 0.3s;

      .switch-thumb {
        width: 68rpx;
        height: 68rpx;
        background: #fff;
        border-radius: 50%;
        position: absolute;
        top: 4rpx;
        left: 4rpx;
        transition: left 0.3s;
        box-shadow: 0 4rpx 8rpx rgba(0,0,0,0.2);
      }
    }

    &.active {
      .switch-track {
        background: var(--active-color);
      }

      .switch-thumb {
        left: 68rpx;
      }
    }
  }
}
</style>
