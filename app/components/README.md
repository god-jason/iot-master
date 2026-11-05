# App UI 组件库

## 简介

`app/components` 是当前项目的 uni-app UI 组件目录，组件统一采用 `Ds` 前缀，主要面向移动端页面搭建。

- 技术栈：`uni-app` + `Vue 3` + `script setup`
- 样式体系：统一使用 `components/variables.scss`
- 组件形态：以展示组件和轻交互组件为主，少量组件带表单或容器能力
- 命名约定：组件文件名与标签名保持一致，例如 `DsButton.vue` 对应 `<DsButton />`

## 使用方式

当前目录没有稳定的统一导出入口，页面层推荐直接导入具体组件文件：

```vue
<template>
  <DsButton text="保存" @click="handleSave" />
</template>

<script setup>
import DsButton from '@/components/DsButton.vue'

const handleSave = () => {
  uni.showToast({ title: '已保存', icon: 'success' })
}
</script>
```

如果项目已配置 easycom，也可以直接在模板中使用组件标签，但仍建议文档和示例以“显式导入”为准，便于排查小程序编译问题。

## 组件总览

### 基础视觉

| 组件 | 用途 | 常用 props | 事件 | 插槽 / 备注 |
|------|------|------------|------|-------------|
| `DsIcon` | 图标字体封装 | `name`、`size`、`color` | `click` | 依赖本地 `ionicons` 字体 |
| `DsAvatar` | 头像展示 | `src`、`icon`、`size`、`shape`、`badge` | `click` | 无图时回退图标 |
| `DsButton` | 按钮 | `text`、`type`、`size`、`loading`、`disabled`、`block` | `click` | 支持默认 slot |
| `DsBadge` | 角标包裹器 | `value`、`max`、`dot`、`type`、`position` | 无 | 默认 slot 为被包裹内容 |
| `DsTag` | 状态标签 | `text`、`type`、`size`、`icon`、`closable` | `click`、`close` | 关闭按钮不冒泡 |
| `DsImage` | 图片展示 | `src`、`mode`、`lazyLoad` | `load`、`error`、`click` | 点击会触发图片预览 |

### 表单输入

| 组件 | 用途 | 常用 props | 事件 | 插槽 / 备注 |
|------|------|------------|------|-------------|
| `DsInput` | 通用输入框 | `modelValue`、`type`、`label`、`placeholder`、`prefixIcon`、`clearable` | `update:modelValue`、`focus`、`blur`、`confirm` | `type='textarea'` 目前不是真正多行输入 |
| `DsInputBar` | 工具栏输入条 | `modelValue`、`leftIcon`、`rightIcon`、`rightText`、`placeholder` | `update:modelValue`、`focus`、`blur`、`confirm`、`right-click` | 适合页脚输入和搜索 |
| `DsNumberInput` | 数字步进器 | `modelValue`、`min`、`max`、`step`、`precision`、`unit` | `update:modelValue`、`change` | 自动做数值钳制 |
| `DsSwitch` | 开关 | `value`、`label`、`activeColor`、`inactiveColor` | `change` | 不是标准 `v-model` 组件 |
| `DsSlider` | 滑块 | `modelValue`、`min`、`max`、`step`、`label`、`unit` | `update:modelValue`、`change` | 松手后才对外确认值 |
| `DsSearchBar` | 搜索栏 | `modelValue`、`placeholder`、`actionText`、`actionIcon`、`clearable` | `update:modelValue`、`confirm`、`action`、`prefix-click`、`suffix-click` | 搜索页优先使用 |
| `DsComboBox` | 组合下拉选择 | `modelValue`、`options`、`placeholder` | `update:modelValue`、`change` | 使用了 `document` 事件，跨端需验证 |
| `DsDropdown` | 轻量下拉菜单 | `modelValue`、`options`、`placeholder`、`direction` | `update:modelValue`、`change` | `direction='up'` 当前实现需谨慎使用 |
| `DsRadio` | 单选组 | `modelValue`、`options` | `update:modelValue`、`change` | 选项建议 `{ label, value }` |
| `DsCheckbox` | 多选组 | `modelValue`、`options` | `update:modelValue`、`change` | 返回数组值 |
| `DsForm` | 表单容器 | `model`、`rules` | `validate`、`submit` | 暴露 `validate` / `submit` |
| `DsFormItem` | 表单项外壳 | `label`、`required`、`error`、`help` | 无 | 与 `DsForm` 配套使用 |

### 布局导航

