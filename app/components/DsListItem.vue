<template>
  <view class="ds-list-item" :class="[`ds-align-${align}`, { 'ds-clickable': clickable, 'ds-active': active }]" @click="handleClick">
    <view v-if="$slots.left" class="ds-list-item-left">
      <slot name="left"></slot>
    </view>
    <view v-else-if="icon" class="ds-list-item-icon">
      <DsIcon :name="icon" :size="iconSize" :color="iconColor" />
    </view>
    <view v-else-if="avatar" class="ds-list-item-avatar">
      <DsAvatar :src="avatar" :size="avatarSize" />
    </view>
    <view class="ds-list-item-content">
      <view class="ds-list-item-title">
        <slot name="title">
        <text>{{ title }}</text>
      </slot>
      <DsTag v-if="tag" :text="tag.text" :type="tag.type" size="small" />
    </view>
    <slot name="desc">
      <slot name="subtitle">
        <text v-if="desc" class="ds-list-item-desc">{{ desc }}</text>
      </slot>
    </slot>
      <slot name="content"></slot>
    </view>
    <view class="ds-list-item-extra">
      <text v-if="extra" class="ds-list-item-extra-text">{{ extra }}</text>
      <slot name="right"></slot>
      <slot name="extra"></slot>
      <DsIcon v-if="arrow" name="next" :size="20" color="#ccc" />
    </view>
  </view>
</template>

<script setup>
import DsIcon from './DsIcon.vue'
import DsAvatar from './DsAvatar.vue'
import DsTag from './DsTag.vue'

defineOptions({
  options: {
    virtualHost: true
  }
})

const props = defineProps({
  title: {
    type: String,
    default: ''
  },
  desc: {
    type: String,
    default: ''
  },
  extra: {
    type: String,
    default: ''
  },
  icon: {
    type: String,
    default: ''
  },
  iconSize: {
    type: Number,
    default: 24
  },
  iconColor: {
    type: String,
    default: '#666'
  },
  avatar: {
    type: String,
    default: ''
  },
  avatarSize: {
    type: String,
    default: 'small'
  },
  tag: {
    type: Object,
    default: null
  },
  arrow: {
    type: Boolean,
    default: false
  },
  align: {
    type: String,
    default: 'center',
    validator: (val) => ['top', 'center', 'bottom'].includes(val)
  },
  active: {
    type: Boolean,
    default: false
  },
  clickable: {
    type: Boolean,
    default: true
  }
})

const emit = defineEmits(['click'])

const handleClick = (event) => {
  emit('click', event)
}
</script>

<style lang="scss">
@import './variables.scss';

.ds-list-item {
  display: flex;
  align-items: center;
  padding: $spacing-lg;
  border-bottom: 1rpx solid $border-light;
  transition: $transition-base;

  &.ds-align-top {
    align-items: flex-start;
  }

  &.ds-align-bottom {
    align-items: flex-end;
  }

  &:last-child {
    border-bottom: none;
  }

  &.ds-clickable {
    &:active {
      background: $bg-secondary;
    }
  }

  &.ds-active {
    background: $primary-light;
  }

  .ds-list-item-left,
  .ds-list-item-icon {
    margin-right: $spacing-base;
    flex-shrink: 0;
  }

  .ds-list-item-avatar {
    margin-right: $spacing-base;
    flex-shrink: 0;
  }

  .ds-list-item-content {
    flex: 1;
    min-width: 0;

    .ds-list-item-title {
      display: flex;
      align-items: center;
      gap: $spacing-xs;
      font-size: $font-size-base;
      font-weight: 500;
      color: $text-primary;
      margin-bottom: $spacing-xs;
    }

    .ds-list-item-desc {
      font-size: $font-size-sm;
      color: $text-placeholder;
      overflow: hidden;
      text-overflow: ellipsis;
      white-space: nowrap;
    }
  }

  .ds-list-item-extra {
    display: flex;
    align-items: center;
    gap: $spacing-xs;
    flex-shrink: 0;

    .ds-list-item-extra-text {
      font-size: $font-size-sm;
      color: $text-secondary;
    }
  }
}
</style>