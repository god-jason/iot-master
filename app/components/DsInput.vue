<template>
  <view class="ds-input-wrapper">
    <view v-if="label" class="ds-input-label">{{ label }}</view>
    <view class="ds-input-inner" :class="{ 'ds-error': hasError }">
      <DsIcon v-if="prefixIcon" :name="prefixIcon" :size="22" color="#999" class="ds-input-icon ds-prefix" />
      <input
        :value="modelValue"
        :type="inputType"
        :placeholder="placeholder"
        :password="type === 'password'"
        :confirm-type="confirmType"
        :adjust-position="adjustPosition"
        :maxlength="maxlength"
        :disabled="disabled"
        :readonly="readonly"
        :auto-focus="autoFocus"
        class="ds-input-field"
        placeholder-class="ds-input-placeholder"
        @input="handleInput"
        @focus="handleFocus"
        @blur="handleBlur"
        @confirm="handleConfirm"
      />
      <view v-if="clearable && modelValue" class="ds-input-clear" @click="handleClear">
        <DsIcon name="close" size="22" color="#999" />
      </view>
      <DsIcon v-else-if="suffixIcon" :name="suffixIcon" :size="22" color="#999" class="ds-input-icon ds-suffix" />
    </view>
    <text v-if="hasError && errorMessage" class="ds-input-error">{{ errorMessage }}</text>
    <text v-if="helpText" class="ds-input-help">{{ helpText }}</text>
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
    default: 'text',
    validator: (val) => ['text', 'password', 'number', 'digit', 'idcard', 'textarea'].includes(val)
  },
  label: {
    type: String,
    default: ''
  },
  placeholder: {
    type: String,
    default: ''
  },
  prefixIcon: {
    type: String,
    default: ''
  },
  suffixIcon: {
    type: String,
    default: ''
  },
  clearable: {
    type: Boolean,
    default: false
  },
  maxlength: {
    type: Number,
    default: -1
  },
  disabled: {
    type: Boolean,
    default: false
  },
  readonly: {
    type: Boolean,
    default: false
  },
  autoFocus: {
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
  },
  error: {
    type: Boolean,
    default: false
  },
  errorMessage: {
    type: String,
    default: ''
  },
  helpText: {
    type: String,
    default: ''
  }
})

const emit = defineEmits(['update:modelValue', 'focus', 'blur', 'confirm'])

const inputType = computed(() => {
  if (props.type === 'textarea') return 'text'
  if (props.type === 'password') return 'text'
  return props.type
})

const hasError = computed(() => props.error || !!props.errorMessage)

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

const handleClear = () => {
  emit('update:modelValue', '')
}
</script>

<style lang="scss">
@import './variables.scss';

.ds-input-wrapper {
  width: 100%;
  min-width: 0;

  .ds-input-label {
    font-size: $font-size-sm;
    color: $text-secondary;
    margin-bottom: $spacing-xs;
  }

  .ds-input-inner {
    display: flex;
    align-items: center;
    width: 100%;
    padding: $spacing-sm $spacing-base;
    background: $bg-primary;
    border: 2rpx solid $border-color;
    border-radius: $radius-base;
    box-sizing: border-box;
    transition: $transition-base;

    &.ds-error {
      border-color: $danger-color;
    }

    &:focus-within {
      border-color: $primary-color;
    }

    .ds-input-icon {
      flex-shrink: 0;

      &.ds-prefix {
        margin-right: $spacing-sm;
      }

      &.ds-suffix {
        margin-left: $spacing-sm;
      }
    }

    .ds-input-field {
      flex: 1;
      width: 0;
      min-width: 0;
      font-size: $font-size-base;
      color: $text-primary;
      background: transparent;
    }

    .ds-input-placeholder {
      color: $text-placeholder;
    }

    .ds-input-clear {
      width: 44rpx;
      height: 44rpx;
      padding: 0;
      margin-left: $spacing-sm;
      display: flex;
      align-items: center;
      justify-content: center;
    }
  }

  .ds-input-error {
    font-size: $font-size-xs;
    color: $danger-color;
    margin-top: $spacing-xs;
    display: block;
  }

  .ds-input-help {
    font-size: $font-size-xs;
    color: $text-placeholder;
    margin-top: $spacing-xs;
    display: block;
  }
}
</style>