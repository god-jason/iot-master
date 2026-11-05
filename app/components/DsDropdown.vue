<template>
  <view class="ds-dropdown">
    <view class="ds-dropdown-trigger" @click.stop="handleToggle">
      <text class="ds-dropdown-text">{{ selectedLabel }}</text>
      <DsIcon name="down_arrow" :size="20" color="#999" :class="{ 'ds-rotate': visible }" />
    </view>
    <view v-if="visible" class="ds-dropdown-menu" :class="directionClass" @click.stop>
      <view 
        v-for="(option, index) in options" 
        :key="index"
        class="ds-dropdown-item"
        :class="{ 'ds-active': modelValue === option.value }"
        @click.stop="handleSelect(option)"
      >
        <text>{{ option.label }}</text>
        <DsIcon v-if="modelValue === option.value" name="affirm" :size="24" color="#5BA3E8" />
      </view>
    </view>
    <view v-if="visible" class="ds-dropdown-mask" @click.stop="handleClose"></view>
  </view>
</template>

<script setup>
import { ref, computed, watch } from 'vue'
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
  options: {
    type: Array,
    default: () => []
  },
  placeholder: {
    type: String,
    default: '请选择'
  },
  direction: {
    type: String,
    default: 'down',
    validator: (val) => ['up', 'down'].includes(val)
  }
})

const emit = defineEmits(['update:modelValue', 'change'])

const visible = ref(false)

const directionClass = computed(() => `ds-direction-${props.direction}`)

const selectedLabel = computed(() => {
  const option = props.options.find(opt => opt.value === props.modelValue)
  return option ? option.label : props.placeholder
})

watch(() => props.modelValue, () => {
  visible.value = false
})

const handleToggle = () => {
  visible.value = !visible.value
}

const handleClose = () => {
  visible.value = false
}

const handleSelect = (option) => {
  emit('update:modelValue', option.value)
  emit('change', option)
  visible.value = false
}
</script>

<style lang="scss">
@import './variables.scss';

.ds-dropdown {
  position: relative;
  display: inline-block;

  .ds-dropdown-trigger {
    display: flex;
    align-items: center;
    gap: $spacing-xs;
    padding: $spacing-sm $spacing-base;
    background: $bg-secondary;
    border-radius: $radius-base;

    .ds-dropdown-text {
      font-size: $font-size-base;
      color: $text-primary;
    }

    .ds-rotate {
      transform: rotate(180deg);
    }
  }

  .ds-dropdown-menu {
    position: absolute;
    top: 100%;
    left: 0;
    right: 0;
    margin-top: $spacing-xs;
    background: $bg-primary;
    border-radius: $radius-base;
    box-shadow: $shadow-lg;
    z-index: 100;
    overflow: hidden;

    &.ds-direction-up {
      top: auto;
      bottom: 100%;
      margin-top: 0;
      margin-bottom: $spacing-xs;
    }

    .ds-dropdown-item {
      display: flex;
      align-items: center;
      justify-content: space-between;
      padding: $spacing-base;
      font-size: $font-size-base;
      color: $text-primary;

      &:active {
        background: $bg-secondary;
      }

      &.ds-active {
        color: $primary-color;
        background: $primary-light;
      }
    }
  }

  .ds-dropdown-mask {
    position: fixed;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    z-index: 99;
  }
}
</style>