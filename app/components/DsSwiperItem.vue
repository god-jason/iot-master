<template>
  <swiper-item @click="handleClick">
    <slot></slot>
  </swiper-item>
</template>

<script setup>
defineOptions({
  options: {
    virtualHost: true
  }
})

const emit = defineEmits(['click'])

const handleClick = () => {
  emit('click')
}
</script>

<style lang="scss">
@import './variables.scss';
</style>