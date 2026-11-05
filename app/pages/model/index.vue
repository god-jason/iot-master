<template>
  <view class="page-model-index">
    <DsNavBar title="物模型管理" />

    <DsSearchBar v-model="searchText" placeholder="搜索物模型" @confirm="handleSearch" />

    <view class="model-content">
      <DsCard v-if="models.length > 0 || loading">
        <DsList>
          <DsListItem
            v-for="model in models"
            :key="model.id" arrow
            @click="goToDetail(model.id)"
          >
            <template #left>
              <view class="model-icon">
                <image v-if="model.icon" :src="model.icon" mode="aspectFill" class="icon-img" />
                <DsIcon v-else name="cpu" :size="24" color="#5BA3E8" />
              </view>
            </template>
            <template #title>
              <text class="model-name">{{ model.name }}</text>
            </template>
            <template #subtitle>
              <text class="model-desc">{{ model.description || '暂无描述' }}</text>
            </template>
            <template #extra>
              <view class="model-meta">
                <DsTag type="default" size="small">{{ model.category || '默认分类' }}</DsTag>
              </view>
            </template>
          </DsListItem>
        </DsList>
      </DsCard>

        <DsLoading v-if="loading" text="加载中..." />
        <DsNoMore v-if="!hasMore && models.length > 0" />

      <DsEmpty v-if="!loading && models.length === 0" text="暂无物模型">
        <template #icon>
          <DsIcon name="cpu" :size="24" color="#ccc" />
        </template>
      </DsEmpty>
    </view>
  </view>
</template>

<script setup>
import { ref } from 'vue'
import { onShow, onPullDownRefresh, onReachBottom } from '@dcloudio/uni-app'
import { getModelList } from '@/api/model'
import DsNavBar from '@/components/DsNavBar.vue'
import DsIcon from '@/components/DsIcon.vue'
import DsSearchBar from '@/components/DsSearchBar.vue'
import DsCard from '@/components/DsCard.vue'
import DsList from '@/components/DsList.vue'
import DsListItem from '@/components/DsListItem.vue'
import DsTag from '@/components/DsTag.vue'
import DsLoading from '@/components/DsLoading.vue'
import DsEmpty from '@/components/DsEmpty.vue'
import DsNoMore from '@/components/DsNoMore.vue'
const models = ref([])
const loading = ref(false)
const page = ref(1)
const pageSize = 20
const hasMore = ref(true)
const total = ref(0)
const searchText = ref('')

const loadModels = async (reset = false) => {
  if (loading.value) return
  if (reset) {
    page.value = 1
    hasMore.value = true
  }
  if (!hasMore.value) return

  loading.value = true
  try {
    const params = {
      page: page.value,
      page_size: pageSize
    }
    if (searchText.value) params.keyword = searchText.value

    const res = await getModelList(params)
    const data = res.data || res
    const list = data.list || data || []
    if (reset) {
      models.value = list
    } else {
      models.value = [...models.value, ...list]
    }
    total.value = data.total || 0
    hasMore.value = models.value.length < total.value
    if (hasMore.value) page.value++
  } catch (e) {
    console.error(e)
  } finally {
    loading.value = false
  }
}

const handleSearch = () => {
  loadModels(true)
}

const goToDetail = (id) => {
  uni.navigateTo({
    url: `/pages/model/detail?id=${id}`
  })
}

onShow(() => {
  loadModels(true)
})

onPullDownRefresh(async () => {
  await loadModels(true)
  uni.stopPullDownRefresh()
})

onReachBottom(() => {
  loadModels()
})
</script>

<style lang="scss">
@import '@/components/variables.scss';

page {
  background: $bg-secondary;
}

.model-container {
  min-height: 100vh;
}

.model-content {
}

.model-icon {
  width: 100rpx;
  height: 100rpx;
  border-radius: $radius-lg;
  background: $bg-secondary;
  display: flex;
  align-items: center;
  justify-content: center;
  overflow: hidden;

  .icon-img {
    width: 100%;
    height: 100%;
  }
}

.model-name {
  font-size: $font-size-lg;
  font-weight: 600;
  color: $text-primary;
}

.model-desc {
  font-size: $font-size-sm;
  color: $text-placeholder;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.model-meta {
  display: flex;
  align-items: center;
  gap: 12rpx;
}

.device-count {
  font-size: $font-size-xs;
  color: $text-secondary;
}

</style>