| 组件 | 用途 | 常用 props | 事件 | 插槽 / 备注 |
|------|------|------------|------|-------------|
| `DsNavBar` | 顶部导航栏 | `title`、`showBack`、`transparent`、`statusBar`、`color` | `click` | `left`、`right` |
| `DsTabs` | 标签页容器 | `modelValue`、`type`、`scrollable` | `update:modelValue`、`change` | 需与 `DsTabItem` 手动联动 |
| `DsTabItem` | 单个标签项 | `label`、`badge`、`index`、`active` | `click` | 激活态依赖外部传入 |
| `DsCard` | 卡片容器 | `title`、`size`、`shadow`、`bordered`、`padding` | `click` | `header-extra`、`footer` |
| `DsSection` | 分区容器 | `title`、`type`、`showLine` | 无 | `extra` |
| `DsList` | 列表容器 | `bordered`、`striped` | 无 | 默认 slot |
| `DsListItem` | 列表项 | `title`、`desc`、`extra`、`icon`、`avatar`、`arrow` | `click` | `content`、`extra` |
| `DsGrid` | 宫格入口 | `items`、`columns`、`gap`、`iconSize` | `click` | 支持 `item-${index}` 动态 slot |
| `DsCollapse` | 折叠面板容器 | `modelValue`、`accordion` | `update:modelValue`、`change` | 默认 slot |
| `DsCollapseItem` | 折叠面板项 | `title`、`index` | 无 | 默认 slot |
| `DsDialog` | 对话框 | `visible`、`title`、`message`、`showCancel` | `update:visible`、`confirm`、`cancel`、`close` | 默认 slot |
| `DsDrawer` | 抽屉层 | `visible`、`title`、`direction`、`width`、`height` | `update:visible`、`close` | `footer` |

### 状态反馈

| 组件 | 用途 | 常用 props | 事件 | 插槽 / 备注 |
|------|------|------------|------|-------------|
| `DsEmpty` | 空状态 | `text`、`subText`、`type`、`icon` | `click` | `action` |
| `DsLoading` | 加载态 | `text`、`type` | 无 | 纯展示组件 |
| `DsResult` | 结果页 | `type`、`title`、`message`、`actions`、`showHome` | `click` | `actions` 通过 props 传函数 |
| `DsNoticeBar` | 通知栏 | `text`、`type`、`icon`、`scrollable`、`closable` | `click`、`close` | `scrollable` 是横向滚动，不是跑马灯 |
| `DsProgress` | 进度条 | `value`、`max`、`label`、`type`、`showText`、`color` | 无 | 纯展示组件 |
| `DsFab` | 悬浮按钮 | `icon`、`text`、`position`、`mini`、`show` | `click` | 当前文本默认不显示 |
| `DsBacktop` | 返回顶部按钮 | `threshold`、`duration` | 无 | 更偏 H5 场景 |
| `DsNoMore` | 列表无更多提示 | `text` | 无 | 纯展示组件 |

### 数据展示

| 组件 | 用途 | 常用 props | 事件 | 插槽 / 备注 |
|------|------|------------|------|-------------|
| `DsDescription` | 描述列表容器 | `title` | 无 | 默认 slot |
| `DsDescriptionItem` | 描述项 | `label`、`value` | 无 | 默认 slot |
| `DsStatistic` | 统计值展示 | `value`、`label`、`icon`、`prefix`、`suffix` | 无 | 非数字值需谨慎 |
| `DsSwiper` | 轮播容器 | `items`、`current`、`autoplay`、`interval`、`pagination` | `update:current`、`change`、`click` | 可数据驱动，也可自定义 slot |
| `DsSwiperItem` | 轮播子项 | 无 | `click` | 默认 slot |
| `DsQrCode` | 二维码 | `value`、`size`、`text`、`level` | `click` | 依赖 canvas 能力 |
| `DsTable` | 轻量表格 | `columns`、`data`、`emptyText` | 无 | 支持 `cell-${key}` 动态 slot |

## 常用示例

### 1. 按钮与卡片

```vue
<template>
  <DsCard title="设备信息" @click="openDetail">
    <text>点击卡片进入详情</text>
    <template #footer>
      <DsButton text="立即查看" size="small" @click="openDetail" />
    </template>
  </DsCard>
</template>

<script setup>
import DsButton from '@/components/DsButton.vue'
import DsCard from '@/components/DsCard.vue'

const openDetail = () => {
  uni.showToast({ title: '打开详情', icon: 'none' })
}
</script>
```

### 2. 输入与搜索

```vue
<template>
  <DsSearchBar
    v-model="keyword"
    placeholder="搜索设备"
    action-text="搜索"
    @confirm="handleSearch"
    @action="handleSearch"
  />
</template>

<script setup>
import { ref } from 'vue'
import DsSearchBar from '@/components/DsSearchBar.vue'

const keyword = ref('')

const handleSearch = () => {
  uni.showToast({ title: keyword.value || '请输入关键词', icon: 'none' })
}
</script>
```

### 3. `DsSwitch` 的正确写法

`DsSwitch` 不是 `modelValue + update:modelValue` 模式，推荐这样使用：

```vue
<template>
  <DsSwitch :value="enabled" label="设备开关" @change="enabled = $event" />
</template>

<script setup>
import { ref } from 'vue'
import DsSwitch from '@/components/DsSwitch.vue'

const enabled = ref(false)
</script>
```

### 4. `DsTabs` 与 `DsTabItem`

当前标签页需要父层自己维护激活态：

