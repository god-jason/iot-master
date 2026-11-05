<template>
  <view class="ds-search-bar">
    <view class="ds-search-input-wrap">
      <view v-if="prefixIcon" class="ds-search-icon ds-prefix" @click="handlePrefixClick">
        <DsIcon :name="prefixIcon" :size="prefixIconSize" :color="prefixIconColor" />
      </view>
      <DsIcon v-else name="search" :size="18" color="#999" />
      <input
        :value="modelValue"
        :placeholder="placeholder"
        :confirm-type="confirmType"
        :adjust-position="adjustPosition"
        class="ds-search-input"
        placeholder-class="ds-search-placeholder"
        @input="handleInput"
        @focus="handleFocus"
        @blur="handleBlur"
        @confirm="handleConfirm"
      />
      <view v-if="clearable && modelValue" class="ds-search-clear" @click="handleClear">
        <DsIcon name="close" size="22" color="#999" />
      </view>
      <view v-else-if="suffixIcon" class="ds-search-icon ds-suffix" @click="handleSuffixClick">
        <DsIcon :name="suffixIcon" :size="suffixIconSize" :color="suffixIconColor" />
      </view>
    </view>
    <view v-if="actionText" class="ds-search-action" @click="handleAction">
      <text>{{ actionText }}</text>
    </view>
    <view v-else-if="actionIcon" class="ds-search-action-icon" @click="handleAction">
      <DsIcon :name="actionIcon" :size="28" color="#5BA3E8" />
    </view>
  </view>
</template>

<script setup>
import DsIcon from './DsIcon.vue'

defineOptions({
  options: {
    virtualHost: true
  }
})

const props = defineProps({
  modelValue: {
    type: String,
    default: ''
  },
  placeholder: {
    type: String,
    default: '请输入搜索内容'
  },
  clearable: {
    type: Boolean,
    default: true
  },
  actionText: {
    type: String,
    default: ''
  },
  actionIcon: {
    type: String,
    default: ''
  },
  confirmType: {
    type: String,
    default: 'search'
  },
  adjustPosition: {
    type: Boolean,
    default: true
  },
  prefixIcon: {
    type: String,
    default: ''
  },
  prefixIconSize: {
    type: Number,
    default: 18
  },
  prefixIconColor: {
    type: String,
    default: '#999'
  },
  suffixIcon: {
    type: String,
    default: ''
  },
  suffixIconSize: {
    type: Number,
    default: 22
  },
  suffixIconColor: {
    type: String,
    default: '#999'
  }
})

const emit = defineEmits(['update:modelValue', 'focus', 'blur', 'confirm', 'action', 'prefix-click', 'suffix-click'])

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

const handleAction = () => {
  emit('action')
}

const handlePrefixClick = () => {
  emit('prefix-click')
}

const handleSuffixClick = () => {
  emit('suffix-click')
}
</script>

<style lang="scss">
@import './variables.scss';

.ds-search-bar {
  display: flex;
  align-items: center;
  gap: $spacing-sm;
  padding: $spacing-xs $spacing-sm;
  background: $bg-primary;

  .ds-search-input-wrap {
    flex: 1;
    display: flex;
    align-items: center;
    gap: $spacing-xs;
    padding: $spacing-xs $spacing-sm;
    background: $bg-secondary;
    border-radius: $radius-base;

    .ds-search-input {
      flex: 1;
      font-size: $font-size-base;
      color: $text-primary;
      background: transparent;
    }

    .ds-search-placeholder {
      color: $text-placeholder;
    }

    .ds-search-icon {
      flex-shrink: 0;
      padding: $spacing-xs;

      &.ds-prefix {
        margin-right: $spacing-xs;
      }

      &.ds-suffix {
        margin-left: $spacing-xs;
      }
    }

    .ds-search-clear {
      width: 44rpx;
      height: 44rpx;
      padding: 0;
      margin-left: $spacing-sm;
      display: flex;
      align-items: center;
      justify-content: center;
    }
  }

  .ds-search-action {
    padding: $spacing-sm $spacing-base;
    font-size: $font-size-base;
    color: $primary-color;
    font-weight: 500;
  }

  .ds-search-action-icon {
    padding: $spacing-sm;
  }
}
</style>