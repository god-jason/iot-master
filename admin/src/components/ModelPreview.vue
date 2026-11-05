<template>
  <div class="model-preview">
    <div class="model-section">
      <div class="section-title">属性</div>
      <el-table :data="model.properties || []" border stripe size="small" v-if="model.properties?.length">
        <el-table-column prop="key" label="标识符" width="120" />
        <el-table-column prop="name" label="名称" width="120" />
        <el-table-column prop="type" label="数据类型" width="100" />
        <el-table-column prop="mode" label="读写模式" width="100">
          <template #default="{ row }">
            {{ row.mode === 'r' ? '只读' : row.mode === 'w' ? '只写' : row.mode === 'rw' ? '读写' : '-' }}
          </template>
        </el-table-column>
        <el-table-column v-if="hasProtocol" prop="register" label="寄存器" width="120" />
        <el-table-column v-if="hasProtocol" prop="format" label="格式" width="120" />
        <el-table-column v-if="hasProtocol" prop="rate" label="倍率" width="80" />
        <el-table-column v-if="hasProtocol" prop="address" label="地址" width="100" />
        <el-table-column prop="threshold" label="阈值" width="80" />
        <el-table-column prop="unit" label="单位" width="80" />
        <el-table-column prop="precision" label="精度" width="80" />
        <el-table-column v-if="hasProtocol" label="小端" width="70">
          <template #default="{ row }">
            {{ row.little_endian ? '是' : '否' }}
          </template>
        </el-table-column>
        <el-table-column v-if="hasProtocol" label="校准" width="80">
          <template #default="{ row }">
            {{ row.correct ?? '-' }}
          </template>
        </el-table-column>
        <el-table-column v-if="hasProtocol" label="取位" width="70">
          <template #default="{ row }">
            {{ row.bit > 0 ? row.bit : '-' }}
          </template>
        </el-table-column>
        <el-table-column v-if="hasProtocol" label="枚举">
          <template #default="{ row }">
            <template v-if="row.enums?.length">
              <el-tag v-for="e in row.enums" :key="e.key" size="small" style="margin-right: 4px">
                {{ e.name }}({{ e.value }})
              </el-tag>
            </template>
            <span v-else>-</span>
          </template>
        </el-table-column>
      </el-table>
      <el-empty v-else description="暂无属性" :image-size="60" />
    </div>

    <div v-if="hasProtocol" class="model-section">
      <div class="section-title">轮询器</div>
      <el-table :data="model.pollers || []" border stripe size="small" v-if="model.pollers?.length">
        <el-table-column prop="register" label="寄存器" width="150" />
        <el-table-column prop="address" label="地址" width="150" />
        <el-table-column prop="length" label="长度" width="100" />
      </el-table>
      <el-empty v-else description="暂无轮询器" :image-size="60" />
    </div>

    <div class="model-section">
      <div class="section-title">事件</div>
      <el-table :data="model.events || []" border stripe size="small" v-if="model.events?.length">
        <el-table-column prop="key" label="标识符" width="150" />
        <el-table-column prop="name" label="名称" width="150" />
        <el-table-column prop="type" label="事件类型" width="120" />
        <el-table-column label="参数">
          <template #default="{ row }">
            <el-tag v-for="p in row.params" :key="p.key" size="small" style="margin-right: 4px">
              {{ p.name }}({{ p.type }})
            </el-tag>
            <span v-if="!row.params?.length">-</span>
          </template>
        </el-table-column>
      </el-table>
      <el-empty v-else description="暂无事件" :image-size="60" />
    </div>

    <div class="model-section">
      <div class="section-title">服务</div>
      <el-table :data="model.services || []" border stripe size="small" v-if="model.services?.length">
        <el-table-column prop="key" label="标识符" width="120" />
        <el-table-column prop="name" label="名称" width="120" />
        <el-table-column prop="type" label="类型" width="100">
          <template #default="{ row }">
            {{ row.type === 'switch' ? '开关' : row.type === 'slider' ? '滑动条' : row.type === 'button' ? '按钮' : row.type === 'form' ? '表单' : '-' }}
          </template>
        </el-table-column>
        <el-table-column label="参数">
          <template #default="{ row }">
            <div v-if="row.params?.length">
              <div v-for="p in row.params" :key="p.key" class="param-info">
                <span>{{ p.name }}({{ p.type }})</span>
                <span v-if="p.min !== undefined && p.min !== null" class="param-meta">min:{{ p.min }}</span>
                <span v-if="p.max !== undefined && p.max !== null" class="param-meta">max:{{ p.max }}</span>
                <span v-if="p.step !== undefined && p.step !== null && p.step !== 0" class="param-meta">step:{{ p.step }}</span>
                <span v-if="p.default !== undefined && p.default !== null" class="param-meta">default:{{ p.default }}</span>
              </div>
            </div>
            <span v-else>-</span>
          </template>
        </el-table-column>
      </el-table>
      <el-empty v-else description="暂无服务" :image-size="60" />
    </div>

    <div class="model-section">
      <div class="section-title">配置</div>
      <el-table :data="model.settings || []" border stripe size="small" v-if="model.settings?.length">
        <el-table-column prop="key" label="标识符" width="120" />
        <el-table-column prop="name" label="名称" width="120" />
        <el-table-column label="配置参数">
          <template #default="{ row }">
            <div v-if="row.params?.length">
              <div v-for="p in row.params" :key="p.key" class="param-info">
                <span>{{ p.name }}({{ p.type }})</span>
                <span v-if="p.min !== undefined && p.min !== null" class="param-meta">min:{{ p.min }}</span>
                <span v-if="p.max !== undefined && p.max !== null" class="param-meta">max:{{ p.max }}</span>
                <span v-if="p.step !== undefined && p.step !== null && p.step !== 0" class="param-meta">step:{{ p.step }}</span>
                <span v-if="p.default !== undefined && p.default !== null" class="param-meta">default:{{ p.default }}</span>
              </div>
            </div>
            <span v-else>-</span>
          </template>
        </el-table-column>
      </el-table>
      <el-empty v-else description="暂无配置" :image-size="60" />
    </div>

    <div class="model-section">
      <div class="section-title">表达式计算</div>
      <el-table :data="model.expressions || []" border stripe size="small" v-if="model.expressions?.length">
        <el-table-column prop="key" label="标识符" width="150" />
        <el-table-column prop="name" label="名称" width="150" />
        <el-table-column prop="expr" label="表达式" min-width="300" />
      </el-table>
      <el-empty v-else description="暂无表达式" :image-size="60" />
    </div>

    <div class="model-section">
      <div class="section-title">过滤器</div>
      <el-table :data="model.filters || []" border stripe size="small" v-if="model.filters?.length">
        <el-table-column prop="name" label="名称" width="150" />
        <el-table-column prop="expr" label="表达式" min-width="300" />
      </el-table>
      <el-empty v-else description="暂无过滤器" :image-size="60" />
    </div>
  </div>
</template>

<script setup>
import { computed } from 'vue'

const props = defineProps({
  model: {
    type: Object,
    default: () => ({})
  },
  currentProtocol: {
    type: Object,
    default: () => null
  }
})

const hasProtocol = computed(() => !!props.currentProtocol)
</script>

<style lang="scss" scoped>
.model-preview {
  .model-section {
    margin-bottom: 24px;

    .section-title {
      font-size: 15px;
      font-weight: 600;
      margin-bottom: 12px;
      color: var(--text-primary);
    }

    .param-info {
      display: flex;
      flex-wrap: wrap;
      gap: 8px;
      margin-bottom: 4px;
      font-size: 12px;

      .param-meta {
        color: #909399;
        font-size: 11px;
      }
    }
  }
}
</style>