```vue
<template>
  <DsTabs v-model="activeTab" type="line">
    <DsTabItem label="全部" :index="0" :active="activeTab === 0" @click="activeTab = 0" />
    <DsTabItem label="在线" :index="1" :active="activeTab === 1" @click="activeTab = 1" />
    <DsTabItem label="离线" :index="2" :active="activeTab === 2" @click="activeTab = 2" />
  </DsTabs>
</template>

<script setup>
import { ref } from 'vue'
import DsTabs from '@/components/DsTabs.vue'
import DsTabItem from '@/components/DsTabItem.vue'

const activeTab = ref(0)
</script>
```

### 5. 表单校验

```vue
<template>
  <DsForm ref="formRef" :model="formData" :rules="rules">
    <DsFormItem label="用户名" required>
      <DsInput v-model="formData.username" placeholder="请输入用户名" />
    </DsFormItem>
    <DsButton text="提交" block @click="handleSubmit" />
  </DsForm>
</template>

<script setup>
import { ref } from 'vue'
import DsButton from '@/components/DsButton.vue'
import DsForm from '@/components/DsForm.vue'
import DsFormItem from '@/components/DsFormItem.vue'
import DsInput from '@/components/DsInput.vue'

const formRef = ref(null)
const formData = ref({ username: '' })

const rules = {
  username: { required: true, message: '请输入用户名' }
}

const handleSubmit = async () => {
  const result = await formRef.value?.validate()
  if (result?.valid) {
    uni.showToast({ title: '校验通过', icon: 'success' })
  }
}
</script>
```

## 详细说明

### 基础视觉

#### `DsIcon`

- 用途：统一封装项目图标字体，屏蔽图标类名前缀差异。
- 适用场景：列表入口、状态提示、按钮前缀、导航栏操作图标。
- 主要 props：`name`、`size=24`、`color='#333'`、`customClass`、`customStyle`。
- 事件：`click`。
- 插槽：无。
- 示例：`<DsIcon name="home" :size="28" color="#5BA3E8" @click="goHome" />`
- 注意事项：依赖 `app/ionicons/ionicons.css` 引入的本地 `ionicons` 字体；建议统一传 `home` 或完整的 `ionicons-home`，不要混入其他图标库前缀。

#### `DsAvatar`

- 用途：展示用户、设备、组织等头像，也可承载角标信息。
- 适用场景：个人中心、通讯录、设备列表、业务卡片头部。
- 主要 props：`src=''`、`icon='member'`、`size='medium'`、`shape='circle'`、`iconColor='#fff'`、`bgColor='#5BA3E8'`、`badge=null`。
- 事件：`click`。
- 插槽：无。
- 示例：`<DsAvatar src="/static/user.png" size="large" :badge="{ type: 'danger', value: 3 }" />`
- 注意事项：`src` 为空时会回退到图标头像；`badge` 更适合传 `{ type, value }` 结构。

#### `DsBadge`

- 用途：为图标、头像、按钮等内容增加数量角标或小红点。
- 适用场景：消息提醒、未读数、待处理数量、状态提示。
- 主要 props：`value=0`、`max=99`、`dot=false`、`visible=true`、`type='danger'`、`position='top-right'`。
- 事件：无。
- 插槽：默认 slot。
- 示例：`<DsBadge :value="12"><DsIcon name="message" /></DsBadge>`
- 注意事项：`value=0` 且 `dot=false` 时默认不显示；角标定位依赖包裹内容尺寸。

#### `DsTag`

- 用途：展示状态、分类、级别等轻量标签。
- 适用场景：设备状态、工单标签、筛选项标识、可关闭标签。
- 主要 props：`text`、`type='default'`、`size='medium'`、`icon`、`closable=false`。
- 事件：`click`、`close`。
- 插槽：无。
- 示例：`<DsTag text="在线" type="success" icon="success" />`
- 注意事项：关闭按钮只触发 `close`，不会继续冒泡触发根节点 `click`。

#### `DsButton`

- 用途：统一按钮样式与交互，支持图标、加载、禁用、块级布局。
- 适用场景：表单提交、弹窗确认、操作入口、页脚主操作。
- 主要 props：`text`、`type='primary'`、`size='medium'`、`icon`、`loading`、`disabled`、`round`、`block`、`bgColor`、`textColor`。
- 事件：`click`。
- 插槽：默认 slot。
- 示例：`<DsButton text="保存" type="primary" block @click="handleSave" />`
- 注意事项：`loading` 或 `disabled` 时内部会阻止点击；如果同时传 `text` 和默认 slot，需要自行避免内容重复。

#### `DsImage`

- 用途：统一图片展示、加载态与点击预览行为。
- 适用场景：设备照片、商品图、相册缩略图、详情图集。
- 主要 props：`src`、`mode='aspectFill'`、`lazyLoad=false`。
- 事件：`load`、`error`、`click`。
- 插槽：无。
- 示例：`<DsImage src="/static/device.jpg" mode="widthFix" @error="useFallback" />`
- 注意事项：点击根节点会触发 `uni.previewImage()`；如果页面不希望预览，建议外层再做一层封装。

### 表单输入

#### `DsInput`

