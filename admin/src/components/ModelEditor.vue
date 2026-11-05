<template>
  <div class="model-editor">
    <div class="model-section">
      <div class="section-header">
        <h4>属性</h4>
        <el-button type="primary" link @click="addProperty">
          <el-icon><Plus /></el-icon>添加属性
        </el-button>
      </div>
      <el-table :data="data.properties" v-if="data.properties?.length" border size="small">
        <el-table-column type="expand" v-if="hasProtocol">
          <template #default="{ row }">
            <div class="expand-wrapper">
              <div class="expand-section">
                <div class="enums-header">
                  <span class="expand-title">枚举映射</span>
                  <el-button type="primary" link size="small" @click="addEnum(row)">
                    <el-icon><Plus /></el-icon>添加枚举
                  </el-button>
                </div>
                <el-table :data="row.enums || []" border size="small" v-if="row.enums?.length">
                  <el-table-column label="标识" width="120">
                    <template #default="{ row: enumRow }">
                      <el-input v-model="enumRow.key" placeholder="标识" size="small" />
                    </template>
                  </el-table-column>
                  <el-table-column label="名称" width="120">
                    <template #default="{ row: enumRow }">
                      <el-input v-model="enumRow.name" placeholder="名称" size="small" />
                    </template>
                  </el-table-column>
                  <el-table-column label="值" width="100">
                    <template #default="{ row: enumRow }">
                      <el-input-number controls-position="right" v-model="enumRow.value" size="small" style="width: 100%" />
                    </template>
                  </el-table-column>
                  <el-table-column label="操作" width="80">
                    <template #default="{ $index }">
                      <el-button link type="danger" size="small" @click="removeEnum(row, $index)">删除</el-button>
                    </template>
                  </el-table-column>
                </el-table>
                <el-empty v-else description="暂无枚举映射" :image-size="40" />
              </div>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="标识符" width="120">
          <template #default="{ row }">
            <el-input v-model="row.key" placeholder="标识" size="small" />
          </template>
        </el-table-column>
        <el-table-column label="名称" width="120">
          <template #default="{ row }">
            <el-input v-model="row.name" placeholder="名称" size="small" />
          </template>
        </el-table-column>
        <el-table-column label="类型" width="100">
          <template #default="{ row }">
            <el-select v-model="row.type" size="small" style="width: 100%">
              <el-option label="int" value="int" />
              <el-option label="float" value="float" />
              <el-option label="bool" value="bool" />
              <el-option label="string" value="string" />
            </el-select>
          </template>
        </el-table-column>
        <el-table-column label="读写模式" width="100">
          <template #default="{ row }">
            <el-select v-model="row.mode" size="small" style="width: 100%">
              <el-option v-for="opt in modeOptions" :key="opt.value" :label="opt.label" :value="opt.value" />
            </el-select>
          </template>
        </el-table-column>
        <el-table-column v-if="hasProtocol" label="寄存器" width="120">
          <template #default="{ row }">
            <el-select v-model="row.register" size="small" style="width: 100%" placeholder="选择寄存器" clearable>
              <el-option v-for="reg in (currentProtocol?.registers || [])" :key="reg.value" :label="reg.name" :value="reg.value" />
            </el-select>
          </template>
        </el-table-column>
        <el-table-column v-if="hasProtocol" label="格式" width="120">
          <template #default="{ row }">
            <el-select v-model="row.format" size="small" style="width: 100%" placeholder="选择格式" clearable>
              <el-option v-for="fmt in (currentProtocol?.formats || [])" :key="fmt.value" :label="fmt.name" :value="fmt.value" />
            </el-select>
          </template>
        </el-table-column>
        <el-table-column v-if="hasProtocol" label="倍率" width="90">
          <template #default="{ row }">
            <el-input-number controls-position="right" v-model="row.rate" :min="0" :step="0.1" size="small" style="width: 100%" />
          </template>
        </el-table-column>
        <el-table-column v-if="hasProtocol" label="地址" width="120">
          <template #default="{ row }">
            <el-input v-model="row.address" placeholder="如 H0001" size="small" />
          </template>
        </el-table-column>
        <el-table-column v-if="hasProtocol" label="小端" width="80">
          <template #default="{ row }">
            <el-switch v-model="row.little_endian" size="small" />
          </template>
        </el-table-column>
        <el-table-column v-if="hasProtocol" label="校准值" width="100">
          <template #default="{ row }">
            <el-input-number controls-position="right" v-model="row.correct" :step="0.1" size="small" style="width: 100%" />
          </template>
        </el-table-column>
        <el-table-column v-if="hasProtocol" label="取位" width="90">
          <template #default="{ row }">
            <el-input-number controls-position="right" v-model="row.bit" :min="0" :max="31" size="small" style="width: 100%" />
          </template>
        </el-table-column>
        <el-table-column label="最小" width="80">
          <template #default="{ row }">
            <el-input-number controls-position="right" v-model="row.min" size="small" style="width: 100%" />
          </template>
        </el-table-column>
        <el-table-column label="最大" width="80">
          <template #default="{ row }">
            <el-input-number controls-position="right" v-model="row.max" size="small" style="width: 100%" />
          </template>
        </el-table-column>
        <el-table-column label="阈值" width="90">
          <template #default="{ row }">
            <el-input-number controls-position="right" v-model="row.threshold" :min="0" size="small" style="width: 100%" />
          </template>
        </el-table-column>
        <el-table-column label="单位" width="80">
          <template #default="{ row }">
            <el-input v-model="row.unit" placeholder="单位" size="small" />
          </template>
        </el-table-column>
        <el-table-column label="精度" width="80">
          <template #default="{ row }">
            <el-input-number controls-position="right" v-model="row.precision" :min="0" size="small" style="width: 100%" />
          </template>
        </el-table-column>
        <el-table-column label="操作" width="80" fixed="right">
          <template #default="{ $index }">
            <el-button link type="danger" @click="removeProperty($index)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
      <el-empty v-else description="暂无属性" :image-size="20" />
    </div>

    <div v-if="hasProtocol" class="model-section">
      <div class="section-header">
        <h4>轮询器</h4>
        <el-button type="primary" link @click="addPoller">
          <el-icon><Plus /></el-icon>添加轮询器
        </el-button>
      </div>
      <el-table :data="data.pollers" v-if="data.pollers?.length" border size="small">
        <el-table-column label="寄存器" width="150">
          <template #default="{ row }">
            <el-select v-model="row.register" size="small" style="width: 100%" placeholder="选择寄存器" clearable>
              <el-option v-for="reg in (currentProtocol?.registers || [])" :key="reg.value" :label="reg.name" :value="reg.value" />
            </el-select>
          </template>
        </el-table-column>
        <el-table-column label="地址" width="150">
          <template #default="{ row }">
            <el-input v-model="row.address" placeholder="如 H0001" size="small" />
          </template>
        </el-table-column>
        <el-table-column label="长度" width="100">
          <template #default="{ row }">
            <el-input-number controls-position="right" v-model="row.length" :min="1" size="small" style="width: 100%" />
          </template>
        </el-table-column>
        <el-table-column label="操作" width="80" fixed="right">
          <template #default="{ $index }">
            <el-button link type="danger" @click="removePoller($index)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
      <el-empty v-else description="暂无轮询器" :image-size="20" />
    </div>

    <div class="model-section">
      <div class="section-header">
        <h4>事件</h4>
        <el-button type="primary" link @click="addEvent">
          <el-icon><Plus /></el-icon>添加事件
        </el-button>
      </div>
      <el-table :data="data.events" v-if="data.events?.length" border size="small">
        <el-table-column type="expand">
          <template #default="{ row }">
            <div class="sub-table-wrapper">
              <div class="sub-table-header">
                <span>参数列表</span>
                <el-button type="primary" link size="small" @click="addEventParam(row)">
                  <el-icon><Plus /></el-icon>添加参数
                </el-button>
              </div>
              <el-table :data="row.params" border size="small" v-if="row.params?.length">
                <el-table-column label="标识符" width="120">
                  <template #default="{ row: param }">
                    <el-input v-model="param.key" placeholder="标识" size="small" />
                  </template>
                </el-table-column>
                <el-table-column label="名称" width="120">
                  <template #default="{ row: param }">
                    <el-input v-model="param.name" placeholder="名称" size="small" />
                  </template>
                </el-table-column>
                <el-table-column label="类型" width="100">
                  <template #default="{ row: param }">
                    <el-select v-model="param.type" size="small" style="width: 100%">
                      <el-option label="int" value="int" />
                      <el-option label="float" value="float" />
                      <el-option label="bool" value="bool" />
                      <el-option label="string" value="string" />
                    </el-select>
                  </template>
                </el-table-column>
                <el-table-column label="操作" width="80">
                  <template #default="{ $index }">
                    <el-button link type="danger" size="small" @click="removeEventParam(row, $index)">删除</el-button>
                  </template>
                </el-table-column>
              </el-table>
              <el-empty v-else description="暂无参数" :image-size="40" />
            </div>
          </template>
        </el-table-column>
        <el-table-column label="标识符" width="200">
          <template #default="{ row }">
            <el-input v-model="row.key" placeholder="标识" size="small" />
          </template>
        </el-table-column>
        <el-table-column label="名称" width="200">
          <template #default="{ row }">
            <el-input v-model="row.name" placeholder="名称" size="small" />
          </template>
        </el-table-column>
        <el-table-column label="参数数" width="200">
          <template #default="{ row }">
            {{ row.params?.length || 0 }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="80" fixed="right">
          <template #default="{ $index }">
            <el-button link type="danger" @click="removeEvent($index)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
      <el-empty v-else description="暂无事件" :image-size="20" />
    </div>

    <div class="model-section">
      <div class="section-header">
        <h4>服务</h4>
        <el-button type="primary" link @click="addService">
          <el-icon><Plus /></el-icon>添加服务
        </el-button>
      </div>
      <el-table :data="data.services" v-if="data.services?.length" border size="small">
        <el-table-column type="expand">
          <template #default="{ row }">
            <div class="sub-table-wrapper">
              <div class="sub-table-header">
                <span>参数列表</span>
                <el-button type="primary" link size="small" @click="addServiceParam(row)">
                  <el-icon><Plus /></el-icon>添加参数
                </el-button>
              </div>
              <el-table :data="row.params" border size="small" v-if="row.params?.length">
                <el-table-column label="标识符" width="100">
                  <template #default="{ row: param }">
                    <el-input v-model="param.key" placeholder="标识" size="small" />
                  </template>
                </el-table-column>
                <el-table-column label="名称" width="100">
                  <template #default="{ row: param }">
                    <el-input v-model="param.name" placeholder="名称" size="small" />
                  </template>
                </el-table-column>
                <el-table-column label="类型" width="80">
                  <template #default="{ row: param }">
                    <el-select v-model="param.type" size="small" style="width: 100%">
                      <el-option label="int" value="int" />
                      <el-option label="float" value="float" />
                      <el-option label="bool" value="bool" />
                      <el-option label="string" value="string" />
                    </el-select>
                  </template>
                </el-table-column>
                <el-table-column label="最小值" width="90">
                  <template #default="{ row: param }">
                    <el-input-number controls-position="right" v-model="param.min" size="small" style="width: 100%" placeholder="min" />
                  </template>
                </el-table-column>
                <el-table-column label="最大值" width="90">
                  <template #default="{ row: param }">
                    <el-input-number controls-position="right" v-model="param.max" size="small" style="width: 100%" placeholder="max" />
                  </template>
                </el-table-column>
                <el-table-column label="步长" width="90">
                  <template #default="{ row: param }">
                    <el-input-number controls-position="right" v-model="param.step" :min="0" size="small" style="width: 100%" placeholder="step" />
                  </template>
                </el-table-column>
                <el-table-column label="默认值" width="100">
                  <template #default="{ row: param }">
                    <el-input v-model="param.default" placeholder="默认值" size="small" />
                  </template>
                </el-table-column>
                <el-table-column label="寄存器绑定" width="100">
                  <template #default="{ row: param }">
                    <el-switch v-model="param.register"/>
                  </template>
                </el-table-column>
                <el-table-column label="操作" width="80">
                  <template #default="{ $index }">
                    <el-button link type="danger" size="small" @click="removeServiceParam(row, $index)">删除</el-button>
                  </template>
                </el-table-column>
              </el-table>
              <el-empty v-else description="暂无参数" :image-size="40" />
            </div>
          </template>
        </el-table-column>
        <el-table-column label="标识符" width="200">
          <template #default="{ row }">
            <el-input v-model="row.key" placeholder="标识" size="small" />
          </template>
        </el-table-column>
        <el-table-column label="名称" width="200">
          <template #default="{ row }">
            <el-input v-model="row.name" placeholder="名称" size="small" />
          </template>
        </el-table-column>
        <el-table-column label="类型" width="200">
          <template #default="{ row }">
            <el-select v-model="row.type" size="small" style="width: 100%">
              <el-option label="开关" value="switch" />
              <el-option label="滑动条" value="slider" />
              <el-option label="按钮" value="button" />
              <el-option label="表单" value="form" />
            </el-select>
          </template>
        </el-table-column>
        <el-table-column label="参数数" width="120">
          <template #default="{ row }">
            {{ row.params?.length || 0 }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="80" fixed="right">
          <template #default="{ $index }">
            <el-button link type="danger" @click="removeService($index)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
      <el-empty v-else description="暂无服务" :image-size="20" />
    </div>

    <div class="model-section">
      <div class="section-header">
        <h4>配置</h4>
        <el-button type="primary" link @click="addSetting">
          <el-icon><Plus /></el-icon>添加配置
        </el-button>
      </div>
      <el-table :data="data.settings" v-if="data.settings?.length" border size="small">
        <el-table-column type="expand">
          <template #default="{ row }">
            <div class="sub-table-wrapper">
              <div class="sub-table-header">
                <span>参数列表</span>
                <el-button type="primary" link size="small" @click="addSettingParam(row)">
                  <el-icon><Plus /></el-icon>添加参数
                </el-button>
              </div>
              <el-table :data="row.params" border size="small" v-if="row.params?.length">
                <el-table-column label="标识符" width="100">
                  <template #default="{ row: param }">
                    <el-input v-model="param.key" placeholder="标识" size="small" />
                  </template>
                </el-table-column>
                <el-table-column label="名称" width="100">
                  <template #default="{ row: param }">
                    <el-input v-model="param.name" placeholder="名称" size="small" />
                  </template>
                </el-table-column>
                <el-table-column label="类型" width="80">
                  <template #default="{ row: param }">
                    <el-select v-model="param.type" size="small" style="width: 100%">
                      <el-option label="int" value="int" />
                      <el-option label="float" value="float" />
                      <el-option label="bool" value="bool" />
                      <el-option label="string" value="string" />
                    </el-select>
                  </template>
                </el-table-column>
                <el-table-column label="最小值" width="90">
                  <template #default="{ row: param }">
                    <el-input-number controls-position="right" v-model="param.min" size="small" style="width: 100%" placeholder="min" />
                  </template>
                </el-table-column>
                <el-table-column label="最大值" width="90">
                  <template #default="{ row: param }">
                    <el-input-number controls-position="right" v-model="param.max" size="small" style="width: 100%" placeholder="max" />
                  </template>
                </el-table-column>
                <el-table-column label="步长" width="90">
                  <template #default="{ row: param }">
                    <el-input-number controls-position="right" v-model="param.step" :min="0" size="small" style="width: 100%" placeholder="step" />
                  </template>
                </el-table-column>
                <el-table-column label="默认值" width="100">
                  <template #default="{ row: param }">
                    <el-input v-model="param.default" placeholder="默认值" size="small" />
                  </template>
                </el-table-column>
                <el-table-column label="寄存器绑定" width="100">
                  <template #default="{ row: param }">
                    <el-switch v-model="param.register"/>
                  </template>
                </el-table-column>
                <el-table-column label="操作" width="80">
                  <template #default="{ $index }">
                    <el-button link type="danger" size="small" @click="removeSettingParam(row, $index)">删除</el-button>
                  </template>
                </el-table-column>
              </el-table>
              <el-empty v-else description="暂无参数" :image-size="40" />
            </div>
          </template>
        </el-table-column>
        <el-table-column label="标识符" width="200">
          <template #default="{ row }">
            <el-input v-model="row.key" placeholder="标识" size="small" />
          </template>
        </el-table-column>
        <el-table-column label="名称" width="200">
          <template #default="{ row }">
            <el-input v-model="row.name" placeholder="名称" size="small" />
          </template>
        </el-table-column>
        <el-table-column label="参数数" width="120">
          <template #default="{ row }">
            {{ row.params?.length || 0 }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="80" fixed="right">
          <template #default="{ $index }">
            <el-button link type="danger" @click="removeSetting($index)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
      <el-empty v-else description="暂无配置" :image-size="20" />
    </div>

    <div class="model-section">
      <div class="section-header">
        <h4>表达式计算</h4>
        <el-button type="primary" link @click="addExpression">
          <el-icon><Plus /></el-icon>添加表达式
        </el-button>
      </div>
      <el-alert type="info" :closable="false" show-icon style="margin-bottom: 12px">
        属性上报（MQTT/broker 上行）或轮询完成后自动计算，支持引用属性值。语法同 <a href="https://expr-lang.org/docs/language-definition" target="_blank">expr-lang</a>。
        可用内置函数: <code>ABS, MIN, MAX, ROUND, SQRT, POW, FLOOR, CEIL, SIN, COS, BAND, BOR, BXOR, BSL, BSR</code>。
        示例: <code>功率 = 电流 * 电压</code>, <code>上限 = ROUND((温度 &amp; 0x00FF) / 10)</code>
      </el-alert>
      <el-table :data="data.expressions" v-if="data.expressions?.length" border size="small">
        <el-table-column label="标识符" width="150">
          <template #default="{ row }">
            <el-input v-model="row.key" placeholder="如 功率" size="small" />
          </template>
        </el-table-column>
        <el-table-column label="名称" width="150">
          <template #default="{ row }">
            <el-input v-model="row.name" placeholder="显示名" size="small" />
          </template>
        </el-table-column>
        <el-table-column label="表达式" min-width="300">
          <template #default="{ row }">
            <el-input v-model="row.expr" placeholder="如 电流 * 电压" size="small" />
          </template>
        </el-table-column>
        <el-table-column label="操作" width="80" fixed="right">
          <template #default="{ $index }">
            <el-button link type="danger" @click="removeExpression($index)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
      <el-empty v-else description="暂无表达式" :image-size="20" />
    </div>

    <div class="model-section">
      <div class="section-header">
        <h4>过滤器</h4>
        <el-button type="primary" link @click="addFilter">
          <el-icon><Plus /></el-icon>添加过滤器
        </el-button>
      </div>
      <el-alert type="info" :closable="false" show-icon style="margin-bottom: 12px">
        对每条上报数据做拦截判断（MQTT/broker 上行与轮询均生效）。任一过滤器结果为 false 则整条丢弃。
        可用 ABS / MIN / MAX / ROUND / SQRT / BAND / BOR 等函数，引用属性名（如 温度 < 50）
      </el-alert>
      <el-table :data="data.filters" v-if="data.filters?.length" border size="small">
        <el-table-column label="名称" width="150">
          <template #default="{ row }">
            <el-input v-model="row.name" placeholder="过滤器名" size="small" />
          </template>
        </el-table-column>
        <el-table-column label="表达式" min-width="300">
          <template #default="{ row }">
            <el-input v-model="row.expr" placeholder="如 温度 > 0 && 温度 < 50" size="small" />
          </template>
        </el-table-column>
        <el-table-column label="操作" width="80" fixed="right">
          <template #default="{ $index }">
            <el-button link type="danger" @click="removeFilter($index)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
      <el-empty v-else description="暂无过滤器" :image-size="20" />
    </div>
  </div>
</template>

<script setup>
import { computed } from 'vue'
import { Plus } from '@element-plus/icons-vue'

const props = defineProps({
  modelValue: {
    type: Object,
    default: () => ({})
  },
  protocols: {
    type: Array,
    default: () => []
  },
  currentProtocol: {
    type: Object,
    default: () => null
  }
})

const emit = defineEmits(['update:modelValue'])

const data = computed(() => props.modelValue)

const hasProtocol = computed(() => !!props.currentProtocol)

const modeOptions = [
  { label: '只读', value: 'r' },
  { label: '只写', value: 'w' },
  { label: '读写', value: 'rw' }
]

const addProperty = () => {
  data.value.properties = data.value.properties || []
  data.value.properties.push({
    key: '',
    name: '',
    type: 'int',
    unit: '',
    precision: 0,
    mode: 'rw',
    register: '',
    format: '',
    rate: 1,
    address: '',
    threshold: 0,
    min: 0,
    max: 0,
    little_endian: false,
    correct: 0,
    bit: 0,
    enums: []
  })
  emit('update:modelValue', data.value)
}

const removeProperty = (index) => {
  data.value.properties.splice(index, 1)
  emit('update:modelValue', data.value)
}

const addEnum = (row) => {
  row.enums = row.enums || []
  row.enums.push({
    key: '',
    name: '',
    value: 0
  })
  emit('update:modelValue', data.value)
}

const removeEnum = (row, index) => {
  row.enums.splice(index, 1)
  emit('update:modelValue', data.value)
}

const addPoller = () => {
  data.value.pollers = data.value.pollers || []
  data.value.pollers.push({
    register: '',
    address: '',
    length: 1
  })
  emit('update:modelValue', data.value)
}

const removePoller = (index) => {
  data.value.pollers.splice(index, 1)
  emit('update:modelValue', data.value)
}

const addEvent = () => {
  data.value.events = data.value.events || []
  data.value.events.push({
    key: '',
    name: '',
    params: []
  })
  emit('update:modelValue', data.value)
}

const removeEvent = (index) => {
  data.value.events.splice(index, 1)
  emit('update:modelValue', data.value)
}

const addEventParam = (event) => {
  event.params = event.params || []
  event.params.push({
    key: '',
    name: '',
    type: 'int'
  })
  emit('update:modelValue', data.value)
}

const removeEventParam = (event, index) => {
  event.params.splice(index, 1)
  emit('update:modelValue', data.value)
}

const addService = () => {
  data.value.services = data.value.services || []
  data.value.services.push({
    key: '',
    name: '',
    type: 'switch',
    params: []
  })
  emit('update:modelValue', data.value)
}

const removeService = (index) => {
  data.value.services.splice(index, 1)
  emit('update:modelValue', data.value)
}

const addServiceParam = (service) => {
  service.params = service.params || []
  service.params.push({
    key: '',
    name: '',
    type: 'int',
    min: null,
    max: null,
    step: null,
    default: '',
    register: false
  })
  emit('update:modelValue', data.value)
}

const removeServiceParam = (service, index) => {
  service.params.splice(index, 1)
  emit('update:modelValue', data.value)
}

const addSetting = () => {
  data.value.settings = data.value.settings || []
  data.value.settings.push({
    key: '',
    name: '',
    params: []
  })
  emit('update:modelValue', data.value)
}

const removeSetting = (index) => {
  data.value.settings.splice(index, 1)
  emit('update:modelValue', data.value)
}

const addSettingParam = (setting) => {
  setting.params = setting.params || []
  setting.params.push({
    key: '',
    name: '',
    type: 'int',
    min: null,
    max: null,
    step: null,
    default: '',
    register: false
  })
  emit('update:modelValue', data.value)
}

const removeSettingParam = (setting, index) => {
  setting.params.splice(index, 1)
  emit('update:modelValue', data.value)
}

const addExpression = () => {
  data.value.expressions = data.value.expressions || []
  data.value.expressions.push({
    key: '',
    name: '',
    expr: ''
  })
  emit('update:modelValue', data.value)
}

const removeExpression = (index) => {
  data.value.expressions.splice(index, 1)
  emit('update:modelValue', data.value)
}

const addFilter = () => {
  data.value.filters = data.value.filters || []
  data.value.filters.push({
    name: '',
    expr: ''
  })
  emit('update:modelValue', data.value)
}

const removeFilter = (index) => {
  data.value.filters.splice(index, 1)
  emit('update:modelValue', data.value)
}
</script>

<style lang="scss" scoped>
.model-editor {
  .model-section {
    margin-bottom: 24px;

    .section-header {
      display: flex;
      justify-content: space-between;
      align-items: center;
      margin-bottom: 12px;

      h4 {
        margin: 0;
        color: #303133;
        font-weight: 600;
      }
    }

    .sub-table-wrapper {
      padding: 12px 16px;
      background: #fafafa;
      overflow-x: auto;

      .sub-table-header {
        display: flex;
        justify-content: space-between;
        align-items: center;
        margin-bottom: 8px;

        span {
          font-size: 13px;
          font-weight: 600;
          color: #606266;
        }
      }

      .el-table {
        min-width: 730px;
      }
    }

    .expand-wrapper {
      padding: 12px 16px;
      background: #f8f9fa;

      .expand-section {
        .expand-title {
          font-size: 13px;
          font-weight: 600;
          color: #606266;
          margin-bottom: 8px;
          display: block;
        }

        .expand-row {
          display: flex;
          gap: 24px;
          margin-bottom: 16px;

          .expand-item {
            display: flex;
            align-items: center;
            gap: 8px;

            .expand-label {
              font-size: 13px;
              color: #606266;
            }
          }
        }

        .expand-enums {
          .enums-header {
            display: flex;
            justify-content: space-between;
            align-items: center;
            margin-bottom: 8px;
          }
        }
      }
    }
  }
}
</style>
