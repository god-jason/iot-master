<template>
  <view class="ds-input-bar">
    <view v-if="resolvedLeftIcon" class="ds-input-bar-left">
      <DsIcon :name="resolvedLeftIcon" :size="32" :color="leftIconColor" />
    </view>
    <input
      :value="modelValue"
      :type="type"
      :placeholder="placeholder"
      :password="type === 'password'"
      :confirm-type="confirmType"
      :adjust-position="adjustPosition"
      :maxlength="maxlength"
      :disabled="disabled"
      class="ds-input-bar-input"
      placeholder-class="ds-input-bar-placeholder"
      @input="handleInput"
      @focus="handleFocus"
      @blur="handleBlur"
      @confirm="handleConfirm"
    />
    <view v-if="resolvedRightIcon" class="ds-input-bar-right" @click="handleRightClick">
      <DsIcon :name="resolvedRightIcon" :size="32" :color="rightIconColor" />
    </view>
    <view v-else-if="rightText" class="ds-input-bar-right ds-text" @click="handleRightClick">
      <text>{{ rightText }}</text>
    </view>
  </view>
</template>

<script setup>
import { computed } from 'vue'
import DsIcon from './DsIcon.vue'

defineOptions({
  options: {
    virtualHost: true
  }
})

const props = defineProps({
  modelValue: {
    type: [String, Number],
    default: ''
  },
  type: {
    type: String,
    default: 'text'
  },
  placeholder: {
    type: String,
    default: ''
  },
  leftIcon: {
    type: String,
    default: ''
  },
  prefixIcon: {
    type: String,
    default: ''
  },
  leftIconColor: {
    type: String,
    default: '#999'
  },
  rightIcon: {
    type: String,
    default: ''
  },
  suffixIcon: {
    type: String,
    default: ''
  },
  rightIconColor: {
    type: String,
    default: '#999'
  },
  rightText: {
    type: String,
    default: ''
  },
  maxlength: {
    type: Number,
    default: -1
  },
  disabled: {
    type: Boolean,
    default: false
  },
  adjustPosition: {
    type: Boolean,
    default: true
  },
  confirmType: {
    type: String,
    default: 'done'
  }
})

const emit = defineEmits(['update:modelValue', 'focus', 'blur', 'confirm', 'right-click'])

const resolvedLeftIcon = computed(() => props.leftIcon || props.prefixIcon)
const resolvedRightIcon = computed(() => props.rightIcon || props.suffixIcon)

const handleInput = (e) => {
  emit('update:modelValue', e.detail.value)
}

const handleFocus = (e) => {
  emit('focus', e)
}

const handleBlur = (e) => {
  emit('blur', e)
}

const handleConfirm = (e) => {
  emit('confirm', e)
}

const handleRightClick = () => {
  emit('right-click')
}
</script>

<style lang="scss">
@import './variables.scss';

.ds-input-bar {
  display: flex;
  align-items: center;
  width: 100%;
  padding: $spacing-sm $spacing-base;
  background: $bg-secondary;
  border-radius: $radius-lg;
  box-sizing: border-box;

  .ds-input-bar-left {
    padding-right: $spacing-sm;
  }

  .ds-input-bar-input {
    flex: 1;
    width: 0;
    min-width: 0;
    font-size: $font-size-base;
    color: $text-primary;
    background: transparent;
  }

  .ds-input-bar-placeholder {
    color: $text-placeholder;
  }

  .ds-input-bar-right {
    padding-left: $spacing-sm;
    flex-shrink: 0;

    &.ds-text {
      font-size: $font-size-sm;
      color: $primary-color;
    }
  }
}
</style>