- 用途：通用文本输入组件，集成标签、前后缀图标、清空能力。
- 适用场景：登录表单、设备名称编辑、搜索输入、备注录入。
- 主要 props：`modelValue`、`type='text'|'password'|'number'|'digit'|'idcard'|'textarea'`、`label`、`placeholder`、`prefixIcon`、`suffixIcon`、`clearable`、`maxlength`、`disabled`、`readonly`、`autoFocus`、`confirmType`、`error`、`errorMessage`、`helpText`。
- 事件：`update:modelValue`、`focus`、`blur`、`confirm`。
- 插槽：无。
- 示例：`<DsInput v-model="form.name" label="设备名称" placeholder="请输入设备名称" clearable />`
- 注意事项：`type='textarea'` 目前仍然基于单行输入实现，不适合真正多行文本。

#### `DsSearchBar`

- 用途：面向搜索页的专用输入条，内置搜索交互与操作区。
- 适用场景：设备检索、列表筛选、订单搜索、顶部搜索栏。
- 主要 props：`modelValue`、`placeholder='请输入搜索内容'`、`clearable=true`、`actionText`、`actionIcon`、`confirmType='search'`、`prefixIcon`、`suffixIcon`。
- 事件：`update:modelValue`、`confirm`、`action`、`focus`、`blur`、`prefix-click`、`suffix-click`。
- 插槽：无。
- 示例：`<DsSearchBar v-model="keyword" action-text="搜索" @confirm="search" @action="search" />`
- 注意事项：搜索页优先使用该组件而不是 `DsInput`；右侧动作建议只保留“文字”或“图标”中的一种主表达。

#### `DsInputBar`

- 用途：适合底部或工具栏区域的紧凑输入条。
- 适用场景：聊天输入、评论输入、页面底部快捷搜索、指令发送。
- 主要 props：`modelValue`、`leftIcon`、`leftIconColor`、`rightIcon`、`rightIconColor`、`rightText`、`placeholder`、`maxlength`、`disabled`、`confirmType`。
- 事件：`update:modelValue`、`focus`、`blur`、`confirm`、`right-click`。
- 插槽：无。
- 示例：`<DsInputBar v-model="message" left-icon="edit" right-text="发送" @right-click="sendMessage" />`
- 注意事项：右侧区域优先显示 `rightIcon`，没传图标时才显示 `rightText`。

#### `DsNumberInput`

- 用途：数字步进器，支持上下限、步长与精度控制。
- 适用场景：数量设置、阈值配置、设备参数录入、计费数值输入。
- 主要 props：`modelValue=0`、`min=0`、`max=99999`、`step=1`、`precision=0`、`unit=''`、`disabled=false`。
- 事件：`update:modelValue`、`change`。
- 插槽：无。
- 示例：`<DsNumberInput v-model="threshold" :min="0" :max="100" :step="5" unit="%" />`
- 注意事项：加减和失焦都会触发数值修正；如果依赖原始输入态，需要在外层单独缓存。

#### `DsSlider`

- 用途：滑动选择数值范围内的单个值。
- 适用场景：音量、亮度、阈值、百分比调节。
- 主要 props：`modelValue=50`、`min=0`、`max=100`、`step=1`、`label`、`unit`、`color='#5BA3E8'`、`showLabel=false`、`showRange=false`。
- 事件：`update:modelValue`、`change`。
- 插槽：无。
- 示例：`<DsSlider v-model="volume" label="音量" unit="%" />`
- 注意事项：拖动过程中主要更新内部状态，松手后才向外确认值；如果要实时联动，需要重点验证交互是否满足预期。

#### `DsSwitch`

- 用途：布尔开关切换，简化是/否状态表达。
- 适用场景：启停控制、通知开关、功能启用、权限切换。
- 主要 props：`value=false`、`label=''`、`activeColor='#5BA3E8'`、`inactiveColor='#e0e0e0'`。
- 事件：`change(nextValue)`。
- 插槽：无。
- 示例：`<DsSwitch :value="enabled" label="设备开关" @change="enabled = $event" />`
- 注意事项：不是标准 `v-model` 组件；颜色 props 当前没有完全驱动最终样式。

#### `DsRadio`

- 用途：单选组，适合有限选项中二选一或多选一。
- 适用场景：状态筛选、模式切换、单项配置。
- 主要 props：`modelValue`、`options=[]`。
- 事件：`update:modelValue`、`change(option)`。
- 插槽：无。
- 示例：`<DsRadio v-model="status" :options="[{ label: '全部', value: '' }, { label: '在线', value: 'online' }]" />`
- 注意事项：推荐选项结构为 `{ label, value, icon? }`；更复杂渲染建议改用自定义组件。

#### `DsCheckbox`

- 用途：多选组，返回数组值。
- 适用场景：标签过滤、权限勾选、批量条件选择、偏好设置。
- 主要 props：`modelValue=[]`、`options=[]`。
- 事件：`update:modelValue`、`change(newValue)`。
- 插槽：无。
- 示例：`<DsCheckbox v-model="types" :options="[{ label: '告警', value: 'alarm' }, { label: '离线', value: 'offline' }]" />`
- 注意事项：返回值始终为数组；当前不支持单项禁用等更细粒度控制。

#### `DsComboBox`

