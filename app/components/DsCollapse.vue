<template>
  <view class="ds-collapse">
    <slot></slot>
  </view>
</template>

<script setup>
import { ref, provide } from 'vue'

defineOptions({
  options: {
    virtualHost: true
  }
})

const props = defineProps({
  modelValue: {
    type: [Number, Array],
    default: 0
  },
  accordion: {
    type: Boolean,
    default: false
  }
})

const emit = defineEmits(['update:modelValue', 'change'])

const activeNames = ref(Array.isArray(props.modelValue) ? props.modelValue : [props.modelValue])

const toggle = (index) => {
  if (props.accordion) {
    activeNames.value = index === activeNames.value[0] ? [] : [index]
  } else {
    const idx = activeNames.value.indexOf(index)
    if (idx > -1) {
      activeNames.value.splice(idx, 1)
    } else {
      activeNames.value.push(index)
    }
  }
  emit('update:modelValue', props.accordion ? activeNames.value[0] : activeNames.value)
  emit('change', activeNames.value)
}

provide('collapseContext', {
  activeNames,
  toggle
})
</script>

<style lang="scss">
@import './variables.scss';

.ds-collapse {
  background: $bg-primary;
  border-radius: $radius-base;
  overflow: hidden;
}
</style>