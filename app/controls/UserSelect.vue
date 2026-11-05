<template>
  <view class="control-mode-select">
    <text class="mode-label">{{ control.name }}</text>
    <view class="mode-options">
      <view
        v-for="(option, index) in control.options"
        :key="index"
        class="mode-option"
        :class="{ active: currentValue === option.value }"
        :style="optionStyle"
        @tap="select(option.value)"
      >
        <Icon 
          v-if="option.icon" 
          :name="option.icon" 
          class="option-icon" 
        />
        <text class="option-label">{{ option.label }}</text>
      </view>
    </view>
  </view>
</template>

<script setup>
import { ref, computed, watch } from 'vue'
import { toStr } from './User.js'
import Icon from './Icon.vue'

const props = defineProps({
  control: {
    type: Object,
    required: true
  },
  value: {
    type: [String, Number, Boolean],
    default: ''
  }
})

const emit = defineEmits(['change', 'action'])

const currentValue = ref(toStr(props.value))

watch(() => props.value, (val) => {
  currentValue.value = toStr(val)
})

const optionStyle = computed(() => {
  const style = props.control.styles?.default || {}
  return {
    '--active-bg': style.color || '#5BA3E8',
    borderRadius: style.border_radius || '24rpx'
  }
})

const select = (value) => {
  currentValue.value = value
  const bindKey = props.control.binds?.value

  if (props.control.events?.change?.key) {
    emit('action', {
      type: 'select',
      controlId: props.control.id,
      eventKey: 'change'
    })
  }

  emit('change', {
    type: 'select',
    value,
    controlId: props.control.id,
    bindKey: bindKey
  })
}
</script>

<style lang="scss" scoped>
.control-mode-select {
  min-height: 120rpx;

  .mode-label {
    font-size: 28rpx;
    color: #333;
    margin-bottom: 24rpx;
    display: block;
  }

  .mode-options {
    display: flex;
    gap: 24rpx;

    .mode-option {
      flex: 1;
      display: flex;
      flex-direction: column;
      align-items: center;
      gap: 16rpx;
      padding: 32rpx 16rpx;
      background: #f5f5f5;
      border-radius: 24rpx;
      transition: background 0.3s;
      user-select: none;
      -webkit-tap-highlight-color: transparent;

      &.active {
        background: var(--active-bg);
        color: #fff;
      }

      .option-icon {
        font-size: 48rpx;
      }

      .option-label {
        font-size: 24rpx;
      }
    }
  }
}
</style>