- 用途：组合式下拉选择器，兼顾输入框视觉与选项弹出。
- 适用场景：简单下拉选择、表单枚举字段、H5 配置页选择器。
- 主要 props：`modelValue`、`options=[]`、`placeholder='请选择'`。
- 事件：`update:modelValue`、`change(option)`。
- 插槽：无。
- 示例：`<DsComboBox v-model="form.region" :options="regionOptions" placeholder="请选择区域" />`
- 注意事项：内部使用 `document.addEventListener` 和 `closest()` 处理点击外部关闭，更偏 H5 场景，小程序和 App 端要重点验证。

#### `DsDropdown`

- 用途：轻量下拉菜单，适合顶部过滤和小范围选项切换。
- 适用场景：列表筛选、排序切换、导航工具条中的选择项。
- 主要 props：`modelValue`、`options=[]`、`placeholder='请选择'`、`direction='down'`。
- 事件：`update:modelValue`、`change(option)`。
- 插槽：无。
- 示例：`<DsDropdown v-model="sortType" :options="sortOptions" placeholder="排序方式" />`
- 注意事项：`direction='up'` 当前实现存在样式不一致，向上展开前建议实测。

#### `DsForm`

- 用途：表单容器，提供规则校验、提交与错误收集能力。
- 适用场景：登录注册、设备参数编辑、工单提交、设置页保存。
- 主要 props：`model`、`rules={}`。
- 事件：`validate`、`submit`。
- 插槽：默认 slot。
- 暴露方法：`validate()`、`validateField()`、`submit()`、`errors`。
- 示例：`<DsForm ref="formRef" :model="formData" :rules="rules"><DsFormItem label="名称"><DsInput v-model="formData.name" /></DsFormItem></DsForm>`
- 注意事项：`min/max` 校验更适合字符串长度校验，数字字段如果需要数值规则，建议自行扩展。

#### `DsFormItem`

- 用途：表单项外壳，负责标签、必填、帮助文本和错误展示。
- 适用场景：与 `DsForm` 搭配构建统一表单布局。
- 主要 props：`label`、`required`、`error`、`help`。
- 事件：无。
- 插槽：默认 slot。
- 示例：`<DsFormItem label="手机号" required help="用于接收告警通知"><DsInput v-model="form.mobile" /></DsFormItem>`
- 注意事项：只负责结构和展示，不会自动注入 `DsForm` 的错误信息，联动需要父层自行处理。

### 布局导航

#### `DsNavBar`

- 用途：页面顶部导航栏，支持返回、标题和左右操作区。
- 适用场景：详情页、设置页、表单页、自定义头部页面。
- 主要 props：`title`、`showBack=true`、`transparent=false`、`statusBar=true`、`color='#333'`。
- 事件：`click`。
- 插槽：`left`、`right`。
- 示例：`<DsNavBar title="设备详情"><template #right><DsIcon name="more" /></template></DsNavBar>`
- 注意事项：点击返回按钮会优先 `navigateBack`，失败时回退到 `/pages/index/index`；带有页面路径假设，复用前请先确认页面结构。

#### `DsSection`

- 用途：统一分区标题与内容块结构。
- 适用场景：首页分组、详情模块、设置分段、概览区块。
- 主要 props：`title`、`type='default'|'card'|'plain'`、`showLine=true`。
- 事件：无。
- 插槽：默认 slot、`extra`。
- 示例：`<DsSection title="告警概览"><template #extra><text>今日</text></template></DsSection>`
- 注意事项：更适合“标题 + 内容”的静态分区，不承担折叠或分页等复杂容器职责。

#### `DsCard`

- 用途：通用卡片容器，统一阴影、边框、头尾布局。
- 适用场景：设备卡片、统计面板、信息摘要、操作入口。
- 主要 props：`title`、`size='default'`、`shadow=true`、`bordered=false`、`padding=''`。
- 事件：`click`。
- 插槽：默认 slot、`header-extra`、`footer`。
- 示例：`<DsCard title="设备信息"><text>当前在线 16 台</text><template #footer><DsButton text="查看详情" size="small" /></template></DsCard>`
- 注意事项：卡片整体可点击；如果内部还有按钮，建议明确处理父子点击行为，避免误触。

#### `DsList`

- 用途：列表容器，负责列表整体边框和条纹样式。
- 适用场景：设置列表、菜单列表、信息列表、详情页块状列表。
- 主要 props：`bordered=false`、`striped=false`。
- 事件：无。
- 插槽：默认 slot。
- 示例：`<DsList bordered><DsListItem title="设备编号" extra="A-1024" /></DsList>`
- 注意事项：只负责列表外层结构，点击能力和子项布局由 `DsListItem` 决定。

#### `DsListItem`

- 用途：单个列表项，支持图标、头像、标题、描述、额外信息与箭头。
- 适用场景：设置项、详情字段、菜单入口、消息列表。
- 主要 props：`title`、`desc`、`extra`、`icon`、`iconSize`、`iconColor`、`avatar`、`avatarSize`、`tag`、`arrow`、`align`、`active`。
- 事件：`click`。
- 插槽：`content`、`extra`。
- 示例：`<DsListItem title="固件版本" desc="v1.2.3" extra="可升级" arrow @click="openUpgrade" />`
- 注意事项：`clickable` 和 `active` 的实现当前不完整，作为静态或轻交互列表项使用更稳妥。

