<template>
  <view class="ds-table">
    <view v-if="columns.length > 0" class="ds-table-header">
      <view v-for="col in columns" :key="col.key" class="ds-table-header-cell" :style="{ width: col.width || 'auto' }">
        <text>{{ col.title }}</text>
      </view>
    </view>
    <view class="ds-table-body">
      <view v-for="(row, rowIndex) in data" :key="rowIndex" class="ds-table-row">
        <view v-for="col in columns" :key="col.key" class="ds-table-cell" :style="{ width: col.width || 'auto' }">
          <slot :name="`cell-${col.key}`" :value="row[col.key]" :row="row" :index="rowIndex">
            {{ row[col.key] !== undefined && row[col.key] !== null ? row[col.key] : '-' }}
          </slot>
        </view>
      </view>
      <view v-if="data.length === 0" class="ds-table-empty">
        <text>{{ emptyText || '暂无数据' }}</text>
      </view>
    </view>
  </view>
</template>

<script setup>
defineOptions({
  options: {
    virtualHost: true
  }
})

defineProps({
  columns: {
    type: Array,
    default: () => []
  },
  data: {
    type: Array,
    default: () => []
  },
  emptyText: {
    type: String,
    default: '暂无数据'
  }
})
</script>

<style lang="scss">
@import './variables.scss';

.ds-table {
  background: $bg-primary;
  border-radius: $radius-lg;
  overflow: hidden;

  .ds-table-header {
    display: flex;
    background: $bg-secondary;
    padding: $spacing-sm 0;

    .ds-table-header-cell {
      padding: $spacing-sm $spacing-lg;
      font-size: $font-size-sm;
      font-weight: 600;
      color: $text-secondary;
      flex: 1;
      min-width: 0;
    }
  }

  .ds-table-body {
    .ds-table-row {
      display: flex;
      border-bottom: 1rpx solid $border-light;

      &:last-child {
        border-bottom: none;
      }

      .ds-table-cell {
        padding: $spacing-base $spacing-lg;
        font-size: $font-size-sm;
        color: $text-primary;
        flex: 1;
        min-width: 0;
        overflow: hidden;
        text-overflow: ellipsis;
        white-space: nowrap;
      }
    }

    .ds-table-empty {
      padding: $spacing-xl;
      text-align: center;
      font-size: $font-size-sm;
      color: $text-placeholder;
    }
  }
}
</style>