#### `DsGrid`

- 用途：宫格入口组件，支持数据驱动和局部自定义内容。
- 适用场景：首页快捷入口、功能导航、操作面板、仪表盘菜单。
- 主要 props：`items=[]`、`columns=4`、`gap=16`、`iconSize=48`、`bordered=false`。
- 事件：`click({ item, index })`。
- 插槽：默认 slot、动态 `item-${index}`。
- 示例：`<DsGrid :items="menus" :columns="4" @click="openMenu" />`
- 注意事项：`bordered` 当前样式效果不稳定；如果需要个别卡片差异化，优先使用 `item-${index}` 动态 slot。

#### `DsTabs`

- 用途：标签页容器，管理当前激活项并切换展示风格。
- 适用场景：列表筛选分栏、状态切换、内容分区。
- 主要 props：`modelValue=0`、`type='line'|'card'`、`scrollable=false`。
- 事件：`update:modelValue`、`change`。
- 插槽：默认 slot。
- 示例：`<DsTabs v-model="activeTab" type="line"><DsTabItem label="全部" :index="0" :active="activeTab===0" /></DsTabs>`
- 注意事项：它更像“状态容器”而不是完整 tabs 数据渲染器，父层通常要手动管理每个子项的激活态。

#### `DsTabItem`

- 用途：单个标签项，与 `DsTabs` 配套使用。
- 适用场景：顶部状态切换、水平分类导航、可滚动标签栏。
- 主要 props：`label`、`badge`、`index`、`active`。
- 事件：`click(index)`。
- 插槽：默认 slot。
- 示例：`<DsTabItem label="在线" :index="1" :active="activeTab===1" @click="activeTab=1" />`
- 注意事项：激活态完全依赖外部传入；单独使用意义不大，建议总是和 `DsTabs` 成对出现。

#### `DsCollapse`

- 用途：折叠面板容器，支持普通展开与手风琴模式。
- 适用场景：FAQ、参数分组、配置块收纳、详情页长内容折叠。
- 主要 props：`modelValue`、`accordion=false`。
- 事件：`update:modelValue`、`change`。
- 插槽：默认 slot。
- 示例：`<DsCollapse v-model="activePanels"><DsCollapseItem title="基础信息" index="base" /></DsCollapse>`
- 注意事项：依赖 `provide/inject`；内部对外部 `modelValue` 的后续强制同步能力有限，父层主动改值时要重点测试。

#### `DsCollapseItem`

- 用途：折叠面板子项，承载标题与可折叠内容。
- 适用场景：与 `DsCollapse` 一起组织分组内容。
- 主要 props：`title`、`index`。
- 事件：无。
- 插槽：默认 slot。
- 示例：`<DsCollapseItem title="网络设置" index="network"><text>这里放网络配置内容</text></DsCollapseItem>`
- 注意事项：必须放在 `DsCollapse` 内部使用；`index` 建议保持稳定唯一。

#### `DsDialog`

- 用途：标准对话框，支持确认、取消、关闭与自定义正文。
- 适用场景：删除确认、风险提示、操作确认、轻量表单弹层。
- 主要 props：`visible`、`title`、`message`、`confirmText='确定'`、`cancelText='取消'`、`showCancel=true`、`showClose=false`、`closeOnClickMask=true`。
- 事件：`update:visible`、`confirm`、`cancel`、`close`。
- 插槽：默认 slot。
- 示例：`<DsDialog :visible="showDialog" title="删除确认" message="确认删除该设备？" @confirm="removeDevice" />`
- 注意事项：默认 slot 会覆盖 `message` 展示；如果弹层内有复杂交互，需要自行管理关闭时机和状态清理。

#### `DsDrawer`

- 用途：侧边或上下抽屉层，适合承载中等复杂度操作区。
- 适用场景：筛选面板、侧边详情、编辑表单、附加操作集合。
- 主要 props：`visible`、`title`、`direction='right'|'left'|'top'|'bottom'`、`width='80%'`、`height='80%'`、`closeOnClickMask=true`。
- 事件：`update:visible`、`close`。
- 插槽：默认 slot、`footer`。
- 示例：`<DsDrawer :visible="showFilter" title="筛选条件"><template #footer><DsButton text="确定" block /></template></DsDrawer>`
- 注意事项：宽高受方向影响；底部主操作建议统一放到 `footer`，避免正文区滚动时丢失。

### 状态反馈

#### `DsEmpty`

- 用途：统一空状态展示，可附带补救动作。
- 适用场景：空列表、无搜索结果、无网络、无权限占位。
- 主要 props：`text='暂无数据'`、`subText=''`、`icon=''`、`iconSize=80`、`iconColor='#ccc'`、`type='default'|'search'|'error'|'network'`。
- 事件：`click`。
- 插槽：`action`。
- 示例：`<DsEmpty type="search" text="未找到相关设备"><template #action><DsButton text="重试" /></template></DsEmpty>`
- 注意事项：更适合单块空态；整页空态建议配合页面级布局一起使用。

#### `DsLoading`

- 用途：统一加载态展示，提供多种视觉风格。
- 适用场景：列表初次加载、模块局部刷新、页面等待态。
- 主要 props：`text='加载中...'`、`type='ring'|'spinner'|'dots'`。
- 事件：无。
- 插槽：无。
- 示例：`<DsLoading text="设备数据加载中..." type="ring" />`
- 注意事项：纯展示组件，不负责遮罩、阻塞交互或超时处理。

#### `DsNoticeBar`

- 用途：横向通知提示条，用于重要但不阻断的消息展示。
- 适用场景：公告、告警提示、维护通知、版本提醒。
- 主要 props：`text`、`type='default'|'success'|'warning'|'danger'|'info'`、`icon`、`scrollable=true`、`closable=false`。
- 事件：`click`、`close`。
- 插槽：无。
- 示例：`<DsNoticeBar text="今日 23:00 将进行系统维护" type="warning" closable />`
- 注意事项：`scrollable` 是横向滚动，不是真正跑马灯；关闭按钮不会触发根节点点击。

#### `DsProgress`

- 用途：展示当前进度、达成率或配额使用情况。
- 适用场景：物模型进度、上传进度、容量占比、执行进展。
- 主要 props：`value=0`、`max=100`、`label`、`type='primary'|'success'|'warning'|'danger'`、`showText=true`、`showPercent=false`、`strokeWidth=12`、`color=''`。
- 事件：无。
- 插槽：无。
- 示例：`<DsProgress :value="68" :max="100" label="升级进度" show-percent />`
- 注意事项：纯展示组件；如果进度值可能越界，建议外部先归一化处理。

#### `DsResult`

- 用途：结果页组件，用于成功、失败、警告、空态等结论性反馈。
- 适用场景：提交成功页、失败兜底页、权限不足页、空业务结果页。
- 主要 props：`type='success'|'fail'|'warning'|'info'|'empty'`、`title`、`message`、`icon`、`actions=[]`、`showHome=false`。
- 事件：`click`。
- 插槽：无。
- 示例：`<DsResult type="success" title="提交成功" message="工单已进入处理队列" :actions="resultActions" />`
- 注意事项：`actions` 需要自行传入 `onClick`；`showHome` 会直接跳转 `/pages/index/index`，带有页面路径假设。

#### `DsFab`

- 用途：悬浮主按钮，突出全局高频操作。
- 适用场景：新建、扫码、快速发布、返回主操作入口。
- 主要 props：`icon='plus'`、`text=''`、`position='right-bottom'|'left-bottom'|'right-top'|'left-top'`、`mini=false`、`show=true`。
- 事件：`click`。
- 插槽：无。
- 示例：`<DsFab icon="plus" position="right-bottom" @click="createDevice" />`
- 注意事项：虽然支持 `text`，但当前样式默认并不明显展示文字，更适合纯图标主操作。

#### `DsBacktop`

- 用途：返回顶部按钮，监听页面滚动后显示。
- 适用场景：长列表、长详情页、文档页、滚动容器较稳定的 H5 页面。
- 主要 props：`threshold=400`、`duration=300`。
- 事件：无。
- 插槽：无。
- 示例：`<DsBacktop :threshold="500" :duration="300" />`
- 注意事项：依赖 DOM 查询和定时轮询，更偏 H5 场景；复杂滚动容器中表现要单独验证。

#### `DsNoMore`

- 用途：列表底部“没有更多了”提示。
- 适用场景：分页加载结束、历史记录到底、消息流终止提示。
- 主要 props：`text='没有更多了'`。
- 事件：无。
- 插槽：无。
- 示例：`<DsNoMore text="没有更多设备记录了" />`
- 注意事项：仅负责文案提示，不包含分页状态机或加载逻辑。

### 数据展示

当前项目已不再保留独立的业务名片组件，推荐直接使用 `DsCard + DsAvatar + DsTag + DsButton + DsDescription` 组合搭建个人卡、设备卡、联系人卡。这样更容易按页面场景调整布局，也能减少单独封装导致的 props 偏差。

#### `DsDescription`

- 用途：描述列表容器，用于组织多组只读字段。
- 适用场景：设备详情、工单详情、用户档案、配置预览。
- 主要 props：`title`。
- 事件：无。
- 插槽：默认 slot。
- 示例：`<DsDescription title="基础信息"><DsDescriptionItem label="设备编号" value="A-1024" /></DsDescription>`
- 注意事项：定位为静态只读信息展示，不负责编辑、校验或交互状态。

#### `DsDescriptionItem`

- 用途：描述列表单项，展示字段名与字段值。
- 适用场景：详情页键值对、摘要信息块、配置确认页。
- 主要 props：`label`、`value`。
- 事件：无。
- 插槽：默认 slot。
- 示例：`<DsDescriptionItem label="安装地址" value="上海市浦东新区张江园区" />`
- 注意事项：默认 slot 会覆盖 `value`；更适合少量结构化字段，不适合复杂富文本。

#### `DsStatistic`

- 用途：统计数值展示，支持前后缀、图标和精度格式化。
- 适用场景：仪表盘指标、首页概览、经营统计、设备数据摘要。
- 主要 props：`value=0`、`label`、`icon`、`iconColor='#5BA3E8'`、`valueColor='#333'`、`prefix`、`suffix`、`precision=0`。
- 事件：无。
- 插槽：无。
- 示例：`<DsStatistic label="今日告警" :value="12" suffix="次" icon="warning" />`
- 注意事项：内部会执行 `Number(value).toFixed(precision)`；非数字文本不适合直接传入。

#### `DsSwiper`

- 用途：轮播容器，支持数据驱动和自定义内容两种模式。
- 适用场景：Banner、活动卡片、设备图片轮播、概览页重点信息切换。
- 主要 props：`items`、`current=0`、`indicatorDots=false`、`autoplay=false`、`interval=3000`、`duration=500`、`circular=false`、`pagination=false`、`indicatorColor`、`indicatorActiveColor`。
- 事件：`update:current`、`change`、`click`。
- 插槽：默认 slot。
- 示例：`<DsSwiper :items="banners" autoplay pagination @click="openBanner" />`
- 注意事项：自定义 slot 与 `items` 混用时，需要验证分页点数量、点击回调和当前索引同步是否符合预期。

#### `DsSwiperItem`

- 用途：轮播子项容器，承载单页轮播内容。
- 适用场景：与 `DsSwiper` 搭配渲染自定义 Banner 或任意卡片内容。
- 主要 props：无。
- 事件：`click`。
- 插槽：默认 slot。
- 示例：`<DsSwiperItem @click="openDetail"><image src="/static/banner.png" mode="aspectFill" /></DsSwiperItem>`
- 注意事项：通常只在 `DsSwiper` 内部使用；点击事件适合传递页面级跳转逻辑。

#### `DsQrCode`

- 用途：根据字符串值生成二维码，可附带说明文本。
- 适用场景：设备配网、登录授权、分享链接、工单核销码。
- 主要 props：`value`、`size=200`、`text=''`、`level='L'|'M'|'Q'|'H'`。
- 事件：`click`。
- 插槽：无。
- 示例：`<DsQrCode value="device://bind/1024" text="扫码绑定设备" />`
- 注意事项：依赖 `qrcode` 和 canvas node 能力；不同端的绘制、清晰度和点击区域需要分别验证。

#### `DsTable`

- 用途：轻量只读表格，支持列配置与自定义单元格 slot。
- 适用场景：统计报表、简易对比表、明细列表、后台配置预览。
- 主要 props：`columns=[]`、`data=[]`、`emptyText='暂无数据'`。
- 事件：无。
- 插槽：动态 `cell-${key}`。
- 示例：`<DsTable :columns="columns" :data="rows"><template #cell-status="{ value }"><DsTag :text="value" /></template></DsTable>`
- 注意事项：定位是轻量展示表格，不包含排序、筛选、固定列和复选等高级能力。

## 主题变量

组件统一使用 `components/variables.scss` 中的变量，例如：

```scss
$primary-color: #5BA3E8;
$success-color: #67c23a;
$warning-color: #e6a23c;
$danger-color: #f56c6c;

$text-primary: #333333;
$text-secondary: #666666;
$text-placeholder: #999999;

$bg-primary: #ffffff;
$bg-secondary: #f5f7fa;

$radius-base: 12rpx;
$shadow-sm: 0 2rpx 8rpx rgba(0, 0, 0, 0.06);
```

## 使用注意事项

### 1. 导入方式

- 推荐直接 `import DsXxx from '@/components/DsXxx.vue'`
- 不建议假设存在稳定的 `@/components` barrel 导出

### 2. 事件与双向绑定

- 大多数输入类组件使用 `modelValue + update:modelValue`
- `DsSwitch` 例外，使用 `value + change`
- 不是所有组件都支持 `@click`，请以本文档“事件”列为准

### 3. uni-app H5 输入框注意事项

- 不要给 `input` / `textarea` 设置 `box-sizing: border-box`
- 在 uni-app H5 中，这会导致输入框无法聚焦或无法正常输入

### 4. 跨端兼容重点

- `DsComboBox` 使用了 `document.addEventListener`，小程序与 App 端要重点测试
- `DsBacktop` 依赖页面 DOM 查询，更偏 H5
- `DsQrCode` 依赖 canvas node，不同端表现需单独验证
- `DsNavBar` 和 `DsResult` 内部写死了部分默认跳转路径，复用前请确认页面结构

### 5. 当前已知实现限制

- `DsInput` 的 `type='textarea'` 目前不是真正的多行输入
- `DsDropdown` 的 `direction='up'` 样式实现存在不一致
- `DsGrid` 的 `bordered` prop 当前效果不稳定
- `DsListItem` 的 `clickable` / `active` 实现不完整
- `DsFab` 的 `text` 目前默认不显示
- `DsSwitch` 的颜色 props 当前未完全生效

如果后续组件实现有调整，请优先同步本 README，避免“示例能看但代码对不上”的情况再次出现。
